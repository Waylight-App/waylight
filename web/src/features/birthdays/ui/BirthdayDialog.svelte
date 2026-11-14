<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { type BirthdayInput, MAX_NAME_LENGTH } from '../domain/birthdays';
  import { type BirthdaysUseCases, InvalidBirthday } from '../usecase/birthdaysUseCases';

  /** Props of the dialog that adds a birthday by hand (#287). */
  interface Props {
    birthdays: BirthdaysUseCases;
    onsaved: () => void;
    onclose: () => void;
  }
  let { birthdays, onsaved, onclose }: Props = $props();

  let form = $state<BirthdayInput>({ name: '', date: '', yearKnown: true });
  let error = $state('');
  let busy = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    try {
      await birthdays.add(form);
      onsaved();
      onclose();
    } catch (err) {
      error = err instanceof InvalidBirthday ? t(err.key) : errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('birthdays.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('birthdays.name')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.name} maxlength={MAX_NAME_LENGTH} required autofocus />
    </label>
    <label class="field">
      <span>{t('birthdays.date')}</span>
      <input class="input" type="date" bind:value={form.date} required />
    </label>
    <label class="check"
      ><input type="checkbox" bind:checked={form.yearKnown} /> {t('birthdays.year_known')}</label
    >
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
  </form>
</Modal>

<style>
  form {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
  }
  form .field {
    justify-self: stretch;
  }
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
</style>
