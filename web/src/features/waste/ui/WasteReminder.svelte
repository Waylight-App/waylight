<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { WasteReminder } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /**
   * The reminder the evening before a collection (#866): off until turned
   * on, 19:00 by default. It goes out through the notification channels, so
   * each change is saved at once and needs no extra button.
   */
  interface Props {
    waste: WasteUseCases;
  }
  let { waste }: Props = $props();

  let reminder = $state<WasteReminder | null>(null);
  let error = $state('');
  let saved = $state(false);

  $effect(() => {
    waste
      .reminder()
      .then((r) => (reminder = r))
      .catch((err) => (error = errorMessage(err)));
  });

  async function save(next: WasteReminder) {
    error = '';
    saved = false;
    const before = reminder;
    reminder = next;
    try {
      reminder = await waste.saveReminder(next);
      saved = true;
    } catch (err) {
      reminder = before;
      error = errorMessage(err);
    }
  }
</script>

{#if reminder}
  {@const r = reminder}
  <fieldset>
    <legend>{t('waste.reminder')}</legend>
    <label class="row"
      ><input
        type="checkbox"
        checked={r.enabled}
        onchange={(e) => void save({ ...r, enabled: e.currentTarget.checked })}
      />{t('waste.reminder_enable')}
      <input
        class="input time"
        type="time"
        value={r.time}
        disabled={!r.enabled}
        aria-label={t('waste.reminder_time')}
        onchange={(e) => e.currentTarget.value && void save({ ...r, time: e.currentTarget.value })}
      /></label
    >
    <p class="muted small">{t('waste.reminder_hint')}</p>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if saved}<p class="muted small" role="status">{t('waste.reminder_saved')}</p>{/if}
  </fieldset>
{:else if error}
  <p class="error" role="alert">{error}</p>
{/if}

<style>
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    margin: var(--space-3) 0;
    display: grid;
    gap: var(--space-1);
  }
  .row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .time {
    width: auto;
  }
  .small {
    margin: 0;
    font-size: var(--fs-xs);
  }
</style>
