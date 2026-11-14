<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import {
    addsAnything,
    describeConfigResult,
    FOREIGN_FORMATS,
    type ForeignFormat,
    type ForeignImport,
  } from '../domain/transfer';
  import type { TransferUseCases } from '../usecase/transferUseCases';

  /** Import from Homepage or Homarr (#77): preview first, then add. */
  interface Props {
    transfer: TransferUseCases;
  }
  let { transfer }: Props = $props();

  let format = $state<ForeignFormat>('homepage');
  let files = $state<File[]>([]);
  let preview = $state<ForeignImport | null>(null);
  let message = $state('');
  let error = $state('');
  let busy = $state(false);

  async function onfiles(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    files = [...(input.files ?? [])];
    input.value = '';
    preview = null;
    message = error = '';
    if (files.length === 0) return;
    busy = true;
    try {
      preview = await transfer.previewForeign(format, files);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function apply() {
    busy = true;
    error = '';
    try {
      const res = await transfer.applyForeign(format, files);
      message = res.result ? `${describeConfigResult(res.result)} ${t('transfer.reload_hint')}` : '';
      preview = null;
      files = [];
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<h4>{t('transfer.foreign_title')}</h4>
<p class="muted">{t('transfer.foreign_intro')}</p>
<div class="row">
  <label class="field"
    ><span>{t('transfer.foreign_format')}</span><select
      class="input"
      bind:value={format}
      onchange={() => ((preview = null), (files = []))}
    >
      {#each FOREIGN_FORMATS as f (f)}<option value={f}>{t(`transfer.foreign_${f}`)}</option>{/each}
    </select></label
  >
  <label class="btn file">
    {busy ? t('common.importing') : t('transfer.foreign_choose')}
    <input
      type="file"
      multiple={format === 'homepage'}
      accept={format === 'homepage' ? '.yaml,.yml' : '.json,application/json'}
      onchange={onfiles}
      disabled={busy}
    />
  </label>
</div>
<p class="muted small">{t(`transfer.foreign_hint_${format}`)}</p>

{#if preview}
  {@const s = preview.summary}
  <div class="preview" role="status">
    <p>
      {t('transfer.foreign_summary', {
        tiles: s.tiles,
        bookmarks: s.bookmarks,
        widgets: s.widgets,
        skipped: s.skipped,
      })}
      {#if s.tab}{t('transfer.foreign_tab', { tab: s.tab })}{/if}
    </p>
    {#if s.warnings.length}
      <details>
        <summary>{t('transfer.foreign_warnings', { n: s.warnings.length })}</summary>
        <ul>
          {#each s.warnings as w, i (i)}<li>{w}</li>{/each}
        </ul>
      </details>
    {/if}
    {#if preview.yaml}
      <details>
        <summary>{t('transfer.foreign_yaml')}</summary>
        <pre>{preview.yaml}</pre>
      </details>
    {/if}
    <button class="btn btn-primary" type="button" onclick={apply} disabled={busy || !addsAnything(s)}
      >{t('transfer.foreign_apply')}</button
    >
  </div>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  h4 {
    margin: var(--space-3) 0 var(--space-1);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    gap: var(--space-2);
  }
  .file {
    position: relative;
  }
  .file input {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .file:focus-within {
    outline: 2px solid var(--color-accent);
  }
  .small {
    font-size: 0.85em;
  }
  .preview {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
    border-left: 3px solid var(--color-accent);
    padding-left: var(--space-2);
  }
  pre {
    max-height: 16rem;
    overflow: auto;
    font-size: 0.8em;
  }
</style>
