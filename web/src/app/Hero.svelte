<script lang="ts">
  import type { CommandHandlers } from './appCommands';
  import { dayPart, greetingKey, momentKey } from '../features/appearance/domain/greeting';
  import { hintExamples } from '../features/search/domain/help';
  import { HINT_INTERVAL_MS, MAX_EXAMPLE_HINTS, nextHint } from '../features/search/domain/hints';
  import type { WeatherConfig } from '../features/weather/domain/weather';
  import { intlLocale, t } from '../shared/i18n';
  import Icon from '../shared/ui/icons/Icon.svelte';
  import type { Container } from './container';

  /** Props of the hero below the top bar (#189): greeting, search field and tabs. */
  interface Props {
    /** Name in the greeting; may be empty. */
    name: string;
    /** Show a line by time of day under the greeting (#512). */
    moment?: boolean;
    /** Id of the element the dashboard renders its tab pills into. */
    tabsId: string;
    container: Container;
    /** Enabled modules; null while loading (the search waits for them). */
    modules: Set<string> | null;
    /** Location of the first weather widget for weather answers (#156). */
    weatherLocation?: WeatherConfig | null;
    /** First widget config of a kind, e.g. the transit stop (#161). */
    widgetConfig?: (kind: string) => Record<string, unknown> | null;
    /** Handlers of the app commands in the search (#476). */
    commandHandlers?: CommandHandlers;
  }
  let {
    name,
    moment = false,
    tabsId,
    container,
    modules,
    weatherLocation = null,
    widgetConfig,
    commandHandlers,
  }: Props = $props();

  let now = $state(new Date());
  let hint = $state(0);
  let focused = $state(false);
  const reducedMotion =
    typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;

  $effect(() => {
    const id = setInterval(() => (now = new Date()), 60_000);
    return () => clearInterval(id);
  });

  // The placeholder rotates every few seconds; paused on focus, off for reduced motion.
  $effect(() => {
    if (reducedMotion) return;
    const id = setInterval(
      () => (hint = nextHint(hint, hints.length, { focused, reducedMotion })),
      HINT_INTERVAL_MS,
    );
    return () => clearInterval(id);
  });

  // Hints after the first come from the help data (#479), so they only name
  // what is enabled. Loaded lazily: the help pulls in every search source,
  // which must not grow the entry chunk.
  let examples = $state<string[]>([]);
  $effect(() => {
    const enabled = modules;
    if (!enabled) return;
    const auth = container.auth;
    void import('./searchHelp').then(({ buildHelp }) => {
      examples = hintExamples(
        buildHelp({
          modules: enabled,
          viewer: { guest: auth.guest, kiosk: auth.kiosk, admin: auth.admin, owner: auth.owner },
          prefixes: { todo: ['t'], note: ['n'], bookmark: ['b'] },
        }),
      ).slice(0, MAX_EXAMPLE_HINTS);
    });
  });
  const hints = $derived([
    t('search.hint.learn'),
    ...examples.map((example) => t('search.hint.try', { example })),
    t('search.hint.help'),
  ]);

  const hello = $derived(t(greetingKey(dayPart(now), name !== ''), { name }));
  const momentLine = $derived(
    moment ? t(momentKey(now), { weekday: now.toLocaleDateString(intlLocale(), { weekday: 'long' }) }) : '',
  );

  // The field is the search (#222): it keeps the text and the focus, and the
  // results open in a panel directly below it (SearchPanel.svelte).
  const listboxId = 'hero-search-results';
  let query = $state('');
  let open = $state(false);
  let activeId = $state<string | undefined>();
  let expanded = $state(false);
  let input = $state<HTMLInputElement>();
  let box = $state<HTMLElement>();
  let panel = $state<{ keydown: (e: KeyboardEvent) => boolean }>();
  /** How long an early Enter waits for the panel before it searches the web. */
  const PANEL_WAIT_MS = 3000;
  let pendingEnter = $state<KeyboardEventInit | null>(null);
  let pendingTimer: ReturnType<typeof setTimeout> | undefined;

  /** Focuses the field and opens the panel (Ctrl+K and /). */
  export function focus() {
    input?.focus();
    open = true;
  }

  function close() {
    open = false;
    expanded = false;
    activeId = undefined;
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      // First Esc closes the panel and keeps the text; the second clears and leaves.
      e.preventDefault();
      if (open) close();
      else {
        query = '';
        input?.blur();
      }
      return;
    }
    if (!open && e.key !== 'Tab') open = true;
    if (panel?.keydown(e)) e.preventDefault();
    else if (e.key === 'Enter' && !panel) {
      // Enter before the panel has loaded waits for it, so "uhrzeit" still
      // gets its answer instead of going straight to the web search. Only
      // if the panel does not arrive in time is the query searched on the web.
      e.preventDefault();
      pendingEnter = { altKey: e.altKey, shiftKey: e.shiftKey, ctrlKey: e.ctrlKey, metaKey: e.metaKey };
      clearTimeout(pendingTimer);
      pendingTimer = setTimeout(() => {
        if (!pendingEnter) return;
        pendingEnter = null;
        const target = modules?.has('search') ? container.search.targetFor(query) : null;
        if (target) window.location.assign(target);
      }, PANEL_WAIT_MS);
    }
  }

  // Replays an Enter pressed while the panel was still loading.
  $effect(() => {
    if (!panel || !pendingEnter) return;
    const keys = pendingEnter;
    pendingEnter = null;
    clearTimeout(pendingTimer);
    panel.keydown(new KeyboardEvent('keydown', { key: 'Enter', ...keys }));
  });

  // A click outside the field and its panel closes the panel.
  $effect(() => {
    if (!open) return;
    const away = (e: PointerEvent) => {
      if (box && !box.contains(e.target as Node)) close();
    };
    document.addEventListener('pointerdown', away);
    return () => document.removeEventListener('pointerdown', away);
  });
