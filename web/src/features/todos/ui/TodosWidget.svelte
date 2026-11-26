<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { intlLocale, t } from '../../../shared/i18n';
  import TodoEditDialog from './TodoEditDialog.svelte';
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { usePolling } from '../../../shared/refresh';
  import {
    capabilitiesOf,
    describeReminder,
    describeRepeat,
    formatDue,
    isDueToday,
    parseTodosConfig,
    type Todo,
    type TodoSource,
  } from '../domain/todo';
  import type { TodoSourcesUseCases, TodosUseCases } from '../usecase/todosUseCases';
  import { useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';

  /** Props of the to-do widget. */
  interface Props {
    /** The to-dos of the list the widget shows (built-in or connected, #169). */
    todos: TodosUseCases;
    config: Record<string, unknown>;
    /** Connected lists, for their names and capabilities. */
    sources?: TodoSourcesUseCases;
  }
  let { todos, config, sources }: Props = $props();

  const cfg = $derived(parseTodosConfig(config));
  let connected = $state<TodoSource[]>([]);
  const caps = $derived(capabilitiesOf(cfg.source, connected));
  const poll = usePolling();
  let list = $state<Todo[]>([]);
  let text = $state('');
  let error = $state('');
  const hasDone = $derived(list.some((t) => t.done));
  let editing = $state<Todo | null>(null);
  const dayName = (d: number) =>
    new Date(2026, 10, 8 + d).toLocaleDateString(intlLocale(), { weekday: 'short' });

  // What is due goes into the card header (#809): overdue first, since it needs a look.
  const setMeta = useHeaderMeta();
  $effect(() => {
    const open = list.filter((x) => !x.done && x.due);
    const overdue = open.filter((x) => x.overdue).length;
    const today = open.filter((x) => !x.overdue && x.due && isDueToday(x.due, new Date())).length;
    setMeta(
      overdue > 0
        ? { text: t('todos.meta_overdue', { n: overdue }), tone: 'warning' }
        : today > 0
          ? { text: t('todos.meta_due_today', { n: today }) }
          : null,
    );
  });

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      list = await todos.list(cfg.showDone);
    } catch (err) {
      error = errorMessage(err);
      // Changed in the other app meanwhile: its version wins, show it (#169).
      if (err instanceof ApiError && err.status === 409) {
        error = t('todos.changed_elsewhere');
        list = await todos.list(cfg.showDone).catch(() => list);
      }
    }
  }

  $effect(() => {
    void cfg;
    if (!cfg.source) {
      void run(async () => {});
      return;
    }
    // A connected list is refreshed every minute, so changes made in the
    // other app show up (the server only downloads what changed).
    sources
      ?.list()
      .then((l) => (connected = l))
      .catch(() => undefined);
    return poll(() => run(async () => {}), 60_000);
  });

  function add(e: SubmitEvent) {
    e.preventDefault();
    if (text.trim() === '') return;
    run(() => todos.add(text)).then(() => (text = ''));
  }
</script>

<form onsubmit={add}>
  <!-- The short placeholder keeps the field on one line; the date syntax stays in the title. -->
  <span class="add"
    ><Icon name="plus" size={16} /><input
      class="input"
      bind:value={text}
      placeholder={t('todos.add_short')}
      title={t('todos.add_a_to_do_2026_12')}
      aria-label={t('todos.new_to_do')}
    /></span
  >
</form>
{#if error}<p class="error">{error}</p>{/if}
<ul class="list">
  {#each list as todo (todo.id)}
    <ListRow nowrap showValue={!!todo.due}>
      <label>
        <input type="checkbox" checked={todo.done} onchange={() => run(() => todos.toggle(todo))} />
        <span class="title" class:done={todo.done}>{todo.title}</span>
      </label>
      {#snippet value()}
        {#if todo.due}
          {@const label = formatDue(todo.due, new Date())}
          <span
            class="due"
            class:hot={!todo.overdue && isDueToday(todo.due, new Date())}
            class:late={todo.overdue}>{label}</span
          >
        {/if}
      {/snippet}
      {#snippet trail()}
        {#if todo.repeat.kind}
          {@const text = describeRepeat(todo.repeat, dayName)}
          <span class="icon" title={text} aria-label={text} role="img"><Icon name="repeat" size={16} /></span>
        {/if}
        {#if todo.reminder.at}
          {@const text = describeReminder(todo.reminder)}
          <span class="icon" title={text} aria-label={text} role="img"><Icon name="bell" size={16} /></span>
        {/if}
        <span class="tools">
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('todos.edit', { title: todo.title })}
            onclick={() => (editing = todo)}><Icon name="pencil" size={16} /></button
          >
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('todos.delete', { title: todo.title })}
            onclick={() => run(() => todos.delete(todo.id))}><Icon name="trash" size={16} /></button
          >
        </span>
      {/snippet}
    </ListRow>
  {:else}
    <li class="muted empty">{t('todos.nothing_to_do')}</li>
  {/each}
</ul>
{#if editing}
  <TodoEditDialog
    todo={editing}
    {todos}
    capabilities={caps}
    onsaved={() => run(async () => {})}
    onclose={() => (editing = null)}
  />
{/if}
{#if hasDone}
  <button class="btn btn-icon muted" type="button" onclick={() => run(() => todos.clearCompleted())}
    >{t('todos.clear_completed')}</button
  >
{/if}

<style>
  form {
    margin-bottom: var(--space-4);
  }
  .add {
    position: relative;
    display: block;
  }
  .add :global(svg) {
    position: absolute;
    inset: 50% auto auto var(--space-3);
    transform: translateY(-50%);
    color: var(--color-text-subtle);
    pointer-events: none;
  }
  .add .input {
    padding-left: var(--space-8);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .empty {
    padding: var(--space-1) 0;
  }
  label {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-width: 0;
  }
  .title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .title.done {
    text-decoration: line-through;
    color: var(--color-text-subtle);
  }
  .due {
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
    white-space: nowrap;
  }
  .due.hot {
    color: var(--color-accent-text);
    font-weight: 500;
  }
  .due.late {
    color: var(--color-danger);
    font-weight: 500;
  }
  .icon {
    font-size: var(--fs-xs);
  }
  /* Edit and delete lie over the end of the row and show on hover or focus,
     so rows keep one line with the due date at the right edge. */
  .tools {
    position: absolute;
    inset: 50% 0 auto auto;
    transform: translateY(-50%);
    display: flex;
    background: var(--color-surface);
    opacity: 0;
    pointer-events: none;
  }
  :global(.row:hover) .tools,
  .tools:focus-within {
    opacity: 1;
    pointer-events: auto;
  }
  @media (hover: none) {
    .tools {
      position: static;
      transform: none;
      opacity: 1;
      pointer-events: auto;
    }
  }
</style>
