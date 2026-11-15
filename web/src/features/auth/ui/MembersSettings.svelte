<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { Group, Guest, Member } from '../domain/auth';
  import type { AuthUseCases } from '../usecase/authUseCases';

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
  let name = $state('');
  let password = $state('');
  let confirmation = $state('');
  let newPassword = $state<Record<string, string>>({});
  let groupName = $state('');

  // Groups can hold members and guests; both lists feed the checkboxes.
  const accounts = $derived([
    ...members.map((m) => ({ id: m.id, name: m.name })),
    ...guests.map((g) => ({ id: g.id, name: `${g.name} (${t('auth.member_guest')})` })),
  ]);

  async function load() {
    try {
      const [list, gs, gr] = await Promise.all([auth.members(), auth.guests(), auth.groups()]);
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

  const toggled = (list: string[], id: string, on: boolean) =>
    on ? [...list, id] : list.filter((x) => x !== id);
</script>

<p class="muted">{t('auth.members_hint')}</p>
{#if !available}<p class="notice" role="status">{t('auth.members_unavailable')}</p>{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<h3>{t('auth.members')}</h3>
<ul class="accounts">
  {#each members as m (m.id)}
    <li>
      <strong>{m.name}</strong>
      {#if m.disabled}<span class="muted">({t('auth.guest_disabled')})</span>{/if}
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
      <strong>{g.name}</strong>
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
  .new {
    margin-top: var(--space-3);
  }
  .notice {
    padding: var(--space-2);
    border-radius: var(--radius-2, 8px);
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
  }
</style>
