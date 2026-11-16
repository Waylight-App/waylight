<script lang="ts">
  import { t, type MessageKey } from '../../shared/i18n';

  /**
   * What a server adds to the extension's local mode (ADR 0019, #291).
   * Static text bundled with the extension: no request leaves the browser
   * until someone follows a link. "short" is a one-line hint that can be
   * dismissed; "full" is the settings section and always shown.
   */
  interface Props {
    variant: 'short' | 'full';
  }
  let { variant }: Props = $props();

  const DISMISSED_KEY = 'trisse-home.fullVersionHintDismissed';
  const docs = 'https://github.com/itadori26/trisse-home/blob/main/docs/';
  const serverGuide = `${docs}getting-started.md`;
  const piGuide = `${docs}raspberry-pi.md`;

  // Storage can be unavailable (private windows, blocked site data); the
  // hint then simply shows again next time.
  function readDismissed(): boolean {
    try {
      return localStorage.getItem(DISMISSED_KEY) === '1';
    } catch {
      return false;
    }
  }
  let dismissed = $state(readDismissed());

  function dismiss() {
    dismissed = true;
    try {
      localStorage.setItem(DISMISSED_KEY, '1');
    } catch {
      // Not remembered; nothing else to do.
    }
  }

  const features: MessageKey[] = [
    'full_version.feature_services',
    'full_version.feature_docker',
    'full_version.feature_parcels',
    'full_version.feature_notifications',
    'full_version.feature_devices',
    'full_version.feature_sync',
  ];
</script>

{#if variant === 'full'}
  <p>{t('full_version.intro')}</p>
  <ul>
    {#each features as key (key)}<li>{t(key)}</li>{/each}
  </ul>
  <p>
    <a href={serverGuide} target="_blank" rel="noopener noreferrer">{t('full_version.server_guide')}</a>
    ·
    <a href={piGuide} target="_blank" rel="noopener noreferrer">{t('full_version.pi_guide')}</a>
  </p>
  <p class="muted">{t('full_version.local_note')}</p>
{:else if !dismissed}
  <p class="hint" role="note">
    <span>{t('full_version.short')}</span>
    <a href={serverGuide} target="_blank" rel="noopener noreferrer">{t('full_version.learn_more')}</a>
    <button class="btn btn-icon" type="button" aria-label={t('full_version.dismiss')} onclick={dismiss}
      >✕</button
    >
  </p>
{/if}

<style>
  .hint {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
    margin: var(--space-3) 0 0;
    padding: var(--space-2);
    border-radius: var(--radius);
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
    font-size: 0.9em;
  }
  .hint span {
    flex: 1;
    min-width: 12rem;
  }
</style>
