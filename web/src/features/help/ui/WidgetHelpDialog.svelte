<script lang="ts">
  import { t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { WidgetHelp } from '../domain/help';

  /** The short help of a widget kind (#678), opened from its "?" or the search. */
  interface Props {
    /** The widget's name in the interface. */
    name: string;
    /** Loads the help; resolves undefined when the server has none. */
    load: () => Promise<WidgetHelp | undefined>;
    /** Shown when the help cannot be loaded, e.g. the picker's description. */
    fallback: string;
    /** Where the documentation starts, for the fallback link. */
    docsUrl: string;
    onclose: () => void;
  }
  let { name, load, fallback, docsUrl, onclose }: Props = $props();

  // svelte-ignore state_referenced_locally
  const help = load().catch(() => undefined);
</script>

<Modal title={t('help.title', { name })} {onclose}>
  {#await help}
    <p class="muted">{t('help.loading')}</p>
  {:then h}
    {#if h}
      <h3>{t('help.shows')}</h3>
      <p>{h.shows}</p>
      <h3>{t('help.set_up')}</h3>
      <p>{h.setUp}</p>
      <p><a href={h.docsUrl} target="_blank" rel="noopener noreferrer">{t('help.docs')} ↗</a></p>
    {:else}
      <p>{fallback}</p>
      <p class="muted">{t('help.unavailable')}</p>
      <p><a href={docsUrl} target="_blank" rel="noopener noreferrer">{t('help.docs')} ↗</a></p>
    {/if}
  {/await}
</Modal>

<style>
  h3 {
    margin: var(--space-3) 0 var(--space-1);
    font-size: 0.95rem;
  }
  h3:first-child {
    margin-top: 0;
  }
  p {
    margin: 0 0 var(--space-2);
    line-height: 1.5;
  }
  a {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
  }
</style>
