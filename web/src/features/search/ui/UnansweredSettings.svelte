<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { UnansweredGroup } from '../domain/unanswered';
  import type { UnansweredQueries } from '../usecase/unansweredQueries';

  /**
   * Settings of the list of queries without an answer (#481): off by
   * default, kept only in this browser, deleted when switched off.
   */
  interface Props {
    queries: UnansweredQueries;
  }
  let { queries }: Props = $props();

  // svelte-ignore state_referenced_locally
  let enabled = $state(queries.enabled);
  // svelte-ignore state_referenced_locally
  let groups = $state<UnansweredGroup[]>(queries.groups());

  function toggle() {
    queries.setEnabled(enabled);
    groups = queries.groups();
  }

  function clear() {
    queries.clear();
    groups = [];
  }
</script>

<h3>{t('search.unanswered_heading')}</h3>
<label class="check">
  <input type="checkbox" bind:checked={enabled} onchange={toggle} />
  {t('search.unanswered_switch')}
</label>
<p class="muted">{t('search.unanswered_hint')}</p>
{#if enabled}
  {#if groups.length === 0}
    <p class="muted">{t('search.unanswered_empty')}</p>
  {:else}
    <ul class="groups">
      {#each groups as g (g.word)}
        <li>
          <strong>{g.word}</strong>
          <span class="muted">({g.count})</span>
          <span class="queries"
            >{g.queries.map((q) => (q.count > 1 ? `${q.query} ×${q.count}` : q.query)).join(' · ')}</span
          >
        </li>
      {/each}
    </ul>
  {/if}
  <p>
    <button class="btn" type="button" onclick={clear} disabled={groups.length === 0}
      >{t('search.unanswered_clear')}</button
    >
  </p>
{/if}

<style>
  .check {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  p {
    margin: var(--space-2) 0;
  }
  .groups {
    margin: var(--space-2) 0;
    padding-left: var(--space-3);
  }
  .groups li {
    margin: var(--space-1, var(--space-1)) 0;
  }
  .queries {
    display: block;
    color: var(--color-text-muted);
    font-size: var(--fs-sm);
    overflow-wrap: anywhere;
  }
</style>
