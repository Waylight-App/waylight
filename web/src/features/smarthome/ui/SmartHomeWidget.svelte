<script lang="ts">
  import { type ChoiceLoader, savedConnections } from '../../../shared/connection/connection';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { changedAgo, type Entity, isOn, parseWidgetConfig, stateText } from '../domain/smarthome';
  import type { SmarthomeUseCases } from '../usecase/smarthomeUseCases';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import type { FoundService } from '../../../shared/setup/setup';
  import { smarthomeSetup } from './smarthomeSetup';

  /** Props of the smart home widget (#49). */
  interface Props {
    smarthome: SmarthomeUseCases;
    config: Record<string, unknown>;
    /** Services found in the home network for the setup (#590). */
    discover?: () => Promise<FoundService[]>;
    /** Saved connections to offer in the setup, and storing the choice (#709). */
    connections?: ChoiceLoader;
    onsave?: (config: Record<string, unknown>) => Promise<void>;
  }
  let { smarthome, config, discover, connections, onsave }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 30_000;
  const ids = $derived(parseWidgetConfig(config).entities);
  const connection = $derived(parseWidgetConfig(config).connectionId);
  let entities = $state<Entity[] | null>(null);
  let error = $state('');
  let unconfigured = $state(false);
  let busy = $state('');
  let now = $state(new Date());

  const load = () =>
    smarthome
      .states(ids, connection)
      .then((list) => ((entities = list), (error = ''), (unconfigured = false), (now = new Date())))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });

  $effect(() => {
    void ids;
    void connection;
    return poll(load, POLL_MS);
  });
  // svelte-ignore state_referenced_locally
  const setup = smarthomeSetup(smarthome, discover);

  async function act(e: Entity) {
    const question = e.action === 'press' ? 'smarthome.confirm_press' : 'smarthome.confirm_toggle';
    if (!confirm(t(question, { name: e.name }))) return;
    busy = e.id;
    try {
      await smarthome.act(e.id, connection);
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }
</script>

{#if unconfigured}
  <SetupInPlace
    text={t('smarthome.connect_under_settings')}
    descriptor={setup}
    instance
    connections={savedConnections(['homeassistant'], connections, config, onsave)}
    ondone={() => void load().catch(() => undefined)}
  />
{:else if ids.length === 0}
  <p class="muted">{t('smarthome.choose_entities')}</p>
{:else if error && !entities}
  <p class="error">{error}</p>
{:else if entities}
  <ul class="tiles">
    {#each entities as e (e.id)}
      <li class:on={isOn(e)}>
        <span class="name" title={e.id}>{e.name}</span>
        <strong class="state">{stateText(e)}</strong>
        <span class="muted small">{changedAgo(e.lastChanged, now)}</span>
        {#if e.action}
          <button
            data-action
            class="btn"
            type="button"
            disabled={busy !== ''}
            aria-pressed={e.action === 'toggle' ? isOn(e) : undefined}
            onclick={() => act(e)}
            >{busy === e.id
              ? '…'
              : e.action === 'press'
                ? t('smarthome.press')
                : isOn(e)
                  ? t('smarthome.turn_off')
                  : t('smarthome.turn_on')}</button
          >
        {/if}
      </li>
    {/each}
  </ul>
  {#if error}<p class="error small">{error}</p>{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .tiles {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
    gap: var(--space-2);
  }
  .tiles li {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-2);
    border: 1px solid transparent;
    background: var(--color-surface-2);
    border-radius: var(--radius, 8px);
  }
  .tiles li.on {
    border-color: var(--color-accent);
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .state {
    font-size: 1.2em;
  }
  .small {
    font-size: 0.8em;
  }
</style>
