<script lang="ts">
  import { t } from '../i18n';
  import { newId } from '../id/uuid';
  import type { ChoiceLoader, ConnectionChoice } from './connection';

  /**
   * "Which connection?" in the settings of a widget that reads a service
   * (#709). One select for a single connection (`value`), checkboxes for a
   * widget that merges several (`values`). It stays hidden while the
   * service has only one connection and none is chosen, so a simple
   * installation sees nothing new.
   */
  interface Props {
    /** Service keys whose connections are offered, e.g. ['immich']. */
    services: string[];
    load: ChoiceLoader;
    /** Chosen connection; '' reads the module settings. */
    value?: string;
    /** Chosen connections of a widget that merges several; none means all. */
    values?: string[];
    /** Show the choice even with a single connection, for settings that offer nothing else. */
    always?: boolean;
  }
  let { services, load, value = $bindable(''), values = $bindable(), always = false }: Props = $props();

  let choices = $state<ConnectionChoice[]>([]);
  // svelte-ignore state_referenced_locally
  const loading = Promise.all(services.map((s) => load(s).catch(() => [])));
  void loading.then((lists) => (choices = lists.flat()));

  const visible = $derived(always || choices.length > 1 || value !== '' || (values?.length ?? 0) > 0);
  const labelId = `connection-${newId()}`;

  function toggle(id: string, on: boolean) {
    const rest = (values ?? []).filter((v) => v !== id);
    values = on ? [...rest, id] : rest;
  }
</script>

{#if visible}
  {#if values}
    <fieldset class="field">
      <legend>{t('connection.which')}</legend>
      {#each choices as c (c.id)}
        <label class="check">
          <input
            type="checkbox"
            checked={values.includes(c.id)}
            onchange={(e) => toggle(c.id, e.currentTarget.checked)}
          />
          {c.name}
        </label>
      {/each}
      <p class="muted hint">{t('connection.all_hint')}</p>
    </fieldset>
  {:else}
    <label class="field" for={labelId}>
      {t('connection.which')}
      <select id={labelId} bind:value>
        <option value="">{t('connection.default')}</option>
        {#each choices as c (c.id)}
          <option value={c.id}>{c.name}</option>
        {/each}
      </select>
    </label>
  {/if}
{/if}

<style>
  .field {
    display: block;
    border: 0;
    padding: 0;
    margin: 0 0 var(--space-2);
  }
  .check {
    display: block;
  }
  .hint {
    margin: 0;
  }
</style>