</script>

<section class="hero">
  <h1>{hello}</h1>
  {#if momentLine}<p class="moment">{momentLine}</p>{/if}
  <div class="search" bind:this={box}>
    <label class="pill" class:active={open} data-tour="search">
      <Icon name="search" size={16} />
      <input
        bind:this={input}
        bind:value={query}
        type="search"
        role="combobox"
        aria-label={t('app.quick_search_ctrl_k_or')}
        aria-keyshortcuts="Control+K /"
        aria-expanded={expanded}
        aria-controls={listboxId}
        aria-autocomplete="list"
        aria-activedescendant={activeId}
        placeholder={hints[hint] ?? hints[0]}
        autocomplete="off"
        onfocus={() => ((focused = true), (open = true))}
        onblur={() => (focused = false)}
        oninput={() => (open = true)}
        {onkeydown}
      />
      <kbd>{open ? 'Esc' : t('app.ctrl_k')}</kbd>
    </label>
    {#if open && modules}
      <!-- Loaded on first use: the search logic is not part of the first paint (#116). -->
      {#await import('./SearchPanel.svelte') then { default: SearchPanel }}
        <SearchPanel
          bind:this={panel}
          bind:query
          bind:activeId
          bind:expanded
          {container}
          {modules}
          {listboxId}
          {weatherLocation}
          {widgetConfig}
          {commandHandlers}
          onclose={close}
          ondone={() => ((query = ''), close(), input?.blur())}
        />
      {/await}
    {/if}
  </div>
  <div id={tabsId} class="tabs"></div>
</section>

<style>
  /* Title left and tabs right on one row, the search field below (#791). The
     DOM keeps title, search, tabs, so the keyboard order is unchanged. */
  .hero {
    display: grid;
    /* The first column may shrink below its content's width: otherwise the
       tab bar's min-content width stretches it, and with it the title and
       the search, past the screen (#226). */
    /* The tab column shrinks to the room left, so many tabs scroll inside
       the bar instead of pushing the title and the page wider. */
    grid-template-columns: minmax(8rem, 1fr) minmax(0, auto);
    grid-template-areas:
      'title tabs'
      'moment moment'
      'search search';
    align-items: center;
    column-gap: var(--space-4);
    row-gap: var(--space-4);
    padding: var(--space-6) var(--space-6) 0;
    max-width: 1600px;
    margin: 0 auto;
  }
  h1 {
    grid-area: title;
    margin: 0;
    font-size: var(--fs-lg);
    font-weight: 500;
    letter-spacing: -0.01em;
    overflow-wrap: anywhere; /* a long name wraps instead of widening the page */
  }
  .search {
    grid-area: search;
    position: relative;
    width: min(32.5rem, 100%);
  }
  .pill {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    box-sizing: border-box;
    height: var(--space-12);
    padding: 0 var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    color: var(--color-text-subtle);
  }
  .pill.active {
    border-color: var(--color-accent-text);
  }
  .pill:focus-within {
    border-color: var(--color-accent-text);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-accent) 45%, transparent);
  }
  .pill input {
    flex: 1;
    min-width: 0;
    border: none;
    background: none;
    font: inherit;
    font-size: var(--fs-sm);
    color: var(--color-text);
    outline: none;
  }
  /* The example hint is the same size as typed text, only quieter. */
  .pill input::placeholder {
    color: var(--color-text-subtle);
    opacity: 1;
  }
  kbd {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    line-height: 1.375rem;
    height: 1.5rem;
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-xs);
    padding: 0 var(--space-2);
    white-space: nowrap;
  }
  .tabs {
    grid-area: tabs;
    max-width: 100%;
    min-width: 0;
  }
  .tabs:empty {
    display: none;
  }
  .moment {
    grid-area: moment;
    margin: 0;
    color: var(--color-text-subtle);
  }
  /* Phones: title above, the tabs are the bottom navigation (#771). */
  @media (max-width: 639px) {
    .hero {
      grid-template-columns: minmax(0, 1fr);
      grid-template-areas:
        'title'
        'moment'
        'search';
      row-gap: var(--space-3);
      padding: var(--space-2) var(--space-4) 0;
    }
    .pill {
      height: 2.75rem;
    }
    kbd {
      display: none;
    }
  }
</style>
