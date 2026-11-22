<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import Modal from '../../../shared/ui/Modal.svelte';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import { type BoardNote, type BoardNoteInput, MAX_TEXT_LENGTH, validateNote } from '../domain/board';
  import type { BoardUseCases } from '../usecase/boardUseCases';

  /** Props of the dialog that writes or changes a note (#685). */
  interface Props {
    board: BoardUseCases;
    /** The note to change, or null for a new one. */
    note: BoardNote | null;
    today: string;
    onsaved: () => void;
    onclose: () => void;
  }
  let { board, note, today, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<BoardNoteInput>(
    note
      ? { text: note.text, pinned: note.pinned, expiresOn: note.expiresOn }
      : { text: '', pinned: false, expiresOn: '' },
  );
  // svelte-ignore state_referenced_locally
  let visibility = $state(note?.visibility ?? 'everyone');
  let groups = $state<ShareGroup[]>([]);
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    board
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateNote(form, today);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await board.save(note?.id ?? null, form, visibility);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={note ? t('board.edit') : t('board.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('board.text')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <textarea class="input" rows="4" bind:value={form.text} maxlength={MAX_TEXT_LENGTH} required autofocus
      ></textarea>
    </label>
    <label class="field">
      <span>{t('board.expires')} {t('common.optional')}</span>
      <input class="input" type="date" min={today} bind:value={form.expiresOn} />
    </label>
    <label class="check"><input type="checkbox" bind:checked={form.pinned} /> {t('board.pin')}</label>
    <VisibilitySelect value={visibility} {groups} onchange={(v) => (visibility = v)} />
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
  </form>
</Modal>

<style>
  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 44px;
  }
</style>
