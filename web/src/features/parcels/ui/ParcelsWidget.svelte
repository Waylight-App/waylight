<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Parcel, parcelName, statusInfo, CARRIERS, type Carrier, carrierName } from '../domain/parcel';
  import type { ParcelsUseCases } from '../usecase/parcelsUseCases';

  /** Props of the parcels widget. */
  interface Props {
    parcels: ParcelsUseCases;
  }
  let { parcels }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 10 * 60_000;
  let list = $state<Parcel[]>([]);
  let loaded = $state(false);
  let adding = $state(false);
  let number = $state('');
  let label = $state('');
  let carrier = $state<Carrier>('dhl');
  let error = $state('');

  const load = () =>
    parcels
      .list()
      .then((l) => ((list = l), (loaded = true), (error = '')))
      .catch((err) => {
        error = errorMessage(err);
        throw err;
      });

  $effect(() => {
    return poll(load, POLL_MS);
  });

  async function add(e: SubmitEvent) {
    e.preventDefault();
    try {
      await parcels.add(number, label, carrier);
      number = label = '';
      carrier = 'dhl';
      adding = false;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(p: Parcel) {
    if (!confirm(t('parcels.confirm_stop', { name: parcelName(p) }))) return;
    try {
      await parcels.remove(p.id);
      list = list.filter((x) => x.id !== p.id);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const day = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'numeric' });
</script>

<div class="head">
  <button class="btn btn-icon" type="button" aria-expanded={adding} onclick={() => (adding = !adding)}
    >{adding ? '✕' : t('parcels.add_button')}</button
  >
</div>
{#if adding}
  <form class="add" onsubmit={add}>
    <input
      class="input"
      bind:value={number}
      placeholder={t('parcels.tracking_number_dhl')}
      aria-label={t('parcels.tracking_number')}
      required
    />
    <select class="input" bind:value={carrier} aria-label={t('parcels.carrier')}>
      {#each CARRIERS as c (c)}<option value={c}>{carrierName(c)}</option>{/each}
    </select>
    <input
      class="input"
      bind:value={label}
      placeholder={t('parcels.label_optional')}
      aria-label={t('parcels.label')}
      maxlength="60"
    />
    <button class="btn btn-primary" type="submit">{t('parcels.track')}</button>
  </form>
{/if}
{#if error}<p class="error small" role="alert">{error}</p>{/if}
<ul class="parcels">
  {#each list as p (p.id)}
    {@const info = statusInfo[p.status]}
    <ListRow nowrap showValue={p.trackable}>
      <span class="name" title={p.trackingNumber}
        >{#if p.discovered}<span title={t('parcels.discovered')} aria-label={t('parcels.discovered')}
            >✉
          </span>{/if}{parcelName(p)}</span
      >
      {#snippet value()}
        <span class="badge" style="color: {info.color}">{t(info.label)}</span>
      {/snippet}
      {#snippet trail()}
        {#if p.trackingUrl}
          <a
            class="btn btn-icon"
            href={p.trackingUrl}
            target="_blank"
            rel="noopener noreferrer"
            aria-label={t('parcels.track_at', { carrier: carrierName(p.carrier), name: parcelName(p) })}>↗</a
          >
        {/if}
        <button
          class="btn btn-icon rm"
          type="button"
          aria-label={t('parcels.remove', { value: parcelName(p) })}
          onclick={() => remove(p)}>✕</button
        >
      {/snippet}
      {#snippet sub()}
        {#if !p.trackable}
          {t('parcels.link_only', { carrier: carrierName(p.carrier) })}
        {:else if p.error}
          {p.error}
        {:else}
          {p.description}{p.location ? ` · ${p.location}` : ''}
          {#if p.estimatedDelivery && p.status !== 'delivered'}
            · expected {day(p.estimatedDelivery)}{/if}
        {/if}
      {/snippet}
    </ListRow>
  {:else}
    {#if loaded}<li class="muted">{t('parcels.no_parcels_tracked')}</li>{/if}
  {/each}
</ul>

<style>
  .head {
    display: flex;
    justify-content: flex-end;
  }
  .add {
    display: grid;
    gap: var(--space-1);
    margin: var(--space-1) 0 var(--space-2);
  }
  .parcels {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .name {
    font-weight: 600;
  }
  .badge {
    font-size: var(--fs-xs);
    white-space: nowrap;
  }
  .rm {
    padding: 0 var(--space-1);
  }
</style>
