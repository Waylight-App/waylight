<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseTodosConfig, type TodoSource } from '../domain/todo';
  import type { TodoSourcesUseCases } from '../usecase/todosUseCases';

  /** Props of the to-do widget settings. */
  interface Props {
    config: Record<string, unknown>;
    /** Connected lists to choose from (#169). */
    sources?: TodoSourcesUseCases;
  }
  let { config, onsave, sources }: Props & { onsave: (config: Record<string, unknown>) => Promise<void> } =
    $props();
  // svelte-ignore state_referenced_locally
  const initial = parseTodosConfig(config);
  let showDone = $state(initial.showDone);
  let source = $state(initial.source);
  let lists = $state<TodoSource[]>([]);
  $effect(() => {
    sources
      ?.list(true)
      .then((l) => (lists = l))
      .catch(() => undefined);
  });
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ showDone, ...(source ? { source } : {}) });
  }}
>
  <label class="field"
    ><span>{t('todos.sources.list')}</span>
    <select class="input" bind:value={source}>
      <option value="">{t('todos.sources.local')}</option>
      {#each lists as l (l.id)}<option value={l.id}>{l.name}</option>{/each}
      {#if source && !lists.some((l) => l.id === source)}<option value={source}
          >{t('todos.sources.missing')}</option
        >{/if}
    </select>
  </label>
  <p class="muted small">{t('todos.sources.widget_hint')}</p>
  <label><input type="checkbox" bind:checked={showDone} /> {t('todos.show_completed_to_dos')}</label>
  <p><button class="btn btn-primary" type="submit">{t('todos.save')}</button></p>
</form>

<style>
  .field {
    display: grid;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
