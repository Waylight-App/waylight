<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { usePolling } from '../../../shared/refresh';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import {
    formatPanelValue,
    hasPanel,
    latest,
    panelLink,
    type PanelValues,
    parsePanelConfig,
    renderUrl,
  } from '../domain/grafana';
  import type { GrafanaUseCases } from '../usecase/grafanaUseCases';

  /** Props of the Grafana panel widget (#574). */
  interface Props {
    config: Record<string, unknown>;
    grafana: GrafanaUseCases;
  }
  let { config, grafana }: Props = $props();
  const cfg = $derived(parsePanelConfig(config));
  const poll = usePolling();

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
  const src = $derived(hasPanel(cfg) && width > 0 && !failed ? renderUrl(cfg, size, dark, stamp) : '');
  const link = $derived(panelLink(cfg));

  // Without the image renderer the widget shows the panel's values instead
  // (#579): the newest value as a stat and the first series as a sparkline.
  let values = $state<PanelValues | null>(null);
  let valuesError = $state('');
  $effect(() => {
    if (!failed || !hasPanel(cfg)) return;
    const c = cfg;
    return poll(
      () =>
        grafana
          .panelValues(c)
          .then((v) => ((values = v), (valuesError = '')))
          .catch((err) => {
            valuesError = errorMessage(err);
            throw err;
          }),
      c.refreshMinutes * 60_000,
    );
  });
  const main = $derived(values?.series[0]);
  const format = (v: number | null) =>
    v === null || !values ? '–' : formatPanelValue(v, values.unit, values.decimals, intlLocale());
</script>

<div class="panel" bind:clientWidth={width} bind:clientHeight={height}>
  {#if !hasPanel(cfg)}
    <p class="muted">{t('grafana.choose_panel')}</p>
  {:else if failed}
    {#if values && main}
      <div class="values">
        <span class="title">{cfg.panelTitle || values.title}</span>
        <span class="stat">{format(latest(main))}</span>
        {#if values.series.length > 1}<span class="muted name">{main.name}</span>{/if}
        <span class="spark">
          <Sparkline
            values={main.points.map((p) => p.v)}
            label={t('grafana.values_over', { title: cfg.panelTitle || values.title, range: cfg.range })}
          />
        </span>
        {#if values.series.length > 1}
          <ul>
            {#each values.series.slice(1) as s, i (s.name + i)}
              <li><span class="muted">{s.name}</span> <strong>{format(latest(s))}</strong></li>
            {/each}
          </ul>
        {/if}
        {#if link}<a href={link} target="_blank" rel="noopener noreferrer">{t('grafana.open_in_grafana')}</a
          >{/if}
      </div>
    {:else if valuesError}
      <p class="muted">{valuesError}</p>
      {#if link}<a href={link} target="_blank" rel="noopener noreferrer">{t('grafana.open_in_grafana')}</a
        >{/if}
    {:else}
      <p class="muted">{t('grafana.loading_values')}</p>
    {/if}
  {:else if src}
    <a href={link || undefined} target="_blank" rel="noopener noreferrer">
      <img {src} alt={t('grafana.panel_alt', { title: cfg.panelTitle })} onerror={() => (failed = true)} />
    </a>
  {/if}
</div>

<style>
  .panel {
    width: 100%;
    height: 100%;
    min-height: 6rem;
  }
  .values {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    height: 100%;
    min-width: 0;
  }
  .title {
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
  .stat {
    font-family: var(--font-display);
    font-size: var(--fs-lg);
    font-weight: 600;
    line-height: 1.1;
  }
  .spark {
    flex: 1;
    min-height: 2rem;
    color: var(--color-accent-text);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    font-size: var(--fs-xs);
  }
  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
</style>
