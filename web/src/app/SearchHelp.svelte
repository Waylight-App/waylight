<script lang="ts">
  import { filterHelp, groupHelp, type HelpEntry } from '../features/search/domain/help';
  import { t } from '../shared/i18n';

  /**
   * The help of the search field (#479): what the search can do, grouped,
   * with examples that go into the field when chosen. The entries come from
   * buildHelp(), so only enabled modules and allowed commands appear.
   */
  interface Props {
    entries: HelpEntry[];
    /** Text after "?" or "hilfe", e.g. "wetter". */
    filter: string;
    /** Puts an example into the search field. */
    onpick: (example: string) => void;
  }
  let { entries, filter, onpick }: Props = $props();

  const sections = $derived(groupHelp(filterHelp(entries, filter)));
</script>

<section class="help" aria-label={t('search.help.title')}>
  <h3>{t('search.help.title')}</h3>
  {#if sections.length === 0}
    <p class="muted">{t('search.help.none', { filter })}</p>
  {:else}
    <p class="muted intro">{t('search.help.intro')}</p>
    {#each sections as s (s.group)}
      <h4>{t(`search.help.group.${s.group}`)}</h4>
      <ul>
        {#each s.entries as e (e.id)}
          <li>
            <span class="text">{e.text}</span>
            <span class="examples">
              {#each e.examples as example (example)}
                <!-- mousedown keeps the focus in the search field, like the result rows. -->
                <button
                  type="button"
                  class="example"
                  onmousedown={(ev) => ev.preventDefault()}
                  onclick={() => onpick(example)}>{example}</button
                >
              {/each}
            </span>
          </li>
        {/each}
      </ul>
    {/each}
  {/if}
</section>

<style>
  .help {
    padding: var(--space-2) var(--space-3);
    max-height: min(60vh, 32rem);
    overflow-y: auto;
  }
  h3 {
    margin: 0 0 0.25rem;
    font-size: var(--fs-sm);
  }
  h4 {
    margin: 0.75rem 0 0.25rem;
    font-size: var(--fs-xs);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  .intro {
    margin: 0;
    font-size: var(--fs-xs);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem 0.75rem;
    align-items: baseline;
    padding: 0.25rem 0;
  }
  .text {
    flex: 1 1 14rem;
  }
  .examples {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }
  .example {
    font: inherit;
    font-size: var(--fs-xs);
    padding: 0.1rem 0.5rem;
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-surface);
    color: inherit;
    cursor: pointer;
  }
  .example:hover,
  .example:focus-visible {
    border-color: var(--color-accent);
  }
</style>
