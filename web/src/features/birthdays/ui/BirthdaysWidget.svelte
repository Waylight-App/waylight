<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { type Birthday, parseConfig, SOON_DAYS, whenKey } from '../domain/birthdays';
  import type { BirthdaysUseCases } from '../usecase/birthdaysUseCases';
  import BirthdayDialog from './BirthdayDialog.svelte';

  /** Props of the birthdays widget (#287). */
  interface Props {
    birthdays: BirthdaysUseCases;
    config: Record<string, unknown>;
  }
  let { birthdays, config }: Props = $props();
  const poll = usePolling();

  const cfg = $derived(parseConfig(config));
  let items = $state<Birthday[]>([]);
  let error = $state('');
  let adding = $state(false);

  async function load() {
    try {
      items = await birthdays.upcoming(cfg.limit);
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  // The list only changes at midnight or after a sync; an hourly refresh is plenty.
  $effect(() => poll(load, 60 * 60_000));

  function when(b: Birthday): string {
    const w = whenKey(b.inDays);
    return t(w.key, w.n === undefined ? {} : { n: w.n });
  }

  function date(b: Birthday): string {
    // Built from the parts so the time zone cannot shift the day.
    return new Date(2000, b.month - 1, b.day).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'long',
    });
  }

  async function remove(b: Birthday) {
    if (!confirm(t('common.confirm_delete', { name: b.name }))) return;
    try {
      await birthdays.remove(b.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul>
  {#each items as b (b.id)}
    <li class:today={b.inDays === 0} class:soon={b.inDays > 0 && b.inDays < SOON_DAYS}>
      <span class="icon" aria-hidden="true">{b.inDays === 0 ? '🎂' : '🎈'}</span>
      <span class="main">
        <span class="name">{b.name}</span>
        <span class="muted small"
          >{date(b)}{#if b.age > 0}
            · {t('birthdays.turns', { age: b.age })}{/if}</span
        >
      </span>
      <span class="when">{when(b)}</span>
      <span class="actions">
        {#if b.source === 'carddav'}
          <span class="muted small" title={t('birthdays.from_book')} aria-label={t('birthdays.from_book')}
            >☁</span
          >
        {:else}
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('birthdays.remove', { name: b.name })}
            onclick={() => remove(b)}>✕</button
          >
        {/if}
      </span>
    </li>
  {:else}
    <li class="muted">{t('birthdays.empty')}</li>
  {/each}
</ul>
<button class="btn add" type="button" onclick={() => (adding = true)}>+ {t('birthdays.new')}</button>

{#if adding}
  <BirthdayDialog {birthdays} onsaved={() => load().catch(() => {})} onclose={() => (adding = false)} />
{/if}

<style>
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
  }
  .icon {
    font-size: 1.2rem;
    width: 1.6rem;
    text-align: center;
  }
  .main {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .name,
  .small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8rem;
  }
  .when {
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  li.today .when,
  li.today .name {
    font-weight: 600;
    color: var(--color-accent);
  }
  li.soon .when {
    font-weight: 600;
  }
  .actions {
    display: flex;
    min-width: 2rem;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.15s;
  }
  li:hover .actions,
  li:focus-within .actions {
    opacity: 1;
  }
  @media (hover: none) {
    .actions {
      opacity: 1;
    }
  }
</style>
