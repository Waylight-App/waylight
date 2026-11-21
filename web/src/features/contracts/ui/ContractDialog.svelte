<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import Modal from '../../../shared/ui/Modal.svelte';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import {
    centsToInput,
    type Contract,
    type ContractInput,
    PERIODS,
    validateContract,
  } from '../domain/contracts';
  import type { ContractsUseCases } from '../usecase/contractsUseCases';

  /** Props of the dialog that enters or changes a contract (#687). */
  interface Props {
    contracts: ContractsUseCases;
    /** The contract to change, or null for a new one. */
    contract: Contract | null;
    currency: string;
    onsaved: () => void;
    onclose: () => void;
  }
  let { contracts, contract, currency, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<ContractInput>(
    contract
      ? {
          name: contract.name,
          provider: contract.provider,
          cost: centsToInput(contract.costCents),
          period: contract.period,
          termEnd: contract.termEnd,
          noticeDays: contract.noticeDays,
          renewMonths: contract.renewMonths,
          remindDays: contract.remindDays,
          note: contract.note,
        }
      : {
          name: '',
          provider: '',
          cost: '',
          period: 'monthly',
          termEnd: '',
          noticeDays: 30,
          renewMonths: 12,
          remindDays: 14,
          note: '',
        },
  );
  // svelte-ignore state_referenced_locally
  let visibility = $state(contract?.visibility ?? '');
  let groups = $state<ShareGroup[]>([]);
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    contracts
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateContract(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await contracts.save(contract?.id ?? null, form, visibility || undefined);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={contract ? t('contracts.edit') : t('contracts.new')} {onclose}>
  <form onsubmit={submit}>
    <div class="row">
      <label class="field">
        <span>{t('contracts.name')}</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input class="input" bind:value={form.name} maxlength="80" required autofocus />
      </label>
      <label class="field">
        <span>{t('contracts.provider')} {t('common.optional')}</span>
        <input class="input" bind:value={form.provider} maxlength="80" />
      </label>
    </div>
    <div class="row">
      <label class="field">
        <span>{t('contracts.cost', { currency })}</span>
        <input class="input" inputmode="decimal" bind:value={form.cost} placeholder="0.00" />
      </label>
      <label class="field">
        <span>{t('contracts.period')}</span>
        <select class="input" bind:value={form.period}>
          {#each PERIODS as p (p)}<option value={p}>{t(`contracts.period.${p}`)}</option>{/each}
        </select>
      </label>
    </div>
    <div class="row">
      <label class="field">
        <span>{t('contracts.term_end')}</span>
        <input class="input" type="date" bind:value={form.termEnd} required />
      </label>
      <label class="field">
        <span>{t('contracts.notice_days')}</span>
        <input class="input" type="number" min="0" max="730" bind:value={form.noticeDays} />
      </label>
    </div>
    <div class="row">
      <label class="field">
        <span>{t('contracts.renew_months')}</span>
        <input class="input" type="number" min="0" max="60" bind:value={form.renewMonths} />
      </label>
      <label class="field">
        <span>{t('contracts.remind_days')}</span>
        <input class="input" type="number" min="0" max="90" bind:value={form.remindDays} />
      </label>
    </div>
    <p class="muted small">{t('contracts.deadline_hint')}</p>
    <label class="field">
      <span>{t('contracts.note')} {t('common.optional')}</span>
      <textarea class="input" rows="2" bind:value={form.note} maxlength="500"></textarea>
    </label>
    {#if contract?.mine !== false && (visibility || !contract)}
      <VisibilitySelect value={visibility || 'private'} {groups} onchange={(v) => (visibility = v)} />
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
  </form>
</Modal>

<style>
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row .field {
    flex: 1 1 9rem;
  }
  .small {
    font-size: 0.8rem;
  }
</style>
