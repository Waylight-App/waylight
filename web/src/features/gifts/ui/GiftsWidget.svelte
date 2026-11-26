<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { type GiftIdea, groupByPerson, type NextBirthday } from '../domain/gifts';
  import type { BirthdaySource, GiftsUseCases } from '../usecase/giftsUseCases';
  import GiftDialog from './GiftDialog.svelte';

  /** Props of the gift ideas widget (#684). */
  interface Props {
    gifts: GiftsUseCases;
    birthdays: BirthdaySource;
  }
  let { gifts, birthdays }: Props = $props();
  const poll = usePolling();

  let ideas = $state<GiftIdea[]>([]);
  let upcoming = $state<NextBirthday[]>([]);
  let error = $state('');
  let editing = $state<{ idea: GiftIdea | null; person: string } | null>(null);

  const groups = $derived(groupByPerson(ideas, upcoming));
  const people = $derived([...new Set([...groups.map((g) => g.person), ...upcoming.map((b) => b.name)])]);

  async function load() {
    try {
      ideas = await gifts.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
    upcoming = await birthdays.upcoming().catch(() => []);
  }

  $effect(() => poll(load, 30 * 60_000));

  function birthdayLabel(b: NextBirthday): string {
    const date = new Date(`${b.next}T12:00:00`).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'short',
    });
    if (b.inDays === 0) return t('gifts.birthday_today');
    return t('gifts.birthday_in', { date, n: b.inDays });
  }

  async function toggle(idea: GiftIdea) {
    try {
      await gifts.setBought(idea, !idea.bought);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(idea: GiftIdea) {
    if (!confirm(t('common.confirm_delete', { name: idea.title }))) return;
    try {
      await gifts.remove(idea.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if groups.length === 0}
  <p class="muted">{t('gifts.empty')}</p>
{/if}
{#each groups as g (g.person)}
  <section>
    <h3>
      <span class="person">{g.person}</span>
      {#if g.birthday}<span class="birthday muted">{birthdayLabel(g.birthday)}</span>{/if}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('gifts.add_for', { person: g.person })}
        onclick={() => (editing = { idea: null, person: g.person })}>+</button
      >
    </h3>
    <ul>
      {#each g.ideas as idea (idea.id)}
        <ListRow>
          {#snippet lead()}
            <input
              type="checkbox"
              checked={idea.bought}
              aria-label={t('gifts.bought_named', { title: idea.title })}
              onchange={() => toggle(idea)}
            />
          {/snippet}
          <span class:bought={idea.bought}>
            {#if idea.link}
              <a href={idea.link} target="_blank" rel="noopener noreferrer">{idea.title}</a>
            {:else}
              {idea.title}
            {/if}
          </span>
          {#snippet sub()}{#if idea.note}{idea.note}{/if}{/snippet}
          {#snippet trail()}
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('gifts.edit_named', { title: idea.title })}
              onclick={() => (editing = { idea, person: idea.person })}>✎</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('gifts.remove', { title: idea.title })}
              onclick={() => remove(idea)}>✕</button
            >
          {/snippet}
        </ListRow>
      {/each}
    </ul>
  </section>
{/each}
<button class="btn add" type="button" onclick={() => (editing = { idea: null, person: '' })}
  >+ {t('gifts.new')}</button
>

{#if editing}
  <GiftDialog
    {gifts}
    idea={editing.idea}
    person={editing.person}
    {people}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = null)}
  />
{/if}

<style>
  section {
    margin-bottom: var(--space-2);
  }
  h3 {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-sm);
    margin: 0;
  }
  .person {
    font-weight: 600;
  }
  .birthday {
    flex: 1;
    font-size: var(--fs-xs);
    font-weight: normal;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  input[type='checkbox'] {
    width: 1.25rem;
    height: 1.25rem;
    margin: var(--space-3);
    flex-shrink: 0;
  }
  .bought {
    text-decoration: line-through;
    color: var(--color-text-subtle);
  }
</style>
