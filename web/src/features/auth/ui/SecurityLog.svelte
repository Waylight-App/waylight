<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import {
    type SecurityEvent,
    eventLabel,
    eventTypes,
    filterByType,
    isWarning,
  } from '../domain/securityEvent';
  import { describeUserAgent } from '../domain/session';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the security log section. */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let events = $state<SecurityEvent[]>([]);
  let error = $state('');
  let type = $state('');
  const types = $derived(eventTypes(events));
  const shown = $derived(filterByType(events, type));

  $effect(() => {
    auth
      .securityEvents()
      .then((l) => (events = l))
      .catch((err) => (error = errorMessage(err)));
  });
</script>

<p class="muted">{t('auth.security_log_hint')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if types.length > 1}
  <label class="field filter"
    ><span>{t('auth.security_log_filter')}</span><select class="input" bind:value={type}>
      <option value="">{t('auth.security_log_all')}</option>
      {#each types as ty (ty)}<option value={ty}>{t(eventLabel(ty))}</option>{/each}
    </select></label
  >
{/if}
<!-- Scrollable, so it must be reachable with the keyboard (WCAG 2.1.1, #113). -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<ul class="events" tabindex="0" aria-label={t('settings.section.security_log')}>
  {#each shown as e (e.id)}
    <li class:warning={isWarning(e.type)}>
      <span class="what"
        >{t(eventLabel(e.type))}{#if e.detail}
          · {e.detail}{/if}</span
      >
      <span class="muted small"
        >{e.at.toLocaleString(intlLocale())}{#if e.ip}
          · {e.ip}{/if}{#if e.userAgent}
          · {describeUserAgent(e.userAgent)}{/if}</span
      >
    </li>
  {:else}
    <li class="muted">{t('auth.security_log_empty')}</li>
  {/each}
</ul>

<style>
  .filter {
    max-width: 20rem;
  }
  .events {
    list-style: none;
    padding: 0;
    max-height: 22rem;
    overflow-y: auto;
  }
  li {
    display: flex;
    flex-direction: column;
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  .warning .what {
    color: var(--color-warning, #b45309);
    font-weight: 600;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
