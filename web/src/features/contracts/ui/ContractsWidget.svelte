<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { type Contract, formatMoney, monthlyTotal, parseConfig } from '../domain/contracts';
  import type { ContractsUseCases } from '../usecase/contractsUseCases';
  import ContractDialog from './ContractDialog.svelte';

  /** Props of the contracts widget (#687). */
  interface Props {
    contracts: ContractsUseCases;
    config: Record<string, unknown>;
  }
  let { contracts, config }: Props = $props();
  const poll = usePolling();

  const cfg = $derived(parseConfig(config));
  let items = $state<Contract[]>([]);
  let error = $state('');
  let editing = $state<Contract | null | undefined>(undefined); // null = new

  const money = (cents: number) => formatMoney(cents, cfg.currency, intlLocale());
  const date = (day: string) =>
    new Date(`${day}T12:00:00`).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });

  async function load() {
    try {
      items = await contracts.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => poll(load, 30 * 60_000));

  async function remove(c: Contract) {
    if (!confirm(t('common.confirm_delete', { name: c.name }))) return;
    try {
      await contracts.remove(c.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if items.length > 0}
  <p class="total">{t('contracts.total', { amount: money(monthlyTotal(items)) })}</p>
{/if}
<ul>
  {#each items as c (c.id)}
    <ListRow nowrap showValue style={c.open ? '' : 'opacity: 0.6'}>
      {c.name}{#if c.provider}<span class="muted">&nbsp;· {c.provider}</span>{/if}
      {#snippet sub()}
        <span class:soon={c.soon}>
          {#if c.open}
            {t('contracts.cancel_by', { date: date(c.deadline) })}
          {:else}
            {t('contracts.ended', { date: date(c.currentEnd) })}
          {/if}
        </span>
      {/snippet}
      {#snippet value()}<span class="cost">{money(c.costCents)} {t(`contracts.per.${c.period}`)}</span
        >{/snippet}
      {#snippet trail()}
        {#if c.mine}
          <span class="actions">
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('contracts.edit_named', { name: c.name })}
              onclick={() => (editing = c)}>✎</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('contracts.remove', { name: c.name })}
              onclick={() => remove(c)}>✕</button
            >
          </span>
        {/if}
      {/snippet}
    </ListRow>
  {:else}
    <li class="muted">{t('contracts.empty')}</li>
  {/each}
</ul>
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('contracts.new')}</button>

{#if editing !== undefined}
  <ContractDialog
    {contracts}
    contract={editing}
    currency={cfg.currency}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = undefined)}
  />
{/if}

<style>
  .total {
    margin: 0 0 var(--space-2);
    font-weight: 600;
  }
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
  }
  .cost {
    font-size: var(--fs-xs);
    white-space: nowrap;
  }
  .soon {
    color: var(--color-warning);
    font-weight: 600;
  }
  .actions {
    display: flex;
  }
</style>
