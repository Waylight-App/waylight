<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Feed, groupByCategory, type ImportReport } from '../domain/feed';
  import {
    catalogueLanguages,
    type CatalogueEntry,
    sameFeedUrl,
    searchCatalogue,
    validEntry,
  } from '../domain/catalogue';
  import { intlLocale } from '../../../shared/i18n';
  import type { RssUseCases } from '../usecase/rssUseCases';

  /** Props of the feed management section. */
  interface Props {
    rss: RssUseCases;
  }
  let { rss }: Props = $props();

  let feeds = $state<Feed[]>([]);
  let url = $state('');
  let error = $state('');
  let busy = $state(false);

  async function run(action: () => Promise<Feed[] | unknown>) {
    error = '';
    busy = true;
    try {
      const result = await action();
      feeds = Array.isArray(result) ? (result as Feed[]) : await rss.feeds();
      return true;
    } catch (err) {
      error = errorMessage(err);
      return false;
    } finally {
      busy = false;
    }
  }

  run(() => rss.feeds());

  // OPML import (#97): the report says what was added, skipped or refused.
  let report = $state<ImportReport | null>(null);
  let picker = $state<HTMLInputElement>();
  const groups = $derived(groupByCategory(feeds));

  // Built-in catalogue (#289): loaded on demand from the app itself, no
  // outside request until a source is subscribed.
  let catalogue = $state<CatalogueEntry[] | null>(null);
  let query = $state('');
  let language = $state('');
  const found = $derived(catalogue ? searchCatalogue(catalogue, query, language) : []);
  const subscribed = (url: string) => feeds.some((f) => sameFeedUrl(f.url, url));
  async function openCatalogue() {
    if (catalogue) return (catalogue = null);
    const raw = (await import('../catalogue/feeds.json')).default as unknown[];
    catalogue = raw.filter(validEntry);
    const ui = intlLocale().slice(0, 2);
    language = catalogueLanguages(catalogue).includes(ui) ? ui : '';
  }
  const languageName = (code: string) =>
    new Intl.DisplayNames([intlLocale()], { type: 'language' }).of(code) ?? code;

  async function importFile() {
    const file = picker?.files?.[0];
    if (!file) return;
    report = null;
    await run(async () => {
      const r = await rss.importOpml(file);
      report = r.report;
      return r.feeds;
    });
    if (picker) picker.value = '';
  }
</script>

{#each groups as g (g.category)}
  {#if g.category}<h4 class="category">{g.category}</h4>{/if}
  <ul class="list">
    {#each g.feeds as f (f.id)}
      <li>
        <span class="name">
          <strong>{f.title}</strong>
          {#if f.unread}<span class="badge">{f.unread}</span>{/if}
          <br /><span class="muted small">{f.url}</span>
          {#if f.lastError}<br /><span class="error small">{f.lastError}</span>{/if}
        </span>
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('rss.unsubscribe', { title: f.title })}
          onclick={() =>
            confirm(t('rss.confirm_unsubscribe', { name: f.title })) && run(() => rss.unsubscribe(f.id))}
          >✕</button
        >
      </li>
    {/each}
  </ul>
{/each}
<form
  class="add"
  onsubmit={async (e) => {
    e.preventDefault();
    if (await run(() => rss.subscribe(url))) url = '';
  }}
>
  <input
    class="input"
    type="url"
    bind:value={url}
    placeholder="https://example.com/feed.xml"
    required
    aria-label={t('rss.feed_url')}
  />
  <button class="btn" type="submit" disabled={busy}>{t('rss.subscribe')}</button>
</form>
<p>
  <button class="btn" type="button" aria-expanded={catalogue !== null} onclick={openCatalogue}
    >{t('rss.catalogue_browse')}</button
  >
</p>
{#if catalogue}
  <section class="catalogue" aria-label={t('rss.catalogue')}>
    <div class="filters">
      <input
        class="input"
        type="search"
        bind:value={query}
        placeholder={t('rss.catalogue_search')}
        aria-label={t('rss.catalogue_search')}
      />
      <select class="input" bind:value={language} aria-label={t('rss.catalogue_language')}>
        <option value="">{t('rss.catalogue_all_languages')}</option>
        {#each catalogueLanguages(catalogue) as l (l)}<option value={l}>{languageName(l)}</option>{/each}
      </select>
    </div>
    {#each found as g (g.topic)}
      <h4 class="category">{t(`rss.topic_${g.topic}`)}</h4>
      <ul class="list">
        {#each g.entries as e (e.url)}
          <li>
            <span class="name"><strong>{e.title}</strong><br /><span class="muted small">{e.url}</span></span>
            {#if subscribed(e.url)}
              <span class="muted small">{t('rss.catalogue_subscribed')}</span>
            {:else}
              <button
                class="btn"
                type="button"
                disabled={busy}
                aria-label={t('rss.catalogue_add', { title: e.title })}
                onclick={() => run(() => rss.subscribe(e.url))}>{t('rss.subscribe')}</button
              >
            {/if}
          </li>
        {/each}
      </ul>
    {:else}
      <p class="muted">{t('rss.catalogue_none')}</p>
    {/each}
  </section>
{/if}
<p>
  <button class="btn btn-icon" type="button" disabled={busy} onclick={() => run(() => rss.refresh())}
    >{t('rss.refresh_all_feeds_now')}</button
  >
</p>
<div class="opml">
  <label class="btn"
    >{t('rss.opml_import')}<input
      bind:this={picker}
      class="file"
      type="file"
      accept=".opml,.xml,text/x-opml,text/xml,application/xml"
      disabled={busy}
      onchange={importFile}
    /></label
  >
  <a class="btn" href={rss.opmlUrl()} download>{t('rss.opml_export')}</a>
</div>
{#if report}
  <div class="report" role="status">
    <p>
      {t('rss.opml_report', {
        added: report.added,
        duplicates: report.duplicates,
        invalid: report.invalid.length,
      })}
    </p>
    {#if report.invalid.length}
      <ul>
        {#each report.invalid as i, n (n)}<li class="muted small">{i.title || i.url}: {i.reason}</li>{/each}
      </ul>
    {/if}
  </div>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .catalogue {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    margin-bottom: var(--space-3);
  }
  .filters {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .filters input {
    flex: 1 1 12rem;
  }
  .filters select {
    flex: 0 1 10rem;
  }
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .list li {
    display: flex;
    align-items: flex-start;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .small {
    font-size: var(--fs-xs);
  }
  .badge {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    border-radius: 999px;
    padding: 0 6px;
    font-size: var(--fs-xs);
  }
  .add {
    display: flex;
    gap: var(--space-2);
  }
  .category {
    margin: var(--space-3) 0 0;
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
  .opml {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin-top: var(--space-2);
  }
  .opml label:has(:focus-visible) {
    outline: 2px solid var(--color-accent-text);
  }
  .file {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .report ul {
    margin: var(--space-1) 0 0;
    padding-left: var(--space-3);
  }
</style>
