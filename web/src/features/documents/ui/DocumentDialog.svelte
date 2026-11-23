<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import Modal from '../../../shared/ui/Modal.svelte';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import { type Doc, type DocInput, FILE_ACCEPT, validateDoc, validateFile } from '../domain/documents';
  import type { DocumentsUseCases } from '../usecase/documentsUseCases';

  /** Props of the dialog that enters or changes a document (#688). */
  interface Props {
    documents: DocumentsUseCases;
    /** The document to change, or null for a new one. */
    doc: Doc | null;
    onsaved: () => void;
    onclose: () => void;
  }
  let { documents, doc, onsaved, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<DocInput>(
    doc
      ? { title: doc.title, purchasedOn: doc.purchasedOn, warrantyEnd: doc.warrantyEnd, note: doc.note }
      : { title: '', purchasedOn: '', warrantyEnd: '', note: '' },
  );
  // svelte-ignore state_referenced_locally
  let visibility = $state(doc?.visibility ?? '');
  let file = $state<File | null>(null);
  let groups = $state<ShareGroup[]>([]);
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    documents
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    const invalid = validateDoc(form) ?? (file ? validateFile(file) : null);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await documents.save(doc?.id ?? null, form, file, visibility || undefined);
      onsaved();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={doc ? t('documents.edit') : t('documents.new')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('documents.title')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={form.title} maxlength="80" required autofocus />
    </label>
    <div class="row">
      <label class="field">
        <span>{t('documents.purchased_on')} {t('common.optional')}</span>
        <input class="input" type="date" bind:value={form.purchasedOn} />
      </label>
      <label class="field">
        <span>{t('documents.warranty_end')} {t('common.optional')}</span>
        <input class="input" type="date" bind:value={form.warrantyEnd} />
      </label>
    </div>
    <label class="field">
      <span>{t('documents.note')} {t('common.optional')}</span>
      <textarea class="input" rows="2" bind:value={form.note} maxlength="500"></textarea>
    </label>
    <label class="field">
      <span>{t('documents.file')} {t('common.optional')}</span>
      <input
        class="input"
        type="file"
        accept={FILE_ACCEPT}
        onchange={(e) => (file = e.currentTarget.files?.[0] ?? null)}
      />
      {#if doc?.fileName}
        <span class="muted small">{t('documents.file_replace', { name: doc.fileName })}</span>
      {/if}
    </label>
    {#if doc?.mine !== false && (visibility || !doc)}
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
    font-size: var(--fs-xs);
  }
</style>
