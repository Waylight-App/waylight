<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { degrees, formatTime, parseTonightConfig, warningLevel, worstWarning } from '../domain/tonight';
  import type { TonightUseCases } from '../usecase/tonightUseCases';
  import { useTonight } from './useTonight.svelte';

  /** Props of the compact Tonight view for small and strip tiles: time, temperature and a warning dot. */
  interface Props {
    tonight: TonightUseCases;
    config: Record<string, unknown>;
    /** One line for strip tiles. */
    strip?: boolean;
  }
  let { tonight, config, strip = false }: Props = $props();

  const cfg = $derived(parseTonightConfig(config));
  const data = useTonight(() => ({ tonight, config }));
  const warning = $derived(worstWarning(data.warnings));
</script>

<div class="lit" class:strip data-testid="tonight-compact" aria-label={t('tonight.label')}>
  <time class="time" datetime={data.now.toISOString()}>{formatTime(data.now, cfg, intlLocale())}</time>
  {#if data.weather}
    <span class="temp">{degrees(data.weather.temperature)}</span>
  {/if}
  {#if warning}
    <span
      class="dot"
      data-level={warningLevel(warning)}
      role="img"
      aria-label={`${t('tonight.warning')}: ${warning.title}`}
      title={warning.title}
    ></span>
  {/if}
</div>

<style>
  .lit {
    box-sizing: border-box;
    height: 100%;
    min-height: 100%;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    align-content: center;
    gap: var(--space-1) var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    background: var(--color-lit);
    color: var(--color-lit-text);
    box-shadow: var(--shadow);
    overflow: hidden;
  }
  .strip {
    flex-wrap: nowrap;
    align-content: normal;
    align-items: center;
  }
  .time,
  .temp {
    font-family: var(--font-display);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1.1;
  }
  .time {
    font-size: clamp(1.6rem, min(15cqi, 38cqh), 4rem);
  }
  .temp {
    font-size: clamp(1.2rem, min(9cqi, 24cqh), 2.4rem);
    color: var(--color-lit-muted);
  }
  .dot {
    align-self: center;
    flex: none;
    width: 0.8em;
    height: 0.8em;
    border-radius: 50%;
    background: var(--color-lit-warning);
  }
  .dot[data-level='problem'] {
    background: var(--color-lit-danger);
  }
</style>
