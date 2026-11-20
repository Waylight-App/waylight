<script lang="ts">
  import PlainHttpWarning from '../../../shared/ui/PlainHttpWarning.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import {
    type AppDevice,
    type Channel,
    type ChannelInput,
    alertTypes,
    isChatChannel,
    newChannel,
  } from '../domain/notifications';
  import type { NotificationsUseCases } from '../usecase/notificationsUseCases';

  /** Props of the additional channels section (#87). */
  interface Props {
    notifications: NotificationsUseCases;
  }
  let { notifications }: Props = $props();

  let channels = $state<Channel[]>([]);
  let devices = $state<AppDevice[]>([]);
  let editing = $state<{ id: string | null; form: ChannelInput; hasSecret: boolean } | null>(null);
  let error = $state('');
  let message = $state('');

  const load = () =>
    Promise.all([notifications.channels(), notifications.appDevices()])
      .then(([c, d]) => ((channels = c), (devices = d)))
      .catch((err) => (error = errorMessage(err)));
  const hasAppChannel = $derived(channels.some((c) => c.kind === 'app'));
  const hasBrowserChannel = $derived(channels.some((c) => c.kind === 'browser'));
  // The browser's permission for system notifications (#583).
  let permission = $state(typeof Notification === 'undefined' ? 'unsupported' : Notification.permission);
  async function allowNotifications() {
    permission = await Notification.requestPermission();
  }

  $effect(() => {
    load();
  });

  function edit(c: Channel) {
    const { id, hasSecret, ...rest } = c;
    editing = { id, form: { ...rest, secret: '' }, hasSecret };
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    if (!editing) return;
    error = message = '';
    try {
      await notifications.saveChannel(editing.id, editing.form);
      editing = null;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function act(action: () => Promise<unknown>, done: string) {
    error = message = '';
    try {
      await action();
      message = done;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const toggleType = (list: string[], type: string, on: boolean) =>
    on ? [...list, type] : list.filter((x) => x !== type);
</script>

<h3>{t('notifications.channels')}</h3>
<p class="muted">{t('notifications.channels_hint')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}
<ul class="list">
  {#each channels as c (c.id)}
    <li>
      <span
        ><strong>{c.name}</strong>
        <span class="muted">({c.kind}, ≥ {t(`notifications.severity.${c.minSeverity}`)})</span></span
      >
      <span class="actions">
        <button
          class="btn"
          type="button"
          onclick={() => act(() => notifications.testChannel(c.id), t('notifications.test_sent'))}
          >{t('notifications.test')}</button
        >
        <button class="btn" type="button" onclick={() => edit(c)}>{t('common.edit')}</button>
        <button
          class="btn btn-danger"
          type="button"
          onclick={() =>
            confirm(t('common.confirm_delete', { name: c.name })) &&
            act(() => notifications.deleteChannel(c.id), '')}>{t('notifications.delete')}</button
        >
      </span>
    </li>
  {:else}
    <li class="muted">{t('notifications.no_channels')}</li>
  {/each}
</ul>

{#if editing}
  <form class="editor" onsubmit={save}>
    <label class="field"
      ><span>{t('notifications.channel_name')}</span><input
        class="input"
        bind:value={editing.form.name}
        required
      /></label
    >
    {#if editing.form.kind === 'app'}
      <p class="muted small">{t('notifications.app_hint')}</p>
    {:else if editing.form.kind === 'browser'}
      <p class="muted small">{t('notifications.browser_hint')}</p>
    {:else if isChatChannel(editing.form.kind)}
      <p class="muted small">
        {t(editing.form.kind === 'slack' ? 'notifications.slack_hint' : 'notifications.teams_hint')}
      </p>
    {:else if editing.form.kind !== 'email'}
      <label class="field"
        ><span
          >{editing.form.kind === 'gotify'
            ? t('notifications.gotify_url')
            : t('notifications.webhook_url')}</span
        ><input class="input" type="url" bind:value={editing.form.url} required /></label
      >
      <PlainHttpWarning url={editing.form.url} />
    {:else}
      <div class="row">
        <label class="field"
          ><span>{t('notifications.smtp_host')}</span><input
            class="input"
            bind:value={editing.form.smtpHost}
            required
          /></label
        >
        <label class="field"
          ><span>{t('notifications.smtp_port')}</span><input
            class="input"
            type="number"
            min="1"
            max="65535"
            bind:value={editing.form.smtpPort}
          /></label
        >
        <label class="field"
          ><span>{t('notifications.smtp_tls')}</span><select class="input" bind:value={editing.form.tls}>
            <option value="starttls">STARTTLS</option>
            <option value="tls">TLS</option>
            <option value="none">{t('notifications.smtp_no_tls')}</option>
          </select></label
        >
        <label class="field"
          ><span>{t('notifications.smtp_user')}</span><input
            class="input"
            bind:value={editing.form.username}
            autocomplete="off"
          /></label
        >
        <label class="field"
          ><span>{t('notifications.mail_from')}</span><input
            class="input"
            type="email"
            bind:value={editing.form.from}
            required
          /></label
        >
        <label class="field"
          ><span>{t('notifications.mail_to')}</span><input
            class="input"
            bind:value={editing.form.to}
            required
          /></label
        >
      </div>
    {/if}
    {#if editing.form.kind !== 'app' && editing.form.kind !== 'browser'}<label class="field"
        ><span
          >{isChatChannel(editing.form.kind)
            ? t('notifications.chat_webhook_url')
            : editing.form.kind === 'gotify'
              ? t('notifications.gotify_token')
              : editing.form.kind === 'webhook'
                ? t('notifications.webhook_secret')
                : t('notifications.smtp_password')}{#if editing.hasSecret}
            {t('common.keep_stored')}{/if}</span
        ><input
          class="input"
          type="password"
          autocomplete="new-password"
          bind:value={editing.form.secret}
          required={isChatChannel(editing.form.kind) && !editing.hasSecret}
        /></label
      >{/if}
    {#if editing.form.kind === 'webhook'}
      <label class="field"
        ><span>{t('notifications.webhook_template')}</span><textarea
          class="input"
          rows="3"
          bind:value={editing.form.template}
          placeholder={'{"title":{{title}},"message":{{body}},"severity":{{severity}},"type":{{type}}}'}
        ></textarea></label
      >
    {/if}
    <label class="field"
      ><span>{t('notifications.min_severity')}</span><select
        class="input"
        bind:value={editing.form.minSeverity}
      >
        <option value="info">{t('notifications.severity.info')}</option>
        <option value="warning">{t('notifications.severity.warning')}</option>
        <option value="critical">{t('notifications.severity.critical')}</option>
      </select></label
    >
    <fieldset class="types">
      <legend>{t('notifications.channel_types')}</legend>
      {#each alertTypes as type (type)}
        <label
          ><input
            type="checkbox"
            checked={editing.form.types.includes(type)}
            onchange={(e) =>
              editing && (editing.form.types = toggleType(editing.form.types, type, e.currentTarget.checked))}
          />
          {t(`notifications.type.${type}`)}</label
        >
      {/each}
    </fieldset>
    <label
      ><input type="checkbox" bind:checked={editing.form.enabled} />
      {t('notifications.channel_enabled')}</label
    >
    <p class="actions">
      <button class="btn btn-primary" type="submit">{t('common.save')}</button>
      <button class="btn" type="button" onclick={() => (editing = null)}>{t('common.cancel')}</button>
    </p>
  </form>
{:else}
  <p class="actions">
    <button
      class="btn"
      type="button"
      onclick={() => (editing = { id: null, form: newChannel('gotify'), hasSecret: false })}>+ Gotify</button
    >
    <button
      class="btn"
      type="button"
      onclick={() => (editing = { id: null, form: newChannel('webhook'), hasSecret: false })}
      >+ Webhook</button
    >
    <button
      class="btn"
      type="button"
      onclick={() => (editing = { id: null, form: newChannel('email'), hasSecret: false })}
      >+ {t('notifications.email')}</button
    >
    <button
      class="btn"
      type="button"
      onclick={() => (editing = { id: null, form: newChannel('slack'), hasSecret: false })}>+ Slack</button
    >
    <button
      class="btn"
      type="button"
      onclick={() => (editing = { id: null, form: newChannel('teams'), hasSecret: false })}
      >+ Microsoft Teams</button
    >
    {#if !hasAppChannel}
      <button
        class="btn"
        type="button"
        onclick={() =>
          (editing = {
            id: null,
            form: { ...newChannel('app'), name: t('notifications.app') },
            hasSecret: false,
          })}>+ {t('notifications.app')}</button
      >
    {/if}
    {#if !hasBrowserChannel}
      <button
        class="btn"
        type="button"
        onclick={() =>
          (editing = {
            id: null,
            form: { ...newChannel('browser'), name: t('notifications.browser') },
            hasSecret: false,
          })}>+ {t('notifications.browser')}</button
      >
    {/if}
  </p>
{/if}

{#if hasBrowserChannel && permission !== 'unsupported'}
  <p class="small">
    {#if permission === 'granted'}
      <span class="muted">{t('notifications.browser_allowed')}</span>
    {:else if permission === 'denied'}
      <span class="muted">{t('notifications.browser_denied')}</span>
    {:else}
      <button class="btn" type="button" onclick={allowNotifications}
        >{t('notifications.browser_allow')}</button
      >
    {/if}
  </p>
{/if}

{#if hasAppChannel || devices.length > 0}
  <h4>{t('notifications.app_devices')}</h4>
  <ul class="list">
    {#each devices as d (d.session)}
      <li>
        <span
          ><strong>{d.name}</strong>
          <span class="muted">({d.platform === 'apns' ? 'iOS' : 'Android'})</span></span
        >
        <button
          class="btn btn-danger"
          type="button"
          onclick={() =>
            confirm(t('common.confirm_delete', { name: d.name })) &&
            act(() => notifications.removeAppDevice(d.session), '')}
          >{t('notifications.app_device_remove')}</button
        >
      </li>
    {:else}
      <li class="muted">{t('notifications.app_devices_empty')}</li>
    {/each}
  </ul>
{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
    flex-wrap: wrap;
  }
  .actions {
    display: flex;
    gap: var(--space-1);
    flex-wrap: wrap;
  }
  .row {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
    gap: var(--space-2);
  }
  .types {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    border: 0;
    padding: 0;
  }
</style>
