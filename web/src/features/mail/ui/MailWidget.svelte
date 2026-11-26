<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { age, type Inbox, parseConfig } from '../domain/mail';
  import type { MailUseCases } from '../usecase/mailUseCases';

  /** Props of the mail widget (#283). */
  interface Props {
    mail: MailUseCases;
    config: Record<string, unknown>;
  }
  let { mail, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseConfig(config));
  let inbox = $state<Inbox | null>(null);
  let error = $state('');
  let now = $state(new Date());

  $effect(() => {
    const account = cfg.account;
    return poll(
      () =>
        mail
          .inbox(account)
          .then((i) => ((inbox = i), (error = ''), (now = new Date())))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      3 * 60_000,
    );
  });
</script>

{#if inbox}
  <p class="head">
    <strong class="count" class:zero={inbox.unread === 0}>{t('mail.unread', { n: inbox.unread })}</strong>
    {#if inbox.webmailUrl}<a href={inbox.webmailUrl} target="_blank" rel="noopener noreferrer"
        >{t('mail.open_webmail')}</a
      >{/if}
  </p>
  <ul class="list">
    {#each inbox.messages as m, i (i)}
      <ListRow nowrap>
        <span class="from">{m.from}</span>
        {#snippet sub()}<span class="subject">{m.subject}</span>{/snippet}
        {#snippet value()}{age(m.date, now)}{/snippet}
      </ListRow>
    {:else}
      <li class="muted">{t('mail.none')}</li>
    {/each}
  </ul>
  {#if inbox.failed.length}<p class="error small">
      {t('mail.failed', { names: inbox.failed.join(', ') })}
    </p>{/if}
{:else if error}
  <p class="muted">{error}</p>
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: var(--space-2);
    margin: 0 0 var(--space-2);
  }
  .count.zero {
    color: var(--color-text-subtle);
    font-weight: 500;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .from {
    font-weight: 600;
  }
  .subject {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
