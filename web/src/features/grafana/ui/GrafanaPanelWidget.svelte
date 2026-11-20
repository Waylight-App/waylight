<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { hasPanel, panelLink, parsePanelConfig, renderUrl } from '../domain/grafana';

  /** Props of the Grafana panel widget (#574). */
  interface Props {
    config: Record<string, unknown>;
  }
  let { config }: Props = $props();
  const cfg = $derived(parsePanelConfig(config));

  let width = $state(0);
  let height = $state(0);
  let stamp = $state(Date.now());
  let failed = $state(false);
  // The image follows the light or dark look of the page.
  let dark = $state(document.documentElement.dataset['theme'] === 'dark');

  $effect(() => {
    const observer = new MutationObserver(
      () => (dark = document.documentElement.dataset['theme'] === 'dark'),
    );
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    return () => observer.disconnect();
  });

  $effect(() => {
    const id = setInterval(() => {
      if (document.visibilityState === 'visible') stamp = Date.now();
    }, cfg.refreshMinutes * 60_000);
    return () => clearInterval(id);
  });

  // Rounded to 50 px, so resizing the widget does not render on every pixel.
  const size = $derived({
    width: Math.max(200, Math.round(width / 50) * 50),
    height: Math.max(100, Math.round(height / 50) * 50),
  });
  const src = $derived(hasPanel(cfg) && width > 0 ? renderUrl(cfg, size, dark, stamp) : '');
  const link = $derived(panelLink(cfg));
</script>

<div class="panel" bind:clientWidth={width} bind:clientHeight={height}>
  {#if !hasPanel(cfg)}
    <p class="muted">{t('grafana.choose_panel')}</p>
  {:else if failed}
    <p class="muted">{t('srv_8ff75b8c58539e49' as never)}</p>
    {#if link}<a href={link} target="_blank" rel="noopener noreferrer">{t('grafana.open_in_grafana')}</a>{/if}
  {:else if src}
    <a href={link || undefined} target="_blank" rel="noopener noreferrer">
      <img
        {src}
        alt={t('grafana.panel_alt', { title: cfg.panelTitle })}
        onerror={() => (failed = true)}
        onload={() => (failed = false)}
      />
    </a>
  {/if}
</div>

<style>
  .panel {
    width: 100%;
    height: 100%;
    min-height: 6rem;
  }
  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
</style>
