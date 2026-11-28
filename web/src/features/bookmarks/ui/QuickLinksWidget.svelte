<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Bookmark, iconSrc, initials } from '../domain/bookmark';
  import { parseQuickLinksConfig, quickLinks, SHORTCUT_LINKS } from '../domain/quickLinks';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';
  import type { QuickLinkTargets } from '../usecase/quickLinkTargets';

  /** Props of the quick links widget (#859). */
  interface Props {
    bookmarks: BookmarksUseCases;
    /** Where the links are published for the digits 1 to 9. */
    targets: QuickLinkTargets;
    config: Record<string, unknown>;
  }
  let { bookmarks, targets, config }: Props = $props();

  const cfg = $derived(parseQuickLinksConfig(config));
  let all = $state<Bookmark[] | null>(null);
  let error = $state('');

  $effect(() => {
    bookmarks
      .list()
      .then((l) => ((all = l.bookmarks), (error = '')))
      .catch((err) => (error = errorMessage(err)));
  });

  const links = $derived(all ? quickLinks(all, cfg) : []);
  // Only web links open from the keyboard, like the launcher.
  const isWeb = (url: string) => /^https?:\/\//i.test(url);

  const owner = Symbol('quick-links');
  $effect(() => targets.set(owner, links.map((b) => (isWeb(b.url) ? b.url : ''))));
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if all && links.length === 0}
  <p class="muted">{t('quicklinks.empty')}</p>
{:else}
  <ul class="tiles">
    {#each links as b, i (b.id)}
      {@const icon = iconSrc(b)}
      {@const key = i < SHORTCUT_LINKS ? String(i + 1) : ''}
      <li>
        <a
          class="tile"
          href={b.url}
          target="_blank"
          rel="noopener noreferrer"
          title={key ? t('quicklinks.key_hint', { title: b.title, key }) : b.title}
          aria-keyshortcuts={key || undefined}
        >
          <!-- The favicon covers the initials; a failed load removes it. -->
          <span class="icon" aria-hidden="true"
            >{initials(b.title)}{#if icon}<img
                src={icon}
                alt=""
                width="16"
                height="16"
                loading="lazy"
                referrerpolicy="no-referrer"
                onerror={(e) => (e.currentTarget as HTMLImageElement).remove()}
              />{/if}</span
          >
          <span class="name">{b.title}</span>
          {#if key}<kbd aria-hidden="true">{key}</kbd>{/if}
        </a>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .tiles {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(6.5rem, 1fr));
    gap: var(--space-2);
  }
  .tile {
    position: relative;
    display: grid;
    justify-items: center;
    align-content: start;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-2);
    border-radius: var(--radius-sm);
    color: inherit;
    text-decoration: none;
    text-align: center;
  }
  .tile:hover,
  .tile:focus-visible {
    background: var(--color-surface-2);
  }
  .icon {
    position: relative;
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-sm);
    background: var(--color-surface-2);
    color: var(--color-text-muted);
    font-size: var(--fs-xs);
    font-weight: 600;
  }
  .icon img {
    position: absolute;
    inset: 0;
    box-sizing: border-box;
    width: 100%;
    height: 100%;
    padding: var(--space-3);
    border-radius: inherit;
    background: var(--color-surface-2);
  }
  /* The full name wraps instead of being cut off. */
  .name {
    font-size: var(--fs-xs);
    overflow-wrap: anywhere;
  }
  kbd {
    position: absolute;
    top: var(--space-1);
    left: var(--space-1);
    font: inherit;
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  /* Small and strip tiles: icons only in a row; the name stays in the tooltip
     and the link's accessible name. */
  @container widget (max-height: 140px) or (max-width: 280px) {
    .tiles {
      display: flex;
      flex-wrap: wrap;
      gap: var(--space-1);
    }
    .tile {
      padding: var(--space-1);
    }
    .icon {
      width: 32px;
      height: 32px;
    }
    .icon img {
      padding: var(--space-2);
    }
    .name {
      position: absolute;
      width: 1px;
      height: 1px;
      overflow: hidden;
      clip-path: inset(50%);
      white-space: nowrap;
    }
    kbd {
      display: none;
    }
  }
</style>
