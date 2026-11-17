<script lang="ts">
  import { t } from '../../../../shared/i18n';
  import Meter from '../../../../shared/ui/Meter.svelte';
  import { copyToClipboard } from '../../../../shared/ui/clipboard';
  import { type ActionItem, type Answer, dataAge } from '../../domain/smart/answer';

  /** Props of one answer of the smart search (#156); generic per answer kind. */
  interface Props {
    answer: Answer;
    /** Opens the widget with the details, when the dashboard has one. */
    onopen?: (widgetKind: string) => void;
    /** Runs a confirmed action. */
    onexecute?: (item: ActionItem) => void;
    /** Puts a suggestion into the search field (#158). */
    onpick?: (text: string) => void;
  }
  let { answer: given, onopen, onexecute, onpick }: Props = $props();

  // A regenerated answer (#477) replaces the given one until the search
  // hands in a different answer, e.g. after the next keystroke.
  let renewed = $state<{ from: Answer; answer: Answer } | null>(null);
  const answer = $derived(renewed?.from === given ? renewed.answer : given);

  const age = $derived(dataAge(answer.at, new Date()));

  // Calculator results can be copied (#160); Enter in the search does the same.
  let copied = $state(false);
  async function copy() {
    if (answer.copy === undefined) return;
    copied = await copyToClipboard(answer.copy);
    if (copied) setTimeout(() => (copied = false), 1500);
  }

  async function regenerate() {
    const from = given;
    const next = await answer.regenerate?.();
    if (next) renewed = { from, answer: next };
    copied = false;
  }
</script>

<article class="answer" data-kind={answer.kind} aria-label={answer.title}>
  <header>
    <h3>{answer.title}</h3>
    {#if age}<span class="age">{t(age.key, { n: age.n })}</span>{/if}
  </header>

  {#if answer.kind === 'metric'}
    <p class="value" data-severity={answer.severity}>
      {#if answer.swatch}<span class="swatch" style:background={answer.swatch} aria-hidden="true"
        ></span>{/if}<span>{answer.value}</span>
    </p>
    {#if answer.percent !== undefined}<Meter percent={answer.percent} label={answer.title} />{/if}
    {#if answer.detail}<p class="detail">{answer.detail}</p>{/if}
  {:else if answer.kind === 'status'}
    <p class="status"><span class="dot" data-state={answer.state} aria-hidden="true"></span>{answer.text}</p>
    {#if answer.detail}<p class="detail">{answer.detail}</p>{/if}
  {:else if answer.kind === 'list'}
    {#if answer.items.length === 0}
      <p class="detail">{answer.empty}</p>
    {:else}
      <ul>
        {#each answer.items as item, i (i)}
          <li>
            {#if item.state}<span class="dot" data-state={item.state} aria-hidden="true"></span>{/if}
            <span class="label">{item.label}</span>
            <span class="detail">{item.value}</span>
          </li>
        {/each}
      </ul>
    {/if}
  {:else if answer.kind === 'text'}
    <p>{answer.text}</p>
    {#if answer.suggestions?.length}
      <p class="chips">
        {#each answer.suggestions as s (s)}
          <button type="button" class="btn chip" onclick={() => onpick?.(s)}>{s}</button>
        {/each}
      </p>
    {/if}
  {:else if answer.kind === 'actions'}
    <p class="detail">{answer.text}</p>
    <p class="chips">
      {#each answer.actions as a (a.action)}
        <button type="button" class="btn btn-primary" onclick={() => confirm(a.confirm) && onexecute?.(a)}
          >{a.label}</button
        >
      {/each}
    </p>
  {/if}

  {#if answer.copy !== undefined}
    <p class="chips">
      <button type="button" class="btn chip" onclick={copy}
        >{t(copied ? 'smart.copied' : 'smart.copy')}</button
      >
      {#if answer.regenerate}
        <button type="button" class="btn chip" onclick={regenerate}>{t('tools.new')}</button>
      {/if}
      <span class="detail">{t('smart.enter_copies')}</span>
    </p>
  {/if}

  {#if answer.widgetKind && onopen}
    <button type="button" class="btn btn-icon open" onclick={() => onopen?.(answer.widgetKind ?? '')}
      >{t('smart.open_widget')}</button
    >
  {/if}
</article>

<style>
  .answer {
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    background: var(--color-surface);
    display: grid;
    gap: var(--space-1);
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: var(--space-2);
  }
  h3 {
    margin: 0;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--color-text-muted);
  }
  p {
    margin: 0;
  }
  .age,
  .detail {
    font-size: 0.8rem;
    color: var(--color-text-muted);
  }
  /*
   * The answer itself reads as the result (#511): large, in the display
   * face and the accent text colour; warnings keep their status colour.
   */
  .value {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-display);
    font-size: 1.6rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
    color: var(--color-accent-text);
  }
  .swatch {
    display: inline-block;
    width: 1em;
    height: 1em;
    vertical-align: -0.1em;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .value[data-severity='warning'] {
    color: var(--color-warning);
  }
  .value[data-severity='critical'] {
    color: var(--color-danger);
  }
  .status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-weight: 500;
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex: none;
    background: var(--color-text-muted);
  }
  .dot[data-state='ok'] {
    background: var(--color-success, #2e9d55);
  }
  .dot[data-state='warning'] {
    background: var(--color-warning);
  }
  .dot[data-state='error'] {
    background: var(--color-danger);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 2px;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .chip {
    padding: 2px var(--space-2);
  }
  .open {
    justify-self: start;
    font-size: 0.8rem;
    padding: 0;
    color: var(--color-accent-text);
  }
</style>
