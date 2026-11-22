<script lang="ts">
  import type { ChoiceLoader } from '../../../shared/connection/connection';
  import ConnectionSelect from '../../../shared/connection/ConnectionSelect.svelte';
  import { t } from '../../../shared/i18n';
  import { parseWidgetConfig } from '../domain/media';

  /** Props of the media widget settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
    connections: ChoiceLoader;
  }
  let { config, onsave, connections }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseWidgetConfig(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <ConnectionSelect services={['jellyfin', 'plex']} load={connections} bind:value={form.connectionId} />
  <label class="check"
    ><input type="checkbox" bind:checked={form.showSessions} /> {t('media.now_playing')}</label
  >
  <label class="check"
    ><input type="checkbox" bind:checked={form.showRecent} /> {t('media.recently_added')}</label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .check {
    display: block;
  }
</style>
