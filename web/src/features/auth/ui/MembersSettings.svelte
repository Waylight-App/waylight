<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { accountLink, type Group, type Guest, type Member } from '../domain/auth';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AccountLinkBox from './AccountLinkBox.svelte';

  /** Props of the member and group management section (admins, #294). */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let available = $state(false);
  let members = $state<Member[]>([]);
  let guests = $state<Guest[]>([]);
  let groups = $state<Group[]>([]);
  let error = $state('');
  // A reset link that was just created, shown once (#296).
  let resetLink = $state<{ name: string; url: string; expires: string } | null>(null);

  async function createResetLink(account: { id: string; name: string }) {
    error = '';
    try {
      const r = await auth.createResetLink(account.id);
      resetLink = {
        name: account.name,
        url: accountLink(location.origin, r.token),
        expires: r.expiresAt.toLocaleString(intlLocale(), { dateStyle: 'medium', timeStyle: 'short' }),
      };
    } catch (err) {
      error = errorMessage(err);
    }
  }
  let name = $state('');
  let password = $state('');
  let confirmation = $state('');
  let newPassword = $state<Record<string, string>>({});
  let groupName = $state('');
  // The group being renamed and its draft name (#388); one at a time.
  let renaming = $state<string | null>(null);
  let renameDraft = $state('');
  let privateByDefault = $state(false);

  // Groups can hold members and guests; both lists feed the checkboxes.
  const accounts = $derived([
    ...members.map((m) => ({ id: m.id, name: m.name })),
    ...guests.map((g) => ({ id: g.id, name: `${g.name} (${t('auth.member_guest')})` })),
  ]);

  async function load() {
    try {
      const [list, gs, gr, priv] = await Promise.all([
        auth.members(),
        auth.guests(),
        auth.groups(),
        auth.privateByDefault(),
      ]);
      privateByDefault = priv;
      available = list.available;
      members = list.members;
      guests = gs;
      groups = gr;
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    load();
  });

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function create(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      await auth.createMember(name, password, confirmation);
      name = password = confirmation = '';
    });
  }

  function createGroup(e: SubmitEvent) {
    e.preventDefault();
    run(async () => {
      await auth.createGroup(groupName, []);
      groupName = '';
    });
  }

  function startRename(g: Group) {
    renaming = g.id;
    renameDraft = g.name;
  }

  function rename(e: SubmitEvent, g: Group) {
    e.preventDefault();
    run(async () => {
      await auth.updateGroup(g, renameDraft, g.accounts);
      renaming = null;
    });
  }

  // Group names of a member, in the order of the group list (#389).
  const groupNames = (ids: string[]) =>
    groups
      .filter((g) => ids.includes(g.id))
      .map((g) => g.name)
      .join(', ');

  const toggled = (list: string[], id: string, on: boolean) =>
    on ? [...list, id] : list.filter((x) => x !== id);
</script>

