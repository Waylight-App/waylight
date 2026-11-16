<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { accountLink, type Invite, remainingTime } from '../domain/auth';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AccountLinkBox from './AccountLinkBox.svelte';

  /** Invitation links for members and guests (admins, #296). */
  interface Props {
    auth: AuthUseCases;
    /** Dashboard tabs a guest invitation can share. */
    tabs: { id: string; name: string }[];
  }
  let { auth, tabs }: Props = $props();

  let invites = $state<Invite[]>([]);
  let membersAvailable = $state(false);
  let error = $state('');
  let name = $state('');
  let role = $state<'member' | 'guest'>('guest');
  let shared = $state<string[]>([]);
  let validDays = $state(7);
  let link = $state<{ name: string; url: string; expires: string } | null>(null);

  const remainingLabel = (d: Date) => {
    const r = remainingTime(d, new Date());
    if (r.unit === 'expired') return t('auth.invite_expired');
    if (r.unit === 'days') return t(r.n === 1 ? 'auth.invite_day_left' : 'auth.invite_days_left', { n: r.n });
    return t(r.n === 1 ? 'auth.invite_hour_left' : 'auth.invite_hours_left', { n: r.n });
  };
  const when = (d: Date) => d.toLocaleString(intlLocale(), { dateStyle: 'medium', timeStyle: 'short' });

  async function load() {
    try {
      const [list, members] = await Promise.all([auth.invites(), auth.members()]);
      invites = list;
      membersAvailable = members.available;
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    load();
  });

  async function create(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      const r = await auth.createInvite({ name, role, tabs: shared, validDays });
      link = {
        name: r.invite.name,
        url: accountLink(location.origin, r.token),
        expires: when(r.invite.expiresAt),
      };
      name = '';
      shared = [];
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function revoke(i: Invite) {
    error = '';
    try {
      await auth.revokeInvite(i.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const toggled = (list: string[], id: string, on: boolean) =>
    on ? [...list, id] : list.filter((x) => x !== id);
</script>

<p class="muted">{t('auth.invites_hint')}</p>
{#if link}
  <AccountLinkBox
    title={t('auth.invite_link_for', { name: link.name })}
    url={link.url}
    expires={link.expires}
    ondone={() => (link = null)}
  />
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<ul class="list">
  {#each invites as i (i.id)}
    <li>
      <strong>{i.name}</strong>
      <span class="muted"
        >({i.role === 'member' ? t('auth.invite_role_member') : t('auth.member_guest')})</span
      >
      <span class="muted small" title={t('auth.invite_expires', { time: when(i.expiresAt) })}
        >{remainingLabel(i.expiresAt)}</span
      >
      <button class="btn" type="button" onclick={() => revoke(i)}>{t('auth.invite_revoke')}</button>
    </li>
  {:else}
    <li class="muted">{t('auth.no_invites')}</li>
  {/each}
</ul>

<form class="new" onsubmit={create}>
  <h3>{t('auth.new_invite')}</h3>
  <label class="field"
    ><span>{t('auth.guest_name')}</span><input
      class="input"
      bind:value={name}
      autocomplete="off"
      required
    /></label
  >
  <label class="field"
    ><span>{t('auth.invite_role')}</span>
    <select class="input" bind:value={role}>
      <option value="guest">{t('auth.member_guest')}</option>
      {#if membersAvailable}<option value="member">{t('auth.invite_role_member')}</option>{/if}
    </select>
  </label>
  {#if role === 'guest' && tabs.length}
    <fieldset>
      <legend class="muted">{t('auth.guest_tabs')}</legend>
      {#each tabs as tab (tab.id)}
        <label
          ><input
            type="checkbox"
            checked={shared.includes(tab.id)}
            onchange={(e) => (shared = toggled(shared, tab.id, e.currentTarget.checked))}
          />
          {tab.name}</label
        >
      {/each}
    </fieldset>
  {/if}
  <label class="field"
    ><span>{t('auth.invite_valid')}</span>
    <select class="input" bind:value={validDays}>
      <option value={1}>{t('auth.invite_one_day')}</option>
      <option value={7}>{t('auth.invite_days', { n: 7 })}</option>
      <option value={30}>{t('auth.invite_days', { n: 30 })}</option>
    </select>
  </label>
  <button class="btn btn-primary" type="submit">{t('auth.create_invite')}</button>
</form>

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  fieldset {
    border: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .new {
    display: grid;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
  h3 {
    font-size: 1rem;
    margin: 0;
  }
  .small {
    font-size: 0.85em;
  }
</style>
