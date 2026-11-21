<script lang="ts">
  import { tick } from 'svelte';
  import { t } from '../i18n';
  import { errorMessage } from '../api-client/http';
  import {
    addressOf,
    type FoundService,
    type SetupDescriptor,
    type SetupPreview,
    type SetupStage,
    stageComplete,
  } from './setup';

  /**
   * Props of the setup assistant shown in place of a widget that is not set
   * up (#588): address, credential with the service's own steps, then a
   * check that shows real data before the widget takes over.
   */
  interface Props {
    descriptor: SetupDescriptor;
    /** Called after Done; the widget loads its data again. */
    ondone: () => void;
  }
  let { descriptor: d, ondone }: Props = $props();

  type Stage = SetupStage | 'done';
  let stage = $state<Stage>('address');
  let values = $state<Record<string, string>>({});
  let error = $state('');
  let checking = $state(false);
  let preview = $state<SetupPreview | null>(null);
  let found = $state<FoundService[]>([]);
  let heading = $state<HTMLElement>();

  const index = $derived(stage === 'address' ? 1 : stage === 'credential' ? 2 : 3);
  const fieldsOf = (s: SetupStage) => d.fields.filter((f) => f.stage === s);
  const link = $derived(stage === 'credential' ? (d.credentialPage?.(addressOf(d, values)) ?? null) : null);

  // Services found in the home network are offered first (#590); the
  // assistant works the same without them.
  $effect(() => {
    d.discover?.()
      .then((list) => (found = list))
      .catch(() => (found = []));
  });

  async function go(next: Stage) {
    error = '';
    stage = next;
    await tick();
    heading?.focus();
  }

  function useFound(s: FoundService) {
    const field = fieldsOf('address')[0];
    if (field) values[field.key] = s.url;
    void go('credential');
  }

  function next(e: SubmitEvent) {
    e.preventDefault();
    if (stageComplete(d, 'address', values)) void go('credential');
  }

  async function check(e: SubmitEvent) {
    e.preventDefault();
    if (!stageComplete(d, 'credential', values)) return;
    error = '';
    checking = true;
    try {
      preview = await d.check({ ...values });
      await go('done');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      checking = false;
    }
  }
</script>

<section class="setup" aria-labelledby="setup-heading">
  <p class="muted step">{t('setup.step', { n: index, total: 3 })}</p>
  <h4 id="setup-heading" tabindex="-1" bind:this={heading}>
    {stage === 'done'
      ? t('setup.connected', { service: d.service })
      : t('setup.title', { service: d.service })}
  </h4>

  {#if stage === 'address'}
    {#if found.length}
      <ul class="found">
        {#each found as s (s.url)}
          <li>
            <span
              >{t('setup.found', { service: d.service, url: s.url })}<br /><small class="muted"
                >{s.hint}</small
              ></span
            >
            <button class="btn btn-primary" type="button" onclick={() => useFound(s)}
              >{t('setup.connect')}</button
            >
          </li>
        {/each}
      </ul>
    {/if}
    <form onsubmit={next}>
      {#each fieldsOf('address') as f (f.key)}
        <label class="field">
          <span>{t(f.label)}</span>
          <input
            class="input"
            type={f.type}
            placeholder={f.placeholder}
            autocomplete={f.autocomplete ?? 'off'}
            bind:value={values[f.key]}
            required
          />
        </label>
      {/each}
      <button class="btn btn-primary" type="submit">{t('setup.next')}</button>
    </form>
  {:else if stage === 'credential'}
    <ol class="steps">
      {#each d.steps as s (s)}<li>{t(s)}</li>{/each}
    </ol>
    {#if link}
      <a class="btn" href={link} target="_blank" rel="noopener noreferrer"
        >{t('setup.open', { service: d.service })}</a
      >
    {/if}
    <form onsubmit={check}>
      {#each fieldsOf('credential') as f (f.key)}
        <label class="field">
          <span>{t(f.label)}</span>
          <input
            class="input"
            type={f.type}
            placeholder={f.placeholder}
            autocomplete={f.autocomplete ?? (f.type === 'password' ? 'new-password' : 'off')}
            bind:value={values[f.key]}
            required
          />
        </label>
      {/each}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <div class="actions">
        <button class="btn" type="button" onclick={() => go('address')}>{t('setup.back')}</button>
        <button class="btn btn-primary" type="submit" disabled={checking}
          >{checking ? t('setup.checking') : t('setup.check')}</button
        >
      </div>
    </form>
  {:else if preview}
    <ul class="preview" role="status">
      {#each preview.lines as line (line)}<li>{line}</li>{/each}
    </ul>
    {#if preview.images?.length}
      <div class="images">
        {#each preview.images.slice(0, 3) as src (src)}<img {src} alt="" loading="lazy" />{/each}
      </div>
    {/if}
    <button class="btn btn-primary" type="button" onclick={ondone}>{t('setup.done')}</button>
  {/if}
</section>

<style>
  .setup {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .step {
    margin: 0;
    font-size: 0.8rem;
  }
  h4 {
    margin: 0;
  }
  h4:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .steps {
    margin: 0;
    padding-left: 1.4em;
  }
  .steps li + li {
    margin-top: var(--space-1);
  }
  .found,
  .preview {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .found li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .images {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-1);
  }
  .images img {
    width: 100%;
    aspect-ratio: 1;
    object-fit: cover;
    border-radius: var(--radius-small);
  }
</style>
