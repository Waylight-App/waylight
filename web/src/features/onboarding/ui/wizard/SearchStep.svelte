<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '../../../../shared/i18n';
  import AnswerCard from '../../../search/ui/answers/AnswerCard.svelte';
  import type { Answer } from '../../../search/domain/smart/answer';
  import type { StepProps } from './wizardContext';

  let { ctx, session }: StepProps = $props();

  const TYPING_MS = 70;
  const TYPED_ANSWER_MS = 300;

  let examples = $state<string[]>([]);
  let query = $state('');
  /** The example that was played last; its chip is highlighted. */
  let played = $state('');
  let answers = $state<Answer[]>([]);
  let asked = $state(false);
  let typing: ReturnType<typeof setInterval> | undefined;
  let pause: ReturnType<typeof setTimeout> | undefined;
  let request: AbortController | undefined;

  const reduceMotion = () => matchMedia('(prefers-reduced-motion: reduce)').matches;

  function stop() {
    clearInterval(typing);
    clearTimeout(pause);
    request?.abort();
  }

  async function ask(text: string) {
    request?.abort();
    const ctrl = (request = new AbortController());
    try {
      const found = await ctx.search.answer(text, ctrl.signal, session.place);
      if (ctrl.signal.aborted) return;
      answers = found;
    } catch {
      if (ctrl.signal.aborted) return;
      answers = [];
    }
    asked = true;
  }

  /** Types an example into the field, then shows its answer; static with reduced motion. */
  function play(text: string) {
    stop();
    played = text;
    answers = [];
    asked = false;
    if (reduceMotion()) {
      query = text;
      void ask(text);
      return;
    }
    query = '';
    let n = 0;
    typing = setInterval(() => {
      n += 1;
      query = text.slice(0, n);
      if (n >= text.length) {
        clearInterval(typing);
        void ask(text);
      }
    }, TYPING_MS);
  }

  /** The user types their own question. */
  function typed() {
    stop();
    played = '';
    answers = [];
    asked = false;
    if (query.trim().length < 2) return;
    pause = setTimeout(() => void ask(query), TYPED_ANSWER_MS);
  }

  onMount(() => {
    void ctx.search.examples(session.selection?.template ?? '').then((list) => {
      examples = list;
      const first = list[0];
      if (first) play(first);
    });
    return stop;
  });
</script>

<h1 class="wz-title" tabindex="-1">{t('wizard.search.title')}</h1>
<p class="wz-lead">{t('wizard.search.body')}</p>
<div class="wz-search">
  <svg
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    aria-hidden="true"><circle cx="11" cy="11" r="7" /><path d="M20 20l-4-4" /></svg
  >
  <input
    bind:value={query}
    oninput={typed}
    aria-label={t('wizard.search.label')}
    autocomplete="off"
    spellcheck="false"
  />
</div>
<div class="wz-answer" aria-live="polite">
  {#each answers as a (a.intent + a.title)}
    <div class="wz-paper"><AnswerCard answer={a} /></div>
  {:else}
    {#if asked}<p class="wz-hint">{t(played ? 'wizard.search.pending' : 'wizard.search.none')}</p>{/if}
  {/each}
</div>
{#if examples.length > 0}
  <ul class="wz-examples" aria-label={t('wizard.search.examples')}>
    {#each examples as example (example)}
      <li>
        <button type="button" aria-pressed={played === example} onclick={() => play(example)}
          >{example}</button
        >
      </li>
    {/each}
  </ul>
{/if}
<p class="wz-caption">{t('wizard.search.tip')}</p>
