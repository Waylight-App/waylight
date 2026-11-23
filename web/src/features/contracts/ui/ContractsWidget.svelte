<script lang="ts">
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
    <li class:soon={c.soon} class:ended={!c.open}>
      <span class="main">
        <span class="name"
          >{c.name}{#if c.provider}<span class="muted"> · {c.provider}</span>{/if}</span
        >
        <span class="meta">
          {#if c.open}
            {t('contracts.cancel_by', { date: date(c.deadline) })}
          {:else}
            {t('contracts.ended', { date: date(c.currentEnd) })}
          {/if}
        </span>
      </span>
      <span class="cost muted">{money(c.costCents)} {t(`contracts.per.${c.period}`)}</span>
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
    </li>
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
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
  }
  .main {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .name,
  .meta {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta,
  .cost {
    font-size: var(--fs-xs);
  }
  .cost {
    white-space: nowrap;
  }
  li.soon .meta {
    color: var(--color-warning);
    font-weight: 600;
  }
  li.ended {
    opacity: 0.6;
  }
  .actions {
    display: flex;
  }
</style>
