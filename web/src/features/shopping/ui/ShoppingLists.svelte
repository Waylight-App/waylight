<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type ShareGroup, sharedVia } from '../../../shared/sharing/sharing';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';
  import { exportUrl, type ShoppingList } from '../domain/shopping';
  import type { ShoppingUseCases } from '../usecase/shoppingUseCases';

  /** Props of the management of the lists kept on waylight (#471). */
  interface Props {
    shopping: ShoppingUseCases;
  }
  let { shopping }: Props = $props();

  let lists = $state<ShoppingList[]>([]);
  let caldav = $state(false);
  let groups = $state<ShareGroup[]>([]);
  // Name drafts per list, so typing in one row does not touch the others.
  let names = $state<Record<string, string>>({});
  let newName = $state('');
  let error = $state('');
  let message = $state('');

  async function reload() {
    const o = await shopping.overview();
    lists = o.lists;
    caldav = o.caldav;
    names = Object.fromEntries(o.lists.map((l) => [l.id, l.name]));
  }

  /** Runs one change, then shows the lists as the server has them now. */
  async function run(action: () => Promise<unknown>) {
    error = message = '';
    try {
      await action();
      await reload();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  run(async () => {});
  $effect(() => {
    shopping
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });

  function create(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      if (await shopping.createList(newName)) newName = '';
    });
  }

  function remove(l: ShoppingList) {
    if (confirm(t('shopping.confirm_delete', { name: l.name }))) run(() => shopping.deleteList(l.id));
  }

  function importCaldav(l: ShoppingList) {
    run(async () => {
      const count = await shopping.importCaldav(l.id);
      message = t('shopping.imported', { count });
    });
  }
</script>

<h3>{t('shopping.lists_title')}</h3>
<p class="muted">{t('shopping.lists_hint')}</p>
<ul class="list">
  {#each lists as l (l.id)}
    <li>
      {#if l.mine}
        <form
          class="row"
          onsubmit={(e) => {
            e.preventDefault();
            run(() => shopping.updateList(l.id, names[l.id] ?? l.name, l.visibility));
          }}
        >
          <label class="grow"
            ><span class="sr-only">{t('shopping.list_name')}</span><input
              class="input"
              bind:value={names[l.id]}
              maxlength="100"
              required
            /></label
          >
          <button class="btn" type="submit" disabled={(names[l.id] ?? '').trim() === l.name}
            >{t('shopping.save_name')}</button
          >
        </form>
        <VisibilitySelect
          value={l.visibility}
          {groups}
          onchange={(v) => run(() => shopping.updateList(l.id, l.name, v))}
        />
      {:else}
        {@const via = sharedVia(l.visibility, groups)}
        <strong>{l.name}</strong>
        <span class="muted small"
          >{via === null
            ? t('shopping.shared_with_you')
            : 'group' in via
              ? t('shopping.shared_with_you_via', { group: via.group })
              : t('shopping.shared_with_everyone')}</span
        >
      {/if}
      <p class="actions small">
        <span class="muted">{t('shopping.open_count', { count: l.open })}</span>
        <span
          >{t('shopping.export')}:
          <a href={exportUrl(l.id, 'text')} download>{t('shopping.export_text')}</a>,
          <a href={exportUrl(l.id, 'ics')} download>{t('shopping.export_ics')}</a></span
        >
        {#if caldav}
          <button class="btn link" type="button" onclick={() => importCaldav(l)}
            >{t('shopping.import_caldav')}</button
          >
        {/if}
        {#if l.mine}
          <button class="btn link danger" type="button" onclick={() => remove(l)}
            >{t('shopping.delete_list')}</button
          >
        {/if}
      </p>
    </li>
  {/each}
</ul>
<form class="row" onsubmit={create}>
  <label class="grow"
    ><span class="sr-only">{t('shopping.new_list_name')}</span><input
      class="input"
      bind:value={newName}
      placeholder={t('shopping.new_list_name')}
      maxlength="100"
    /></label
  >
  <button class="btn btn-primary" type="submit" disabled={!newName.trim()}>{t('shopping.create_list')}</button
  >
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  .list {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
  }
  .list li {
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border);
  }
  .row {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  .grow {
    flex: 1;
    min-width: 0;
  }
  .grow .input {
    width: 100%;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
    margin: var(--space-1) 0 0;
  }
  .small {
    font-size: 0.85em;
  }
</style>
