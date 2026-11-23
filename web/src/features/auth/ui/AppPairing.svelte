<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import { qrDataUrl } from '../../../shared/ui/qr';
  import { pairingLink, type PairingCode } from '../domain/appPairing';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the pairing of the mobile app (#559). */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let pairing = $state<PairingCode | null>(null);
  let error = $state('');

  // The code is only fetched on request: every code is a credential for a
  // few minutes, so none is created just by opening the settings.
  async function show() {
    error = '';
    try {
      pairing = await auth.newPairingCode();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const qr = $derived(pairing ? qrDataUrl(pairingLink(location.origin, pairing.code), 4) : '');
</script>

<p class="muted">{t('auth.pair_app_hint')}</p>
{#if pairing}
  <img class="qr" src={qr} alt={t('auth.pair_app_qr_alt')} />
  <!-- Devices without a camera, such as a watch, take the code typed in (#605). -->
  <p class="muted small">{t('auth.pair_app_type_hint')}</p>
  <div class="code"><CopySnippet code={pairing.code} label={t('auth.pair_app_code')} /></div>
  <p class="muted small">
    {t('auth.pair_app_valid_until', {
      value: pairing.expiresAt.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' }),
    })}
  </p>
  <button class="btn" type="button" onclick={show}>{t('auth.pair_app_new_code')}</button>
{:else}
  <button class="btn" type="button" onclick={show}>{t('auth.pair_app_show_code')}</button>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .qr {
    display: block;
    width: 180px;
    height: 180px;
    image-rendering: pixelated;
    background: #fff; /* scanners need a light quiet zone in dark themes */
    border-radius: var(--radius-sm);
  }
  .code {
    max-width: 16rem;
    font-size: var(--fs-md);
    letter-spacing: 0.08em;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
