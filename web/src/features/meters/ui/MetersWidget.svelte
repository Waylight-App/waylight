<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import BarChart from '../../../shared/ui/BarChart.svelte';
  import {
    type Meter,
    type MeterDetail,
    type ReadingInput,
    today,
    validateReading,
    yearChoices,
  } from '../domain/meters';
  import type { MetersUseCases } from '../usecase/metersUseCases';
  import MeterDialog from './MeterDialog.svelte';

  /** Props of the meter readings widget (#689). */
  interface Props {
    meters: MetersUseCases;
  }
  let { meters }: Props = $props();
  const poll = usePolling();

  const currentYear = new Date().getFullYear();
  const SHOWN_READINGS = 5;

  let items = $state<Meter[]>([]);
  let selectedId = $state('');
  let year = $state(currentYear);
  let detail = $state<MeterDetail | null>(null);
  let error = $state('');
  let editing = $state<Meter | null | undefined>(undefined); // null = new
  let entry = $state<ReadingInput>({ day: today(), value: null });
  let busy = $state(false);

  const selected = $derived(items.find((m) => m.id === selectedId) ?? items[0]);
  const number = (v: number) => v.toLocaleString(intlLocale(), { maximumFractionDigits: 2 });
  const withUnit = (v: number) => `${number(v)} ${selected?.unit ?? ''}`.trim();
  const date = (day: string) =>
    new Date(`${day}T12:00:00`).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  const monthNames = Array.from({ length: 12 }, (_, i) =>
    new Date(2000, i, 1).toLocaleDateString(intlLocale(), { month: 'short' }),
  );

  async function loadDetail() {
    if (!selected) {
      detail = null;
      return;
    }
    detail = await meters.detail(selected.id, year);
  }

  async function load() {
    try {
      items = await meters.list();
      await loadDetail();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => poll(load, 30 * 60_000));

  async function choose(id: string, y: number) {
    selectedId = id;
    year = y;
    try {
      await loadDetail();
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function addReading(e: SubmitEvent) {
    e.preventDefault();
    if (!selected) return;
    const invalid = validateReading(entry, detail?.readings ?? []);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await meters.addReading(selected.id, entry);
      entry = { day: today(), value: null };
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function removeReading(day: string) {
    if (!selected) return;
    try {
      await meters.removeReading(selected.id, day);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function removeMeter() {
    const m = selected;
    if (!m || !confirm(t('common.confirm_delete', { name: m.name }))) return;
    try {
      await meters.remove(m.id);
      selectedId = '';
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function saved(m: Meter) {
    selectedId = m.id;
    await load().catch(() => {});
  }

  const labels = $derived(
    (detail?.consumption.months ?? []).map((v, i) =>
      v === null
        ? t('meters.month_unknown', { month: monthNames[i] ?? '' })
        : t('meters.month', { month: monthNames[i] ?? '', value: withUnit(v) }),
    ),
  );
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if selected}
  <div class="top">
    <label class="select">
      <span class="sr-only">{t('meters.select')}</span>
      <select class="input" value={selected.id} onchange={(e) => choose(e.currentTarget.value, year)}>
        {#each items as m (m.id)}<option value={m.id}>{m.name}</option>{/each}
      </select>
    </label>
    {#if selected.mine}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('meters.edit_named', { name: selected.name })}
        onclick={() => (editing = selected)}>✎</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('meters.remove', { name: selected.name })}
        onclick={() => removeMeter()}>✕</button
      >
    {/if}
  </div>

  <form class="entry" onsubmit={addReading}>
    <label class="field">
      <span>{t('meters.day')}</span>
      <input class="input" type="date" bind:value={entry.day} max={today()} required />
    </label>
    <label class="field">
      <span>{t('meters.value')} ({selected.unit})</span>
      <input
        class="input"
        type="number"
        inputmode="decimal"
        step="any"
        min="0"
        bind:value={entry.value}
        required
      />
    </label>
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('meters.add_reading')}</button>
  </form>

  {#if detail}
    <div class="year">
      <label>
        <span class="sr-only">{t('meters.year')}</span>
        <select
          class="input"
          value={year}
          onchange={(e) => choose(selected.id, Number(e.currentTarget.value))}
        >
          {#each yearChoices(detail.years, currentYear) as y (y)}<option value={y}>{y}</option>{/each}
        </select>
      </label>
      <span class="total">
        {#if detail.consumption.total === null}
          <span class="muted">{t('meters.year_unknown', { year })}</span>
        {:else}
          {t(detail.consumption.complete ? 'meters.year_total' : 'meters.year_total_partial', {
            year,
            value: withUnit(detail.consumption.total),
          })}
        {/if}
      </span>
    </div>
    <BarChart
      values={detail.consumption.months}
      {labels}
      ticks={monthNames.map((m) => m.slice(0, 1))}
      label={t('meters.chart', { year })}
    />

    {#if detail.readings.length === 0}
      <p class="muted">{t('meters.no_readings')}</p>
    {:else}
      <ul class="readings">
        {#each detail.readings.slice(0, SHOWN_READINGS) as r (r.day)}
          <ListRow>
            {t('meters.reading_line', { date: date(r.day), value: withUnit(r.value) })}
            {#snippet trail()}
              {#if selected.mine}
                <button
                  class="btn btn-icon"
                  type="button"
                  aria-label={t('meters.remove_reading', { date: date(r.day) })}
                  onclick={() => removeReading(r.day)}>✕</button
                >
              {/if}
            {/snippet}
          </ListRow>
        {/each}
      </ul>
    {/if}
  {/if}
{:else}
  <p class="muted">{t('meters.empty')}</p>
{/if}
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('meters.new')}</button>

{#if editing !== undefined}
  <MeterDialog {meters} meter={editing} onsaved={saved} onclose={() => (editing = undefined)} />
{/if}

<style>
  .top,
  .year {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .select {
    flex: 1;
    min-width: 0;
  }
  .entry {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .entry .field {
    flex: 1 1 8rem;
  }
  .total {
    flex: 1;
    min-width: 0;
    font-weight: 600;
  }
  .readings {
    list-style: none;
    margin: var(--space-2) 0;
    padding: 0;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
</style>
