<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { TOKEN_HINT, WIKI_KIND_NAMES, WIKI_KINDS, type Wiki, type WikiInput } from '../domain/wiki';
  import type { WikiUseCases } from '../usecase/wikiUseCases';

  /** Settings → Wiki search (#321): connect BookStack, Outline or Wiki.js. */
  interface Props {
    wiki: WikiUseCases;
  }
  let { wiki }: Props = $props();

  let wikis = $state<Wiki[]>([]);
  let editing = $state<{ id: string | null; form: WikiInput; hasToken: boolean } | null>(null);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    wiki
      .list()
      .then((l) => (wikis = l))
      .catch((err) => (error = errorMessage(err)));
  });

  const empty = (): WikiInput => ({ name: '', kind: 'bookstack', url: '', token: '' });

  async function act(action: () => Promise<void>, done: string) {
    error = message = '';
    busy = true;
    try {
      await action();
      message = done;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function save(e: SubmitEvent) {
    e.preventDefault();
    const ed = editing;
    if (!ed) return;
    void act(async () => {
      const saved = await wiki.save(ed.id, ed.form);
      wikis = ed.id ? wikis.map((w) => (w.id === saved.id ? saved : w)) : [...wikis, saved];
      editing = null;
    }, t('wiki.saved'));
  }

  function remove(w: Wiki) {
    if (!confirm(t('common.confirm_delete', { name: w.name }))) return;
    void act(async () => {
      await wiki.remove(w.id);
      wikis = wikis.filter((x) => x.id !== w.id);
    }, '');
  }
</script>

<p class="muted">{t('wiki.intro')}</p>
<p class="notice" role="note">{t('wiki.shared_token')}</p>

<ul class="list">
  {#each wikis as w (w.id)}
    <li>
      <span><strong>{w.name}</strong> <span class="muted">({WIKI_KIND_NAMES[w.kind]}, {w.url})</span></span>
      <span class="actions">
        <button
          class="btn"
          type="button"
          disabled={busy}
          onclick={() => act(() => wiki.test(w.id), t('wiki.test_ok'))}>{t('wiki.test')}</button
        >
        <button
          class="btn"
          type="button"
          onclick={() =>
            (editing = {
              id: w.id,
              form: { name: w.name, kind: w.kind, url: w.url, token: '' },
              hasToken: w.hasToken,
            })}>{t('common.edit')}</button
        >
        <button class="btn btn-danger" type="button" onclick={() => remove(w)}>{t('wiki.delete')}</button>
      </span>
    </li>
  {:else}
    <li class="muted">{t('wiki.none')}</li>
  {/each}
</ul>

{#if editing}
  <form class="editor" onsubmit={save}>
    <label class="field"
      ><span>{t('wiki.name')}</span><input
        class="input"
        bind:value={editing.form.name}
        required
        maxlength="60"
      /></label
    >
    <label class="field"
      ><span>{t('wiki.kind')}</span><select class="input" bind:value={editing.form.kind}>
        {#each WIKI_KINDS as k (k)}<option value={k}>{WIKI_KIND_NAMES[k]}</option>{/each}
      </select></label
    >
    <label class="field"
      ><span>{t('wiki.url')}</span><input
        class="input"
        type="url"
        bind:value={editing.form.url}
        required
        placeholder="https://wiki.example.org"
      /></label
    >
    <label class="field"
      ><span>{t('wiki.token')}</span><input
        class="input"
        type="password"
        autocomplete="off"
        bind:value={editing.form.token}
        required={!editing.hasToken}
        placeholder={editing.hasToken ? t('wiki.token_stored') : ''}
      /></label
    >
    <p class="muted small">{t(TOKEN_HINT[editing.form.kind])}</p>
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
      <button class="btn" type="button" onclick={() => (editing = null)}>{t('common.cancel')}</button>
    </p>
  </form>
{:else}
  <p>
    <button class="btn" type="button" onclick={() => (editing = { id: null, form: empty(), hasToken: false })}
      >+ {t('wiki.add')}</button
    >
  </p>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  .notice {
    border-left: 3px solid var(--color-accent);
    padding-left: var(--space-2);
  }
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .list li {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: var(--space-2);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .editor {
    display: grid;
    gap: var(--space-2);
  }
  .small {
    font-size: 0.85em;
    margin: 0;
  }
</style>
