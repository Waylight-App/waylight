<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import TileLayout from '../../../shared/ui/TileLayout.svelte';
  import { shoppingSetup } from './shoppingSetup';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    CALDAV,
    resolveCurrent,
    type ShoppingItem,
    type ShoppingOverview,
    withDone,
  } from '../domain/shopping';
  import type { ShoppingUseCases } from '../usecase/shoppingUseCases';
  import { offerUndo, undoQueue } from '../../../shared/ui/undo.svelte';
  import { SvelteSet } from 'svelte/reactivity';

  /** Props of the shopping list widget. */
  interface Props {
    shopping: ShoppingUseCases;
  }
  let { shopping }: Props = $props();
  const poll = usePolling();
  // svelte-ignore state_referenced_locally
  const setup = shoppingSetup(shopping);

  const POLL_MS = 60_000;
  let overview = $state<ShoppingOverview>({ lists: [], caldav: false, current: '' });
  let current = $state('');
  let items = $state<ShoppingItem[]>([]);
  let loaded = $state(false);
  let draft = $state('');
  let error = $state('');
  let busy = $state(false);
  const source = $derived(current ? shopping.source(current) : null);
  // Ticked off and removed items wait for their undo toast (#867); the list
  // shows the change at once.
  const ticked = new SvelteSet<string>();
  const removed = new SvelteSet<string>();
  const shown = $derived(
    items.filter((i) => !removed.has(i.id)).map((i) => (ticked.has(i.id) ? { ...i, done: true } : i)),
  );
  const doneCount = $derived(shown.filter((i) => i.done).length);
  // The choice only appears when there is something to choose between.
  const choices = $derived(overview.lists.length + (overview.caldav ? 1 : 0));

  // One poll reads both the lists and the shown items, so a list someone
  // shares with this account shows up without reloading the page.
  async function load() {
    try {
      overview = await shopping.overview();
      current = resolveCurrent(overview);
      items = current ? await shopping.source(current).items() : [];
      loaded = true;
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => {
    return poll(load, POLL_MS);
  });

  async function pick(value: string) {
    current = value;
    items = [];
    try {
      await shopping.setCurrent(value);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function createList() {
    busy = true;
    try {
      const list = await shopping.createList(t('shopping.default_list_name'));
      if (list) await pick(list.id);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function add(e: SubmitEvent) {
    e.preventDefault();
    if (!source) return;
    busy = true;
    const text = draft;
    try {
      const it = await source.add(text);
      if (it) items = [...items.filter((i) => !i.done), it, ...items.filter((i) => i.done)];
      // Cleared only once the server confirmed the item, and only if nothing
      // new was typed meanwhile: a quick second entry must not be lost.
      if (draft === text) draft = '';
      error = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function toggle(it: ShoppingItem) {
    const src = source;
    if (!src) return;
    if (!it.done) {
      offerUndo(ticked, it.id, t('shopping.ticked', { name: it.summary }), () =>
        send(async () => {
          await src.setDone(it.id, true);
          items = withDone(items, it.id, true);
        }),
      );
      return;
    }
    const before = items;
    items = withDone(items, it.id, !it.done);
    try {
      await source.setDone(it.id, !it.done);
      error = '';
    } catch (err) {
      items = before;
      error = errorMessage(err);
    }
  }

  function remove(it: ShoppingItem) {
    const src = source;
    if (!src) return;
    offerUndo(removed, it.id, t('shopping.removed', { name: it.summary }), () =>
      send(async () => {
        await src.remove(it.id);
        items = items.filter((i) => i.id !== it.id);
      }),
    );
  }

  /** Sends a quick action once its undo time is over; a failure is shown and brings the item back. */
  async function send(action: () => Promise<void>) {
    try {
      await action();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  async function clear() {
    if (!source) return;
    try {
      // A tick-off still waiting for its undo toast is part of "done".
      await undoQueue.flush();
      await source.clearCompleted();
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if loaded && !current}
  <SetupInPlace
    text={t('shopping.no_list')}
    descriptor={setup}
    instance={false}
    label={t('shopping.connect_list')}
    ondone={() => void load().catch(() => undefined)}
  >
    <button class="btn btn-primary" type="button" disabled={busy} onclick={() => void createList()}
      >{t('shopping.create_list')}</button
    >
  </SetupInPlace>
{:else}
  <TileLayout>
    {#if choices > 1}
      <label class="pick"
        ><span class="sr-only">{t('shopping.list')}</span>
        <select class="input" value={current} onchange={(e) => pick(e.currentTarget.value)}>
          {#each overview.lists as l (l.id)}
            <option value={l.id}>{l.name}</option>
          {/each}
          {#if overview.caldav}<option value={CALDAV}>{t('shopping.caldav_list')}</option>{/if}
        </select></label
      >
    {/if}
    <ul class="items">
      {#each shown as it (it.id)}
        <ListRow>
          <label class="item" class:done={it.done}
            ><input type="checkbox" checked={it.done} onchange={() => toggle(it)} />
            <span>{it.summary}</span></label
          >
          {#snippet trail()}
            <button
              class="btn link remove"
              type="button"
              aria-label={t('shopping.remove_item', { name: it.summary })}
              onclick={() => remove(it)}>×</button
            >
          {/snippet}
        </ListRow>
      {:else}
        {#if loaded}<li class="muted">{t('shopping.the_list_is_empty')}</li>{/if}
      {/each}
    </ul>
    {#if doneCount}
      <button class="btn link" type="button" onclick={clear}
        >{t('shopping.remove_done', { doneCount: doneCount })}</button
      >
    {/if}
    {#snippet footer()}
      {#if error}<p class="error small" role="alert">{error}</p>{/if}
      <form class="add" onsubmit={add}>
        <input
          class="input"
          bind:value={draft}
          placeholder={t('shopping.add_item')}
          aria-label={t('shopping.new_item')}
          maxlength="200"
          disabled={!loaded}
        />
        <button class="btn btn-primary" type="submit" disabled={busy || !draft.trim()}
          >{t('shopping.add')}</button
        >
      </form>
    {/snippet}
  </TileLayout>
{/if}

<style>
  .pick {
    display: block;
    margin-bottom: var(--space-1);
  }
  .pick select {
    width: 100%;
  }
  .add {
    display: flex;
    gap: var(--space-1);
  }
  .add .input {
    flex: 1;
    min-width: 0;
  }
  .items {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .item {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    cursor: pointer;
  }
  .done span {
    text-decoration: line-through;
    color: var(--color-text-subtle);
  }
  .remove {
    margin-top: 0;
    padding: 0 var(--space-1);
    opacity: 0.6;
  }
  .remove:hover,
  .remove:focus-visible {
    opacity: 1;
  }
  .link {
    margin-top: var(--space-1);
    font-size: var(--fs-xs);
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
