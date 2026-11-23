<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { Tab } from '../domain/dashboard';

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
  }: Props = $props();

  // One actions menu per tab in edit mode instead of seven icon buttons
  // (#274). It is positioned fixed, because the tab bar scrolls sideways.
  type Item = { label: string; run: () => void; disabled?: boolean; danger?: boolean };
  let menu = $state<{ tab: Tab; items: Item[]; x: number; y: number; opener: HTMLElement } | null>(null);
  let menuEl = $state<HTMLElement>();

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
    menu = { tab, items, x: Math.min(r.left, window.innerWidth - 260), y: r.bottom + 4, opener };
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

<nav class="tabs" aria-label={t('dashboard.dashboard_tabs')} data-tour="tabs">
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
        {#if !tab.mine}<span class="team" aria-hidden="true">👥</span>{/if}{tab.name}
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
  {/if}
</nav>
{#if menu}
  <div
    class="menu"
    role="menu"
    tabindex="-1"
    aria-label={t('dashboard.tab_actions', { name: menu.tab.name })}
    style:left="{menu.x}px"
    style:top="{menu.y}px"
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
    padding: 8px var(--space-3);
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
  /* Centred tab pills below the search (#189); overflow scrolls with faded edges. */
  .tabs {
    display: flex;
    flex-wrap: nowrap;
    justify-content: safe center;
    gap: var(--space-1);
    align-items: center;
    overflow-x: auto;
    scrollbar-width: none;
    max-width: 100%;
    padding: 2px 12px;
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
    padding: var(--space-1) var(--space-4);
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
  .active .name {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    font-weight: 600;
  }
  .name:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
</style>
