<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseConfig, symbolsFromText } from '../domain/markets';

  /** Props of the markets widget settings: the watchlist. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let text = $state(parseConfig(config).symbols.join('\n'));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ symbols: symbolsFromText(text).slice(0, 12) });
  }}
>
  <label class="field"
    ><span>{t('markets.symbols')}</span><textarea class="input" rows="6" bind:value={text} spellcheck="false"
    ></textarea></label
  >
  <p class="muted small">{t('markets.symbols_hint')}</p>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  textarea {
    font-family: ui-monospace, monospace;
  }
  .small {
    font-size: max(0.8rem, 0.85em);
  }
</style>
