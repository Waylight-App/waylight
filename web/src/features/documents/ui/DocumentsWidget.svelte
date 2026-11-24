<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { type Doc, formatSize } from '../domain/documents';
  import type { DocumentsUseCases } from '../usecase/documentsUseCases';
  import DocumentDialog from './DocumentDialog.svelte';

  /** Props of the warranty and documents widget (#688). */
  interface Props {
    documents: DocumentsUseCases;
  }
  let { documents }: Props = $props();
  const poll = usePolling();

  let items = $state<Doc[]>([]);
  let error = $state('');
  let editing = $state<Doc | null | undefined>(undefined); // null = new

  const date = (day: string) =>
    new Date(`${day}T12:00:00`).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  const warranty = (d: Doc) =>
    d.status === 'none'
      ? t('documents.status.none')
      : t(`documents.status.${d.status}`, { date: date(d.warrantyEnd) });

  async function load() {
    try {
      items = await documents.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => poll(load, 30 * 60_000));

  async function remove(d: Doc) {
    if (!confirm(t('common.confirm_delete', { name: d.title }))) return;
    try {
      await documents.remove(d.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
<ul>
  {#each items as d (d.id)}
    <ListRow nowrap style={d.status === 'ended' ? 'opacity: 0.6' : ''}>
      {d.title}
      {#snippet sub()}<span class:ending={d.status === 'ending'}>{warranty(d)}</span>{/snippet}
      {#snippet trail()}
        <span class="actions">
          {#if d.fileName}
            <a
              class="btn btn-icon"
              href={documents.fileUrl(d.id)}
              download={d.fileName}
              title={formatSize(d.fileSize)}
              aria-label={t('documents.download', { name: d.fileName })}><Icon name="download" size={16} /></a
            >
          {/if}
          {#if d.mine}
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('documents.edit_named', { name: d.title })}
              onclick={() => (editing = d)}>✎</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('documents.remove', { name: d.title })}
              onclick={() => remove(d)}>✕</button
            >
          {/if}
        </span>
      {/snippet}
    </ListRow>
  {:else}
    <li class="muted">{t('documents.empty')}</li>
  {/each}
</ul>
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('documents.new')}</button>

{#if editing !== undefined}
  <DocumentDialog
    {documents}
    doc={editing}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = undefined)}
  />
{/if}

<style>
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
  }
  .ending {
    color: var(--color-warning);
    font-weight: 600;
  }
  .actions {
    display: flex;
  }
</style>
