<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { PresenceApi } from './presence';
  import { texts } from './texts';

  /** Admin switch of presence (#346) in the settings dialog. */
  interface Props {
    context: PluginContext;
  }
  let { context }: Props = $props();

  const api = $derived(new PresenceApi(context.http));
  const tr = (key: string) => text(texts, key, context.language);
  let enabled = $state<boolean | null>(null);
  let error = $state('');

  $effect(() => {
    api
      .enabled()
      .then((on) => (enabled = on))
      .catch((err) => (error = context.errorMessage(err)));
  });

  async function toggle(e: Event) {
    const on = (e.currentTarget as HTMLInputElement).checked;
    try {
      enabled = await api.setEnabled(on);
      error = '';
    } catch (err) {
      error = context.errorMessage(err);
    }
  }
</script>

<p class="muted">{tr('section_hint')}</p>
{#if enabled !== null}
  <label><input type="checkbox" checked={enabled} onchange={toggle} /> {tr('enabled')}</label>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
