<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { COMPACT_REFRESH_MS, compactText, type CompactState, isAlarming } from '../domain/compact';
  import type { CompactUseCases } from '../usecase/compactUseCases';

  /**
   * Props of the system summary (#593): the compact states of a tab's
   * status widgets (services, backups, drives) in one line, which can stand
   * in for the widgets themselves.
   */
  interface Props {
    compact: CompactUseCases;
    /** One entry per module; title is the widget's name. */
    sources: { module: string; title: string }[];
    /** Whether the status widgets are folded into this line. */
    collapsed: boolean;
    ontoggle: () => void;
  }
  let { compact, sources, collapsed, ontoggle }: Props = $props();

  let states = $state<Record<string, CompactState>>({});

  async function load() {
    const wanted = sources;
    const results = await Promise.allSettled(wanted.map((s) => compact.state(s.module)));
    const next: Record<string, CompactState> = {};
    results.forEach((r, i) => {
      const s = wanted[i];
      if (s && r.status === 'fulfilled') next[s.module] = r.value;
    });
    states = next;
  }

  const poll = usePolling();
  $effect(() => poll(load, COMPACT_REFRESH_MS));
</script>

<div class="summary" data-testid="system-summary">
  <ul aria-label={t('compact.summary')}>
    {#each sources as s (s.module)}
      {@const state = states[s.module]}
      {#if state}
        <li data-level={state.level}>
          {#if isAlarming(state.level)}<span class="dot" aria-hidden="true"></span>{/if}
          <span class="name">{s.title}:</span>
          {compactText(state.title)}
        </li>
      {/if}
    {/each}
  </ul>
  <button class="btn" type="button" aria-expanded={!collapsed} onclick={ontoggle}
    >{t(collapsed ? 'compact.show_all' : 'compact.collapse')}</button
  >
</div>

<style>
  .summary {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius);
    border: 1px solid var(--color-border);
    background: var(--color-surface);
  }
  ul {
    flex: 1;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-3);
    margin: 0;
    padding: 0;
    list-style: none;
    min-width: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  li + li::before {
    content: '·';
    margin-right: var(--space-2);
    color: var(--color-text-muted);
  }
  .name {
    color: var(--color-text-muted);
  }
  .dot {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
  }
  [data-level='attention'] .dot {
    background: var(--color-warning);
  }
  [data-level='problem'] .dot {
    background: var(--color-danger);
  }
</style>