<p class="muted">{t('auth.members_hint')}</p>
{#if !available}<p class="notice" role="status">{t('auth.members_unavailable')}</p>{/if}
{#if resetLink}
  <AccountLinkBox
    title={t('auth.reset_link_for', { name: resetLink.name })}
    url={resetLink.url}
    expires={resetLink.expires}
    ondone={() => (resetLink = null)}
  />
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<!-- Asked once members exist (ADR 0018, migration step 3): until now the
     admin's entries were visible to everyone on the instance. -->
<label class="privacy" class:highlight={members.length > 0 && !privateByDefault}>
  <input
    type="checkbox"
    checked={privateByDefault}
    onchange={(e) => {
      const on = e.currentTarget.checked;
      run(() => auth.setPrivateByDefault(on));
    }}
  />
  <span>
    <strong>{t('auth.private_by_default')}</strong><br />
    <span class="muted">{t('auth.private_by_default_hint')}</span>
  </span>
</label>

<h3>{t('auth.members')}</h3>
<ul class="accounts">
  {#each members as m (m.id)}
    <li>
      <strong>{m.name}</strong>
      {#if m.disabled}<span class="muted">({t('auth.guest_disabled')})</span>{/if}
      {#if groupNames(m.groups)}<div class="muted small">
          {t('auth.member_groups', { groups: groupNames(m.groups) })}
        </div>{/if}
      <div class="row">
        <button class="btn" type="button" onclick={() => run(() => auth.updateMember(m, !m.disabled))}
          >{m.disabled ? t('auth.guest_enable') : t('auth.guest_disable')}</button
        >
        <input
          class="input"
          type="password"
          autocomplete="new-password"
          placeholder={t('auth.guest_new_password')}
          aria-label={t('auth.guest_new_password_for', { name: m.name })}
          bind:value={newPassword[m.id]}
        />
        <button
          class="btn"
          type="button"
          disabled={!newPassword[m.id]}
          onclick={() =>
            run(async () => {
              await auth.updateMember(m, m.disabled, newPassword[m.id] ?? '');
              newPassword[m.id] = '';
            })}>{t('auth.guest_set_password')}</button
        >
        <button class="btn" type="button" onclick={() => createResetLink(m)}
          >{t('auth.reset_link_create')}</button
        >
        <button
          class="btn btn-danger"
          type="button"
          onclick={() =>
            confirm(t('auth.member_delete_confirm', { name: m.name })) && run(() => auth.deleteMember(m.id))}
          >{t('auth.guest_delete')}</button
        >
      </div>
    </li>
  {:else}
    <li class="muted">{t('auth.no_members')}</li>
  {/each}
</ul>

{#if available}
  <form class="new" onsubmit={create}>
    <h3>{t('auth.new_member')}</h3>
    <label class="field"
      ><span>{t('auth.guest_name')}</span><input
        class="input"
        bind:value={name}
        autocomplete="off"
        required
      /></label
    >
    <label class="field"
      ><span>{t('auth.password_at_least_12_characters')}</span><input
        class="input"
        type="password"
        autocomplete="new-password"
        bind:value={password}
        required
      /></label
    >
    <label class="field"
      ><span>{t('auth.repeat_password')}</span><input
        class="input"
        type="password"
        autocomplete="new-password"
        bind:value={confirmation}
        required
      /></label
    >
    <button class="btn btn-primary" type="submit">{t('auth.create_member')}</button>
  </form>
{/if}

<h3>{t('auth.groups')}</h3>
<p class="muted">{t('auth.groups_hint')}</p>
<ul class="accounts">
  {#each groups as g (g.id)}
    <li>
      {#if renaming === g.id}
        <form class="row" onsubmit={(e) => rename(e, g)}>
          <input
            class="input"
            bind:value={renameDraft}
            aria-label={t('auth.group_rename_label', { name: g.name })}
            maxlength="40"
          />
          <button class="btn btn-primary" type="submit">{t('common.save')}</button>
          <button class="btn" type="button" onclick={() => (renaming = null)}>{t('common.cancel')}</button>
        </form>
      {:else}
        <strong>{g.name}</strong>
        <button class="btn" type="button" onclick={() => startRename(g)}>{t('auth.group_rename')}</button>
      {/if}
      <fieldset>
        <legend class="muted">{t('auth.group_accounts')}</legend>
        {#each accounts as a (a.id)}
          <label
            ><input
              type="checkbox"
              checked={g.accounts.includes(a.id)}
              onchange={(e) =>
                run(() => auth.updateGroup(g, g.name, toggled(g.accounts, a.id, e.currentTarget.checked)))}
            />
            {a.name}</label
          >
        {:else}
          <span class="muted">{t('auth.group_no_accounts')}</span>
        {/each}
      </fieldset>
      <button
        class="btn btn-danger"
        type="button"
        onclick={() =>
          confirm(t('auth.group_delete_confirm', { name: g.name })) && run(() => auth.deleteGroup(g.id))}
        >{t('auth.group_delete')}</button
      >
    </li>
  {:else}
    <li class="muted">{t('auth.no_groups')}</li>
  {/each}
</ul>

<form class="new row" onsubmit={createGroup}>
  <input
    class="input"
    bind:value={groupName}
    placeholder={t('auth.group_name')}
    aria-label={t('auth.group_name')}
    maxlength="40"
    required
  />
  <button class="btn btn-primary" type="submit">{t('auth.create_group')}</button>
</form>

<style>
  .accounts {
    list-style: none;
    padding: 0;
  }
  .accounts li {
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  fieldset {
    border: 0;
    padding: 0;
    margin: var(--space-1) 0;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .row .input {
    flex: 1;
    min-width: 10rem;
  }
  .small {
    font-size: 0.85em;
  }
  .new {
    margin-top: var(--space-3);
  }
  .privacy {
    display: flex;
    gap: var(--space-2);
    align-items: flex-start;
    margin: var(--space-2) 0;
    padding: var(--space-2);
    border-radius: var(--radius-2, 8px);
  }
  .privacy.highlight {
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
    outline: 2px solid var(--color-accent, currentColor);
  }
  .notice {
    padding: var(--space-2);
    border-radius: var(--radius-2, 8px);
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
  }
</style>
