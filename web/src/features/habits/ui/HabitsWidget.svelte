<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { editableDays, type Habit, parseHabitsConfig } from '../domain/habits';
  import type { HabitsUseCases } from '../usecase/habitsUseCases';
  import HabitDialog from './HabitDialog.svelte';
  import { useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';

  /** Props of the habit tracker widget (#106). */
  interface Props {
    habits: HabitsUseCases;
    config: Record<string, unknown>;
  }
  let { habits, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseHabitsConfig(config));

  let list = $state<Habit[]>([]);
  let error = $state('');
  let editing = $state<Habit | null | undefined>(undefined); // null = new

  async function load() {
    try {
      list = await habits.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }
  // The day changes at midnight; a quarter-hour poll keeps "today" right.
  $effect(() => poll(load, 15 * 60_000));

  async function run(p: Promise<Habit>) {
    try {
      const h = await p;
      list = list.map((x) => (x.id === h.id ? h : x));
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(h: Habit) {
    if (!confirm(t('common.confirm_delete', { name: h.name }))) return;
    try {
      await habits.remove(h.id);
      editing = undefined;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  // How many of today's habits are done is the card's header meta (#818).
  const setMeta = useHeaderMeta();
  $effect(() =>
    setMeta(
      list.length > 0
        ? { text: t('habits.meta_today', { done: list.filter((h) => h.doneToday).length, of: list.length }) }
        : null,
    ),
  );

  const yesterday = $derived(editableDays(new Date()).yesterday);
  const dateLabel = (iso: string) =>
    new Date(`${iso}T12:00:00`).toLocaleDateString(intlLocale(), {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
    });
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul class="habits">
  {#each list as h (h.id)}
    <ListRow nowrap showValue={cfg.showHeatmap}>
      {#snippet lead()}
        <label class="today" class:done={h.doneToday}>
          <input
            type="checkbox"
            checked={h.doneToday}
            aria-label={t('habits.done_today', { name: h.name })}
            onchange={() => run(habits.toggleToday(h.id))}
          />
          {#if h.emoji}<span class="emoji" aria-hidden="true">{h.emoji}</span>{/if}
        </label>
      {/snippet}
      <button
        class="name"
        class:done={h.doneToday}
        type="button"
        onclick={() => (editing = h)}
        title={t('habits.edit')}>{h.name}</button
      >
      {#snippet sub()}
        <span class="streak" title={t('habits.streak_title', { best: h.best })}
          >{h.current === 1
            ? t('habits.streak_one')
            : h.current > 1
              ? t('habits.streak', { n: h.current })
              : t('habits.no_streak')}</span
        >
      {/snippet}
      {#snippet value()}
        <!-- The last 7 days as squares; a wide tile shows all four weeks. -->
        <span class="week" role="group" aria-label={t('habits.last_weeks', { name: h.name })}>
          {#each h.heatmap as c (c.date)}
            {#if c.date === yesterday}
              <button
                type="button"
                class="cell"
                class:due={c.due}
                class:done={c.done}
                aria-pressed={c.done}
                aria-label={t('habits.toggle_yesterday', { name: h.name })}
                title={t('habits.toggle_yesterday', { name: h.name })}
                onclick={() => run(habits.toggleYesterday(h.id))}
              ></button>
            {:else}
              <span
                class="cell"
                class:due={c.due}
                class:done={c.done}
                title={`${dateLabel(c.date)}${c.done ? ' ✓' : ''}`}
              ></span>
            {/if}
          {/each}
        </span>
      {/snippet}
    </ListRow>
  {:else}
    <li class="muted">{t('habits.empty')}</li>
  {/each}
</ul>
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('habits.new')}</button>

{#if editing !== undefined}
  <HabitDialog
    {habits}
    item={editing}
    onsaved={() => load().catch(() => {})}
    ondelete={editing ? () => editing && remove(editing) : undefined}
    onclose={() => (editing = undefined)}
  />
{/if}

<style>
  .habits {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
  }
  .today {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    cursor: pointer;
  }
  .today input {
    width: 1.1rem;
    height: 1.1rem;
    accent-color: var(--color-accent);
  }
  .emoji {
    font-size: var(--fs-sm);
  }
  .name {
    max-width: 100%;
    border: none;
    background: none;
    padding: 0;
    text-align: left;
    font: inherit;
    color: inherit;
    cursor: pointer;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .name.done {
    color: var(--color-text-subtle);
  }
  .streak {
    white-space: nowrap;
  }
  .week {
    display: grid;
    grid-template-columns: repeat(7, 12px);
    gap: var(--space-1);
  }
  /* Four weeks of history; only the last one shows unless the tile is wide. */
  .cell:nth-child(-n + 21) {
    display: none;
  }
  @container widget (min-width: 560px) {
    .week {
      grid-template-columns: repeat(14, 12px);
    }
    .cell:nth-child(-n + 21) {
      display: block;
    }
  }
  .cell {
    width: 12px;
    height: 12px;
    border-radius: var(--radius-xs);
    border: none;
    padding: 0;
    background: var(--color-surface-2);
  }
  .cell.due:not(.done) {
    box-shadow: inset 0 0 0 1px var(--color-border-strong);
  }
  .cell.done {
    background: var(--color-accent);
  }
  button.cell {
    cursor: pointer;
    outline: 1px dashed var(--color-accent-text);
    outline-offset: 1px;
  }
</style>
