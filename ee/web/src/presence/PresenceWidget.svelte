<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { nextStatus, parseConfig, poll, PresenceApi, settable, type Board, type Status } from './presence';
  import { texts } from './texts';

  /** The presence widget (#346): who is in the office, at home or away today. */
  interface Props {
    context: PluginContext;
    config?: Record<string, unknown>;
  }
  let { context, config }: Props = $props();

  const api = $derived(new PresenceApi(context.http));
  const group = $derived(parseConfig(config).group);
  const tr = (key: string, params: Record<string, string | number> = {}) => text(texts, key, context.language, params);
  const date = (day: string) =>
    new Date(`${day}T12:00:00`).toLocaleDateString(context.language, { weekday: 'short', day: 'numeric', month: 'numeric' });
  const label = (s: Status) => tr(s || 'none');

  const readOnly = $derived(context.level('presence') === 'read');

  let board = $state<Board | null>(null);
  let disabled = $state(false);
  let error = $state('');
  let returnOn = $state('');
  let busy = $state(false);

  async function load() {
    try {
      board = await api.board(group);
      disabled = false;
      error = '';
    } catch (err) {
      if ((err as { status?: number })?.status === 404) {
        disabled = true;
        board = null;
      } else {
        error = context.errorMessage(err);
      }
    }
  }

  $effect(() => {
    void group;
    return poll(() => void load());
  });

  async function press(s: Status) {
    const next = nextStatus(board?.mine ?? null, s);
    if (next === null || busy) return;
    busy = true;
    try {
      await api.set(next, next === 'away' ? returnOn : '');
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if disabled}
  <p class="muted">{tr('disabled')}</p>
{:else if board}
  {#if board.mine}
    <div class="mine" role="group" aria-label={tr('my_status')}>
      {#each settable as s (s)}
        <button
          type="button"
          class="btn"
          class:btn-primary={board.mine.status === s}
          aria-pressed={board.mine.status === s}
          disabled={readOnly || busy || board.mine.source === 'absence'}
          onclick={() => press(s)}>{label(s)}</button
        >
      {/each}
    </div>
    {#if board.mine.source === 'absence'}
      <p class="muted small">{label(board.mine.status)} · {tr('from_absence')}</p>
    {:else if !readOnly}
      <label class="small return"
        >{tr('return_on')} <input class="input" type="date" min={board.day} bind:value={returnOn} /></label
      >
    {/if}
  {/if}
  {#if readOnly}<p class="muted small">{tr('read_only')}</p>{/if}
  {#if board.people.length === 0}
    <p class="muted">{tr('empty')}</p>
  {:else}
    <ul>
      {#each board.people as p (p.accountId)}
        <li>
          <span class="dot {p.status || 'none'}" aria-hidden="true"></span>
          <span class="name">{p.name}</span>
          <span class="muted small">
            {label(p.status)}{#if p.returnOn}, {tr('back', { date: date(p.returnOn) })}{/if}
          </span>
        </li>
      {/each}
    </ul>
  {/if}
  <p class="muted small counts">{tr('counts', board.counts)}</p>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .mine {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .return {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-block: var(--space-2);
  }
  .small {
    font-size: 0.85em;
  }
  ul {
    list-style: none;
    margin: var(--space-2) 0 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1, 4px) 0;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .dot {
    width: 0.6em;
    height: 0.6em;
    border-radius: 50%;
    flex: none;
    background: var(--color-border, rgba(127, 127, 127, 0.4));
  }
  .dot.office {
    background: var(--color-success, #2e7d32);
  }
  .dot.home {
    background: var(--color-accent);
  }
  .dot.away {
    background: var(--color-warning, #b26a00);
  }
  .counts {
    margin-top: var(--space-2);
  }
</style>
