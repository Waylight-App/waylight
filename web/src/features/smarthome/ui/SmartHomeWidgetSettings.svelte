<script lang="ts">
  import { type ChoiceLoader, withConnectionChoice } from '../../../shared/connection/connection';
  import ConnectionSelect from '../../../shared/connection/ConnectionSelect.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { type Entity, filterEntities, MAX_ENTITIES, parseWidgetConfig } from '../domain/smarthome';
  import type { SmarthomeUseCases } from '../usecase/smarthomeUseCases';

  /** Props of the smart home widget settings: which entities to show. */
  interface Props {
    smarthome: SmarthomeUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
    connections: ChoiceLoader;
  }
  let { smarthome, config, onsave, connections }: Props = $props();
  // svelte-ignore state_referenced_locally
  let selected = $state<string[]>(parseWidgetConfig(config).entities);
  // svelte-ignore state_referenced_locally
  let connectionId = $state(parseWidgetConfig(config).connectionId);
  let all = $state<Entity[]>([]);
  let query = $state('');
  let error = $state('');
  const shown = $derived(filterEntities(all, query).slice(0, 200));

  $effect(() => {
    smarthome
      .entities(connectionId)
      .then((l) => (all = l))
      .catch((err) => (error = errorMessage(err)));
  });

  function toggle(id: string, on: boolean) {
    selected = on ? [...selected, id].slice(0, MAX_ENTITIES) : selected.filter((x) => x !== id);
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave(withConnectionChoice({ entities: selected }, connectionId));
  }}
>
  <ConnectionSelect
    services={['homeassistant']}
    load={connections}
    bind:value={() => connectionId, (id) => ((connectionId = id), (selected = []))}
  />
  <label class="field"
    ><span>{t('smarthome.search_entities')}</span><input
      class="input"
      type="search"
      bind:value={query}
      placeholder="switch.coffee"
    /></label
  >
  <p class="muted">{t('smarthome.selected', { count: selected.length, max: MAX_ENTITIES })}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <ul class="pick">
    {#each shown as e (e.id)}
      <li>
        <label
          ><input
            type="checkbox"
            checked={selected.includes(e.id)}
            disabled={!selected.includes(e.id) && selected.length >= MAX_ENTITIES}
            onchange={(ev) => toggle(e.id, ev.currentTarget.checked)}
          />
          {e.name} <span class="muted">{e.id}</span></label
        >
      </li>
    {/each}
  </ul>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .pick {
    list-style: none;
    padding: 0;
    max-height: 18rem;
    overflow: auto;
  }
</style>
