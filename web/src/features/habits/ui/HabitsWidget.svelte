<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { editableDays, type Habit, heatWeeks, parseHabitsConfig } from '../domain/habits';
  import type { HabitsUseCases } from '../usecase/habitsUseCases';
  import HabitDialog from './HabitDialog.svelte';

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
    <ListRow nowrap>
      {#snippet lead()}
        <label class="today" class:done={h.doneToday}>
          <input
            type="checkbox"
            checked={h.doneToday}
            aria-label={t('habits.done_today', { name: h.name })}
            onchange={() => run(habits.toggleToday(h.id))}
          />
          <span class="emoji" aria-hidden="true"
            >{#if h.emoji}{h.emoji}{:else}<Icon name="done" size={16} />{/if}</span
          >
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
        {#if cfg.showHeatmap}
          <div class="heat" role="group" aria-label={t('habits.last_weeks', { name: h.name })}>
            {#each heatWeeks(h.heatmap) as week, w (w)}
              {#each week as c (c.date)}
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
            {/each}
          </div>
        {/if}
      {/snippet}
      {#snippet value()}
        <span class="streak" title={t('habits.streak_title', { best: h.best })}>
          {#if h.current > 0}<Icon name="flame" size={16} /> {h.current}{:else}<span class="muted">–</span
            >{/if}
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
    color: var(--color-text);
  }
  .heat {
    display: grid;
    grid-template-columns: repeat(28, 1fr);
    gap: var(--space-1);
    max-width: 22rem;
  }
  .cell {
    aspect-ratio: 1;
    border-radius: var(--radius-sm);
    border: none;
    padding: 0;
    background: color-mix(in srgb, var(--color-text-subtle) 10%, transparent);
  }
  .cell.due {
    background: color-mix(in srgb, var(--color-text-subtle) 25%, transparent);
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
