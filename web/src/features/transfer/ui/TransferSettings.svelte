<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { describeConfigResult, describeResult } from '../domain/transfer';
  import type { TransferUseCases } from '../usecase/transferUseCases';

  /** Props of the export/import section. */
  interface Props {
    transfer: TransferUseCases;
  }
  let { transfer }: Props = $props();

  let message = $state('');
  let error = $state('');
  let busy = $state(false);
  let configMessage = $state('');
  let configError = $state('');
  let configBusy = $state(false);

  async function onconfig(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (!confirm(t('transfer.config_confirm'))) return;
    configBusy = true;
    configError = configMessage = '';
    try {
      const res = await transfer.applyConfigFile(await file.text());
      configMessage = `${describeConfigResult(res)} ${t('transfer.reload_hint')}`;
    } catch (err) {
      configError = errorMessage(err);
    } finally {
      configBusy = false;
    }
  }

  async function onfile(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (!confirm(t('transfer.confirm_import'))) return;
    busy = true;
    error = message = '';
    try {
      const res = await transfer.importFile(await file.text());
      message = `${describeResult(res)} ${t('transfer.reload_hint')}`;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<h4>{t('transfer.export')}</h4>
<p class="muted">
  {t('transfer.downloads_your_layout_bookmarks_notes_to')}
</p>
<p><a class="btn" href={transfer.exportUrl()} download>{t('transfer.download_export')}</a></p>

<h4>{t('transfer.import')}</h4>
<p class="muted">
  {t('transfer.restores_an_export_e_g_on')}
</p>
<label class="btn file">
  {busy ? t('common.importing') : t('transfer.choose_file')}
  <input type="file" accept="application/json,.json" onchange={onfile} disabled={busy} />
</label>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<h4>{t('transfer.config_title')}</h4>
<p class="muted">{t('transfer.config_intro')}</p>
<p class="actions">
  <a class="btn" href={transfer.configUrl()} download="trisse-home.yaml">{t('transfer.config_download')}</a>
  <label class="btn file">
    {configBusy ? t('common.importing') : t('transfer.config_apply')}
    <input
      type="file"
      accept=".yaml,.yml,application/yaml,text/yaml"
      onchange={onconfig}
      disabled={configBusy}
    />
  </label>
</p>
{#if configError}<p class="error" role="alert">{configError}</p>{/if}
{#if configMessage}<p class="muted" role="status">{configMessage}</p>{/if}

<style>
  h4 {
    margin: var(--space-3) 0 var(--space-1);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
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
</style>
