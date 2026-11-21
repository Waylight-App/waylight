<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, serverMessage, t } from '../../../shared/i18n';
  import { formatBytes } from '../../../shared/ui/format';
  import type { BackupFile, BackupPreview, RestoreSource } from '../domain/backup';
  import type { BackupUseCases } from '../usecase/backupUseCases';

  /** Props of the guided restore (#46, #677). */
  interface Props {
    backups: BackupUseCases;
    local: BackupFile[];
  }
  let { backups, local }: Props = $props();

  let remote = $state<BackupFile[]>([]);
  // "local:<name>", "remote:<name>" or "upload"; empty until the owner picks one.
  let choice = $state('');
  let file = $state<File | undefined>(undefined);
  let preview = $state<BackupPreview | null>(null);
  let previewError = $state('');
  let checking = $state(false);
  let understood = $state(false);
  let password = $state('');
  let code = $state('');
  let error = $state('');
  let busy = $state(false);
  let safetyCopy = $state<BackupFile | null>(null);

  $effect(() => {
    backups
      .listRemote()
      .then((l) => (remote = l))
      .catch(() => (remote = [])); // no off-site target
  });

  const source = $derived((choice.split(':', 1)[0] || '') as RestoreSource | '');
  const name = $derived(choice.includes(':') ? choice.slice(choice.indexOf(':') + 1) : '');
  const chosen = $derived(
    source === 'local'
      ? local.find((f) => f.name === name)
      : source === 'remote'
        ? remote.find((f) => f.name === name)
        : undefined,
  );

  // A local backup is checked as soon as it is picked; off-site backups and
  // uploads are checked by the server when the owner confirms.
  $effect(() => {
    const n = source === 'local' ? name : '';
    preview = null;
    previewError = '';
    understood = false;
    if (!n) return;
    let stale = false;
    checking = true;
    backups
      .preview(n)
      .then((p) => {
        if (!stale) preview = p;
      })
      .catch((err) => {
        if (!stale) previewError = errorMessage(err);
      })
      .finally(() => {
        if (!stale) checking = false;
      });
    return () => (stale = true);
  });

  const blocked = $derived(
    !choice || (source === 'local' && (checking || !preview?.restorable)) || (source === 'upload' && !file),
  );

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (blocked || !understood || !source) return;
    error = '';
    busy = true;
    try {
      const res = await backups.restore({
        password,
        code,
        confirmed: understood,
        source,
        name: name || undefined,
        file: source === 'upload' ? file : undefined,
      });
      safetyCopy = res.safetyCopy;
      password = code = '';
      setTimeout(() => location.reload(), 8000);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const date = (f: BackupFile) => f.createdAt.toLocaleString(intlLocale());
  const problemText = (p: NonNullable<BackupPreview['problem']>) =>
    serverMessage(p.messageKey, p.params) ?? p.message;
</script>

<h3>{t('backup.restore')}</h3>
{#if safetyCopy}
  <div class="notice" role="status">
    <p>{t('backup.restore_started', { date: date(safetyCopy) })}</p>
    <p>{t('backup.restarting')}</p>
  </div>
{:else}
  <p class="muted">{t('backup.restore_intro')}</p>
  <form onsubmit={submit}>
    <fieldset>
      <legend>{t('backup.restore_step_pick')}</legend>
      <ul class="choices">
        {#each local as f (f.name)}
          <li>
            <label class="choice">
              <input type="radio" name="restore-source" value={`local:${f.name}`} bind:group={choice} />
              <span class="what">{t('backup.source_local', { date: date(f) })}</span>
              <span class="muted">{formatBytes(f.size)}</span>
            </label>
          </li>
        {/each}
        {#each remote as f (f.name)}
          <li>
            <label class="choice">
              <input type="radio" name="restore-source" value={`remote:${f.name}`} bind:group={choice} />
              <span class="what">{t('backup.source_remote', { date: date(f) })}</span>
              <span class="muted">{formatBytes(f.size)}</span>
            </label>
          </li>
        {/each}
        <li>
          <label class="choice">
            <input type="radio" name="restore-source" value="upload" bind:group={choice} />
            <span class="what">{t('backup.source_upload')}</span>
          </label>
        </li>
      </ul>
      {#if source === 'upload'}
        <label class="field">
          <span>{t('backup.backup_file')}</span>
          <input
            class="input"
            type="file"
            accept=".db,application/vnd.sqlite3,application/octet-stream"
            required
            onchange={(e) => (file = (e.currentTarget as HTMLInputElement).files?.[0])}
          />
        </label>
      {/if}
    </fieldset>

    {#if choice}
      <fieldset>
        <legend>{t('backup.restore_step_check')}</legend>
        {#if source === 'local'}
          {#if checking}
            <p role="status">{t('backup.preview_checking')}</p>
          {:else if previewError}
            <p class="error" role="alert">{previewError}</p>
          {:else if preview?.problem}
            <p class="error" role="alert">{problemText(preview.problem)}</p>
          {:else if preview}
            <ul class="facts" aria-live="polite">
              <li>{t('backup.preview_ok')}</li>
              <li>
                {t('backup.preview_contents', {
                  tabs: preview.tabs,
                  widgets: preview.widgets,
                  modules: preview.modules,
                })}
              </li>
              {#if preview.updates > 0}
                <li>{t('backup.preview_updates', { count: preview.updates })}</li>
              {/if}
            </ul>
          {/if}
        {:else}
          <p class="muted">{t('backup.preview_on_confirm')}</p>
        {/if}
      </fieldset>
    {/if}

    {#if choice && !blocked}
      <fieldset>
        <legend>{t('backup.restore_step_confirm')}</legend>
        <div class="warning">
          <p>
            {chosen ? t('backup.restore_warning_since', { date: date(chosen) }) : t('backup.restore_warning')}
          </p>
          <p>{t('backup.restore_safety')}</p>
        </div>
        <label class="check">
          <input type="checkbox" bind:checked={understood} required />
          <span>{t('backup.restore_understood')}</span>
        </label>
        <label class="field">
          <span>{t('backup.confirm_password')}</span>
          <input
            class="input"
            type="password"
            bind:value={password}
            autocomplete="current-password"
            required
          />
        </label>
        <label class="field">
          <span>{t('backup.confirm_code')}</span>
          <input class="input" bind:value={code} inputmode="numeric" autocomplete="one-time-code" />
        </label>
        {#if error}<p class="error" role="alert">{error}</p>{/if}
        <button class="btn btn-danger" type="submit" disabled={busy || !understood || !password}
          >{busy ? t('backup.restoring') : t('backup.restore_now')}</button
        >
      </fieldset>
    {/if}
  </form>
{/if}

<style>
  fieldset {
    margin: 0 0 var(--space-3);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  legend {
    padding: 0 var(--space-1);
    font-weight: 600;
  }
  .choices,
  .facts {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .facts li {
    padding: 2px 0;
  }
  .choice,
  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 44px;
    cursor: pointer;
  }
  .choice .what {
    flex: 1;
  }
  .choice input,
  .check input {
    width: 20px;
    height: 20px;
    flex: none;
  }
  .warning {
    margin-bottom: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background: color-mix(in srgb, var(--color-danger) 12%, transparent);
    border: 1px solid var(--color-danger);
    border-radius: var(--radius-small);
  }
  .warning p {
    margin: var(--space-1) 0;
  }
  .notice {
    padding: var(--space-2) var(--space-3);
    background: var(--color-surface-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
</style>
