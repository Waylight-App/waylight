<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { t } from '../../../shared/i18n';
  import type { Tab } from '../domain/dashboard';
  import { BOTTOM_NAV_QUERY } from '../domain/mobile';
  import { tabIcon } from '../domain/tabIcon';

  /** Props of the tab bar. */
  interface Props {
    tabs: Tab[];
    activeId: string;
    editing: boolean;
    onselect: (id: string) => void;
    onadd: () => void;
    onrename: (tab: Tab) => void;
    ondelete: (tab: Tab) => void;
    /** Moves a tab by `delta` positions (-1 = left, +1 = right). */
    onmove: (tab: Tab, delta: number) => void;
    /** Opens the background dialog of a tab. */
    onbackground: (tab: Tab) => void;
    /** Opens the default-tab rules of a tab (#71). */
    onrules: (tab: Tab) => void;
    /** Saves a tab as template. */
    ontemplate: (tab: Tab) => void;
    /** Opens the widget picker from the bottom navigation (#847); omitted where adding is not allowed. */
    onwidgets?: () => void;
    /** Whether the tabs show above 640 px; a lone tab only shows in the bottom navigation, beside "Widgets". */
    wide?: boolean;
  }
  let {
    tabs,
    activeId,
    editing,
    onselect,
    onadd,
    onrename,
    ondelete,
    onmove,
    onbackground,
    onrules,
    ontemplate,
    onwidgets,
    wide = true,
  }: Props = $props();

  // One actions menu per tab in edit mode instead of seven icon buttons
  // (#274). It is positioned fixed, because the tab bar scrolls sideways.
  type Item = { label: string; run: () => void; disabled?: boolean; danger?: boolean };
  let menu = $state<{
    tab: Tab;
    items: Item[];
    x: number;
    y: number;
    up: boolean;
    opener: HTMLElement;
  } | null>(null);
  let menuEl = $state<HTMLElement>();

  // Below 640 px the tabs are a bottom navigation (#771). The CSS does the
  // layout; this only tells the page how much room the bar takes and which
  // way the actions menu opens.
  const bottomQuery = typeof matchMedia === 'function' ? matchMedia(BOTTOM_NAV_QUERY) : null;
  let bottomNav = $state(bottomQuery?.matches ?? false);
  $effect(() => {
    if (!bottomQuery) return;
    const update = () => (bottomNav = bottomQuery.matches);
    bottomQuery.addEventListener('change', update);
    return () => bottomQuery.removeEventListener('change', update);
  });
  $effect(() => {
    if (!bottomNav) return;
    const root = document.documentElement;
    root.style.setProperty(
      '--bottom-nav-h',
      'calc(var(--space-12) + var(--space-2) + env(safe-area-inset-bottom, 0px))',
    );
    return () => root.style.removeProperty('--bottom-nav-h');
  });

  function openMenu(tab: Tab, i: number, opener: HTMLElement) {
    if (menu?.tab.id === tab.id) return closeMenu();
    const r = opener.getBoundingClientRect();
    const items: Item[] = [
      { label: t('dashboard.rename_tab'), run: () => onrename(tab) },
      { label: t('dashboard.move_tab_left'), run: () => onmove(tab, -1), disabled: i === 0 },
      { label: t('dashboard.move_tab_right'), run: () => onmove(tab, 1), disabled: i === tabs.length - 1 },
      { label: t('dashboard.tab_background'), run: () => onbackground(tab) },
      { label: t('dashboard.tab_rules'), run: () => onrules(tab) },
      { label: t('dashboard.save_tab_as_template'), run: () => ontemplate(tab) },
      {
        label: t('dashboard.delete_tab'),
        run: () => ondelete(tab),
        disabled: tabs.length === 1,
        danger: true,
      },
    ];
    // The bottom navigation sits at the screen's edge, so its menu opens upwards.
    const up = bottomNav;
    menu = {
      tab,
      items,
      x: Math.min(r.left, window.innerWidth - 260),
      y: up ? window.innerHeight - r.top + 4 : r.bottom + 4,
      up,
      opener,
    };
    queueMicrotask(() => menuEl?.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus());
  }

  function closeMenu(focusOpener = true) {
    const opener = menu?.opener;
    menu = null;
    if (focusOpener) opener?.focus();
  }

  function choose(item: Item) {
    closeMenu(false);
    item.run();
  }

  function menuKey(e: KeyboardEvent) {
    const buttons = [...(menuEl?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)') ?? [])];
    const i = buttons.indexOf(document.activeElement as HTMLButtonElement);
    const next =
      e.key === 'ArrowDown'
        ? (i + 1) % buttons.length
        : e.key === 'ArrowUp'
          ? (i - 1 + buttons.length) % buttons.length
          : e.key === 'Home'
            ? 0
            : e.key === 'End'
              ? buttons.length - 1
              : null;
    if (next !== null) {
      e.preventDefault();
      buttons[next]?.focus();
    } else if (e.key === 'Escape' || e.key === 'Tab') {
      e.preventDefault();
      e.stopPropagation();
      closeMenu();
    }
  }

  $effect(() => {
    if (!menu) return;
    const outside = (e: PointerEvent) => {
      if (!menuEl?.contains(e.target as Node) && e.target !== menu?.opener) closeMenu(false);
    };
    const close = () => closeMenu(false);
    document.addEventListener('pointerdown', outside, true);
    window.addEventListener('resize', close);
    return () => {
      document.removeEventListener('pointerdown', outside, true);
      window.removeEventListener('resize', close);
    };
  });
  $effect(() => {
    if (!editing) menu = null;
  });
</script>

<nav class="tabs" class:narrow-only={!wide} aria-label={t('dashboard.dashboard_tabs')} data-tour="tabs">
  {#each tabs as tab, i (tab.id)}
    <div class="tab" class:active={tab.id === activeId}>
      <button
        type="button"
        class="name"
        id="tab-name-{tab.id}"
        data-tour="tab-{tab.name}"
        aria-current={tab.id === activeId ? 'page' : undefined}
        title={editing && tab.mine
          ? t('dashboard.rename_hint')
          : !tab.mine
            ? t('dashboard.team_tab')
            : undefined}
        onclick={() => onselect(tab.id)}
        ondblclick={() => editing && tab.mine && onrename(tab)}
        onkeydown={(e) => {
          if (editing && tab.mine && e.key === 'F2') {
            e.preventDefault();
            onrename(tab);
          }
        }}
      >
        {#if !tab.mine}<span class="team" aria-hidden="true"><Icon name="users" size={16} /></span>{/if}<span
          class="ico"
          aria-hidden="true"><Icon name={tab.mine ? tabIcon(tab.name) : 'users'} size={16} /></span
        ><span class="label">{tab.name}</span>
      </button>
      <!-- Team tabs (#295) are maintained by their owner only. -->
      {#if editing && tab.mine}
        <button
          class="btn btn-icon more"
          type="button"
          aria-label={t('dashboard.tab_actions_button')}
          aria-describedby="tab-name-{tab.id}"
          title={t('dashboard.tab_actions', { name: tab.name })}
          aria-haspopup="menu"
          aria-expanded={menu?.tab.id === tab.id}
          onclick={(e) => openMenu(tab, i, e.currentTarget)}>⋯</button
        >
      {/if}
    </div>
  {/each}
  {#if editing}
    <button class="btn btn-icon" type="button" onclick={onadd}>{t('dashboard.tab')}</button>
  {:else if onwidgets}
    <button type="button" class="name widgets" title={t('dashboard.widgets_nav_hint')} onclick={onwidgets}
      ><span class="ico" aria-hidden="true"><Icon name="widgets" size={16} /></span><span class="label"
        >{t('dashboard.widgets_nav')}</span
      ></button
    >
  {/if}
</nav>
{#if menu}
  <div
    class="menu"
    role="menu"
    tabindex="-1"
    aria-label={t('dashboard.tab_actions', { name: menu.tab.name })}
    style:left="{menu.x}px"
    style:top={menu.up ? undefined : `${menu.y}px`}
    style:bottom={menu.up ? `${menu.y}px` : undefined}
    bind:this={menuEl}
    onkeydown={menuKey}
  >
    {#each menu.items as item (item.label)}
      <button
        type="button"
        role="menuitem"
        class:danger={item.danger}
        disabled={item.disabled}
        onclick={() => choose(item)}>{item.label}</button
      >
    {/each}
  </div>
{/if}

<style>
  .more {
    margin-left: -6px;
    opacity: 0.75;
  }
  .more:hover,
  .more[aria-expanded='true'] {
    opacity: 1;
  }
  .menu {
    position: fixed;
    z-index: 50;
    min-width: 220px;
    display: flex;
    flex-direction: column;
    padding: var(--space-1);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-pop);
  }
  .menu button {
    text-align: left;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-sm);
    cursor: pointer;
    min-height: 32px;
  }
  .menu button:hover:not(:disabled),
  .menu button:focus-visible {
    background: var(--color-surface-2);
    outline: none;
  }
  .menu button:disabled {
    opacity: 0.45;
    cursor: default;
  }
  .menu button.danger {
    color: var(--color-danger);
  }
  /* Tab pills at the end of the title row (#791); overflow scrolls with faded edges. */
  .tabs {
    display: flex;
    flex-wrap: nowrap;
    justify-content: safe flex-end;
    gap: var(--space-1);
    align-items: center;
    overflow-x: auto;
    scrollbar-width: none;
    max-width: 100%;
    padding: var(--space-1) var(--space-3);
    mask-image: linear-gradient(to right, transparent 0, #000 12px, #000 calc(100% - 12px), transparent 100%);
  }
  .tabs::-webkit-scrollbar {
    display: none;
  }
  .tab {
    display: flex;
    align-items: center;
    flex: none;
    border-radius: 999px;
  }
  .name {
    border: none;
    background: none;
    min-height: 32px;
    padding: 0 var(--space-4);
    border-radius: 999px;
    cursor: pointer;
    color: var(--color-text-muted);
    font: inherit;
    font-weight: 500;
    white-space: nowrap;
  }
  .name:hover {
    color: var(--color-text);
  }
  /* The selected tab is a step lighter with a 1 px inner ring, not an accent fill. */
  .active .name {
    background: var(--color-surface-2);
    color: var(--color-text);
    box-shadow: inset 0 0 0 1px var(--color-border-strong);
  }
  .name:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .ico {
    display: none;
  }
  /* The widgets entry and a lone tab belong to the bottom navigation only. */
  .widgets,
  .narrow-only {
    display: none;
  }
  /* Phones (#771): a fixed bottom navigation with icon over label, 56 px
     high plus the safe area, and at least 44 px wide targets. */
  @media (max-width: 639px) {
    .tabs {
      position: fixed;
      inset: auto 0 0;
      z-index: 40;
      justify-content: safe space-around;
      gap: 0;
      padding: var(--space-2) var(--space-2) calc(var(--space-4) + env(safe-area-inset-bottom, 0px));
      background: var(--color-bg);
      border-top: 1px solid var(--color-border);
      mask-image: none;
    }
    .tab {
      flex: 1 0 auto;
      min-width: 5rem;
      border-radius: 0;
      justify-content: center;
    }
    .name {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: var(--space-1);
      flex: 1;
      min-height: var(--space-12);
      min-width: 44px;
      padding: 0 var(--space-2);
      border-radius: 0;
      color: var(--color-text-subtle);
      font-size: var(--fs-xs);
      line-height: var(--lh-xs);
    }
    .ico {
      display: inline-flex;
    }
    .narrow-only {
      display: flex;
    }
    /* Always the last entry: it stays at the right edge while many tabs scroll beneath it. */
    .widgets {
      display: flex;
      position: sticky;
      right: 0;
      flex: 1 0 auto;
      min-width: 5rem;
      background: var(--color-bg);
      /* Covers the bar's end padding too, so no scrolled tab peeks out behind it. */
      box-shadow: var(--space-2) 0 0 var(--color-bg);
    }
    .ico :global(svg) {
      width: 20px;
      height: 20px;
    }
    .team {
      display: none;
    }
    .active .name {
      background: none;
      box-shadow: none;
      color: var(--color-accent-text);
    }
    .tabs > .btn-icon,
    .more {
      min-width: 44px;
      min-height: 44px;
      margin-left: 0;
    }
  }
</style>
