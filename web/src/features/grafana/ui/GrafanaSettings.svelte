<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { GrafanaUseCases } from '../usecase/grafanaUseCases';

  /** Props of the Grafana connection settings. */
  interface Props {
    grafana: GrafanaUseCases;
  }
  let { grafana }: Props = $props();

  let url = $state('');
  let token = $state('');
  let hasToken = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    grafana
      .settings()
      .then((s) => ((url = s.url), (hasToken = s.hasToken)))
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await grafana.saveSettings(url, token);
      hasToken = s.hasToken;
      token = '';
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field"
    ><span>{t('grafana.url')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.1.2:3000"
      required
    /></label
  >
  <label class="field">
    <span>{t('grafana.token', { value: hasToken ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={token}
      autocomplete="new-password"
      required={!hasToken}
    />
  </label>
  <p class="muted">{t('grafana.token_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
