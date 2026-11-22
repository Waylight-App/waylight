<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import Modal from '../../../shared/ui/Modal.svelte';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import { KINDS, type Meter, type MeterInput, validateMeter } from '../domain/meters';
  import type { MetersUseCases } from '../usecase/metersUseCases';

  /** Props of the dialog that creates or changes a meter (#689). */
  interface Props {
    meters: MetersUseCases;
    /** The meter to change, or null for a new one. */
    meter: Meter | null;
    onsaved: (saved: Meter) => void;
    onclose: () => void;
  }
  let { meters, meter, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<MeterInput>(
    meter
      ? { name: meter.name, kind: meter.kind, unit: meter.unit }
      : { name: '', kind: 'electricity', unit: 'kWh' },
  );
  // A new meter belongs to the household, so it starts shared with everyone.
  // svelte-ignore state_referenced_locally
  let visibility = $state(meter?.visibility ?? 'everyone');
  let groups = $state<ShareGroup[]>([]);
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    meters
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateMeter(form);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      onsaved(await meters.save(meter?.id ?? null, form, visibility));
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={meter ? t('meters.edit') : t('meters.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('meters.name')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.name} maxlength="60" required autofocus />
    </label>
    <div class="row">
      <label class="field">
        <span>{t('meters.kind')}</span>
        <select class="input" bind:value={form.kind}>
          {#each KINDS as k (k)}<option value={k}>{t(`meters.kind.${k}`)}</option>{/each}
        </select>
      </label>
      <label class="field">
        <span>{t('meters.unit')}</span>
        <input
          class="input"
          bind:value={form.unit}
          maxlength="12"
          required
          placeholder={t('meters.unit_hint')}
        />
      </label>
    </div>
    {#if meter?.mine !== false}
      <VisibilitySelect value={visibility} {groups} onchange={(v) => (visibility = v)} />
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
</style>
