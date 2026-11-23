<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { KeywordKind, KeywordTarget, OwnKeyword } from '../domain/keywords';

  /** Props of the list of own keywords (#478) in the search settings. */
  interface Props {
    keywords: OwnKeyword[];
    /** Tiles, bookmarks and widgets to pick from; without it only addresses and searches. */
    targets?: KeywordTarget[];
    onadd: (k: OwnKeyword) => Promise<boolean>;
    ondelete: (keyword: string) => void;
  }
  let { keywords, targets = [], onadd, ondelete }: Props = $props();

  const kindLabel: Record<
    KeywordKind,
    'search.kw_service' | 'search.kw_bookmark' | 'search.kw_widget' | 'search.kw_url' | 'search.kw_query'
  > = {
    service: 'search.kw_service',
    bookmark: 'search.kw_bookmark',
    widget: 'search.kw_widget',
    url: 'search.kw_url',
    query: 'search.kw_query',
  };
  // Only kinds that have something to pick, plus the two typed ones.
  const kinds = $derived(
    (['service', 'bookmark', 'widget', 'url', 'query'] as const).filter(
      (k) => k === 'url' || k === 'query' || targets.some((x) => x.kind === k),
    ),
  );

  let word = $state('');
  let kind = $state<KeywordKind>('url');
  let picked = $state('');
  let text = $state('');
  const choices = $derived(targets.filter((x) => x.kind === kind));
  const typed = $derived(kind === 'url' || kind === 'query');

  /** What the keyword shows next to it; for a typed target, the target itself. */
  function describe(k: OwnKeyword): string {
    return k.label || k.target;
  }

  async function add(e: SubmitEvent) {
    e.preventDefault();
    const choice = choices.find((x) => x.target === picked);
    const k: OwnKeyword = typed
      ? { keyword: word, kind, target: text, label: '' }
      : { keyword: word, kind, target: picked, label: choice?.label ?? '' };
    if (await onadd(k)) word = picked = text = '';
  }
</script>

<h4>{t('search.own_keywords')}</h4>
<p class="muted">{t('search.own_keywords_hint')}</p>
{#if keywords.length}
  <ul class="keywords">
    {#each keywords as k (k.keyword)}
      <li>
        <code>{k.keyword}</code>
        <span>→ {describe(k)}</span>
        <span class="muted small">{t(kindLabel[k.kind])}</span>
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('search.delete_keyword', { keyword: k.keyword })}
          onclick={() => ondelete(k.keyword)}>✕</button
        >
      </li>
    {/each}
  </ul>
{/if}
<form class="add" onsubmit={add}>
  <input
    class="input"
    bind:value={word}
    placeholder="nas"
    aria-label={t('search.own_keyword')}
    maxlength="24"
    required
  />
  <select class="input" bind:value={kind} aria-label={t('search.keyword_target_kind')}>
    {#each kinds as k (k)}<option value={k}>{t(kindLabel[k])}</option>{/each}
  </select>
  {#if typed}
    <input
      class="input"
      bind:value={text}
      type={kind === 'url' ? 'url' : 'text'}
      placeholder={kind === 'url' ? 'https://nas.local:5001' : t('search.keyword_query_placeholder')}
      aria-label={t('search.keyword_target')}
      required
    />
  {:else}
    <select class="input" bind:value={picked} aria-label={t('search.keyword_target')} required>
      <option value="" disabled>{t('search.keyword_pick')}</option>
      {#each choices as c (c.target)}<option value={c.target}>{c.label}</option>{/each}
    </select>
  {/if}
  <button class="btn" type="submit">{t('search.add')}</button>
</form>

<style>
  h4 {
    margin: var(--space-4) 0 0;
  }
  .keywords {
    list-style: none;
    padding: 0;
  }
  .keywords li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .small {
    font-size: var(--fs-xs);
  }
  .add {
    display: grid;
    grid-template-columns: 6rem 9rem 1fr auto;
    gap: var(--space-2);
  }
  @media (max-width: 600px) {
    .add {
      grid-template-columns: 1fr;
    }
  }
</style>
