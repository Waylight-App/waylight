<script lang="ts">
  import CaptureBookmarklet from '../features/capture/ui/CaptureBookmarklet.svelte';
  import { TimelineSettings } from '../features/timeline';
  import { BriefingSettings } from '../features/briefing';
  import { AboutSettings, ForecastSettings } from '../features/system';
  import { intlLocale, t } from '../shared/i18n';
  import { tick } from 'svelte';
  import { groupSections, moveFocus } from './settingsNav';
  import { registeredSections } from './settingsSections';
  import { type CatalogSection, OWN_SECTIONS, visibleSections } from './settingsCatalog';
  import { AdGuardSettings } from '../features/adguard';
  import { SpeedtestSettings } from '../features/speedtest';
  import { ImageUpdateSettings } from '../features/docker';
  import { FritzBoxSettings } from '../features/fritzbox';
  import { NextcloudSettings } from '../features/nextcloud';
  import { ShoppingSettings, suggestListUrl } from '../features/shopping';
  import { ParcelsSettings } from '../features/parcels';
  import { TransferSettings } from '../features/transfer';
  import { WasteSettings } from '../features/waste';
  import { WireGuardSettings } from '../features/wireguard';
  import { type Appearance, AppearanceSettings } from '../features/appearance';
  import {
    ApiTokensSettings,
    ChangePasswordForm,
    PasskeySettings,
    GuestsSettings,
    SecurityLog,
    SessionsList,
    TwoFactorSettings,
  } from '../features/auth';
  import { BackupSettings } from '../features/backup';
  import { BackupSourcesManager } from '../features/backupmonitor';
  import { NotificationSettings } from '../features/notifications';
  import { BookmarkImport, BookmarkLinkSettings, FoldersManager } from '../features/bookmarks';
  import { CalendarSettings } from '../features/calendar';
  import { FeedsManager } from '../features/rss';
  import { SearchHistorySettings, SearchSettings } from '../features/search';
  import { keywordTargets } from './searchKeywords';
  import { searchUsage, unansweredQueries } from './searchUsage';
  import UnansweredSettings from '../features/search/ui/UnansweredSettings.svelte';
  import { ServicesManager } from '../features/services';
  import Modal from '../shared/ui/Modal.svelte';
  import type { Container } from './container';

  /** Props of the global settings dialog. */
  interface Props {
    container: Container;
    modules: Set<string>;
    /** Who owns each module's data (ADR 0018); members see personal ones. */
    dataKinds?: Record<string, string>;
    appearance: Appearance;
    onappearance: (a: Appearance) => void;
    /** Restarts the guided tour; omitted when the onboarding module is disabled. */
    onrestarttour?: () => void;
    /** Section to open first, e.g. "appearance" from the launcher (#261). */
    section?: string;
    onclose: () => void;
  }
  let {
    container,
    modules,
    dataKinds = {},
    appearance,
    onappearance,
    onrestarttour,
    section,
    onclose,
  }: Props = $props();

  type Section = CatalogSection;
  /** Display name of a section: a plugin's own label or the translated key. */
  const nameOf = (s: Section) => s.title?.() ?? t(s.label);
  const sections: Section[] = [...OWN_SECTIONS, ...registeredSections];
  const visible = $derived(
    visibleSections(sections, {
      modules,
      dataKinds,
      profile: appearance.profile,
      localMode: container.localMode,
      guest: container.auth.guest,
      member: container.auth.member,
      admin: container.auth.admin,
      owner: container.auth.owner,
    }),
  );
  let guestTabs = $state<{ id: string; name: string }[]>([]);
  $effect(() => {
    if (!container.auth.guest) container.dashboard.load().then((tabs) => (guestTabs = tabs));
  });
  const groups = $derived(groupSections(visible, nameOf, intlLocale()));
  const ordered = $derived(groups.flatMap((g) => g.sections));
  // svelte-ignore state_referenced_locally
  let active = $state(section ?? '');
  const current = $derived((visible.some((s) => s.id === active) && active) || ordered[0]?.id || 'account');
  const registered = $derived(registeredSections.find((s) => s.id === current));
  // Phones show the list first and one section at a time (#261).
  // svelte-ignore state_referenced_locally
  let showing = $state(Boolean(section));
  let nav = $state<HTMLElement>();
  let content = $state<HTMLElement>();

  function choose(id: string) {
    active = id;
    showing = true;
    if (content) content.scrollTop = 0;
  }

  /** Arrow keys, Home and End move between the entries; Enter selects. */
  function navKey(e: KeyboardEvent) {
    const buttons = [...(nav?.querySelectorAll<HTMLButtonElement>('button[data-section]') ?? [])];
    const next = moveFocus(e.key, buttons.indexOf(e.target as HTMLButtonElement), buttons.length);
    if (next === null) return;
    e.preventDefault();
    buttons[next]?.focus();
  }

  async function back() {
    showing = false;
    await tick();
    nav?.querySelector<HTMLButtonElement>('button[aria-current="true"]')?.focus();
  }
</script>

<Modal title={t('app.settings')} size="wide" {onclose}>
  <div class="layout" class:showing>
    <nav class="sections" aria-label={t('settings.sections')} bind:this={nav}>
      {#each groups as g (g.id)}
        <div class="group" role="group" aria-labelledby="settings-group-{g.id}">
          <h3 id="settings-group-{g.id}">{t(g.label)}</h3>
          {#each g.sections as s (s.id)}
            <button
              type="button"
              data-section={s.id}
              class:current={current === s.id}
              aria-current={current === s.id ? 'true' : undefined}
              tabindex={current === s.id ? 0 : -1}
              onkeydown={navKey}
              onclick={() => choose(s.id)}>{nameOf(s)}</button
            >
          {/each}
        </div>
      {/each}
    </nav>
    <section
      class="content"
      bind:this={content}
      aria-label={(() => {
        const s = visible.find((x) => x.id === current);
        return s ? nameOf(s) : t('app.settings');
      })()}
    >
      <button class="btn back" type="button" onclick={back}>← {t('settings.back')}</button>
      {#if current === 'appearance'}
        <AppearanceSettings appearance={container.appearance} current={appearance} onchange={onappearance} />
      {:else if current === 'search'}
        <SearchSettings search={container.search} keywordTargets={() => keywordTargets(container, modules)} />
        <CaptureBookmarklet />
        <SearchHistorySettings usage={searchUsage} />
        <UnansweredSettings queries={unansweredQueries} />
      {:else if current === 'services'}
        <ServicesManager services={container.services} />
      {:else if current === 'feeds'}
        <FeedsManager rss={container.rss} />
      {:else if current === 'folders'}
        <FoldersManager bookmarks={container.bookmarks} />
        {#if container.auth.admin}<BookmarkLinkSettings bookmarks={container.bookmarks} />{/if}
      {:else if current === 'calendar'}
        <CalendarSettings calendar={container.calendar} />
      {:else if current === 'transfer'}
        <TransferSettings transfer={container.transfer} />
        {#if modules.has('bookmarks')}<BookmarkImport bookmarks={container.bookmarks} />{/if}
      {:else if current === 'security-log'}
        <SecurityLog auth={container.auth} />
      {:else if current === 'tokens'}
        <ApiTokensSettings auth={container.auth} {modules} />
      {:else if current === 'guests'}
        <GuestsSettings auth={container.auth} tabs={guestTabs} />
      {:else if current === 'members'}
        {#await import('../features/auth/ui/MembersSettings.svelte') then { default: MembersSettings }}
          <MembersSettings auth={container.auth} />
        {/await}
      {:else if current === 'full-version'}
        {#await import('./local/FullVersionHint.svelte') then { default: FullVersionHint }}
          <FullVersionHint variant="full" />
        {/await}
      {:else if current === 'invites'}
        {#await import('../features/auth/ui/InvitesSettings.svelte') then { default: InvitesSettings }}
          <InvitesSettings auth={container.auth} tabs={guestTabs} />
        {/await}
      {:else if current === 'kiosks'}
        {#await import('../features/auth/ui/KiosksSettings.svelte') then { default: KiosksSettings }}
          <KiosksSettings auth={container.auth} tabs={guestTabs} />
        {/await}
      {:else if current === 'timeline'}
        <TimelineSettings timeline={container.timeline} />
      {:else if current === 'briefing'}
        <BriefingSettings
          briefing={container.briefing}
          canPush={modules.has('notifications') && container.auth.admin}
        />
      {:else if current === 'forecast'}
        <ForecastSettings system={container.system} />
      {:else if current === 'about'}
        <AboutSettings system={container.system} />
      {:else if current === 'parcels'}
        <ParcelsSettings parcels={container.parcels} />
      {:else if current === 'shopping'}
        <ShoppingSettings
          shopping={container.shopping}
          suggest={async () => {
            const c = await container.calendar.settings();
            return c.configured
              ? { listUrl: suggestListUrl(c.calendarUrls[0] ?? ''), username: c.username }
              : null;
          }}
        />
      {:else if current === 'waste'}
        <WasteSettings waste={container.waste} />
      {:else if current === 'wireguard'}
        <WireGuardSettings wireguard={container.wireguard} />
      {:else if current === 'fritzbox'}
        <FritzBoxSettings fritzbox={container.fritzbox} />
      {:else if current === 'nextcloud'}
        <NextcloudSettings nextcloud={container.nextcloud} />
      {:else if current === 'docker-updates'}
        <ImageUpdateSettings docker={container.docker} />
      {:else if current === 'speedtest'}
        <SpeedtestSettings speedtest={container.speedtest} />
      {:else if current === 'adguard'}
        <AdGuardSettings adguard={container.adguard} />
      {:else if registered}
        {#await registered.load() then Section}
          <Section {...registered.props(container, { modules })} />
        {/await}
      {:else if current === 'notifications'}
        <NotificationSettings notifications={container.notifications} />
      {:else if current === 'backupmonitor'}
        <BackupSourcesManager backupMonitor={container.backupMonitor} />
      {:else if current === 'backups'}
        <BackupSettings backups={container.backup} />
      {:else}
        {#if onrestarttour}
          <h3>{t('app.guided_tour')}</h3>
          <p class="muted">{t('app.shows_the_most_important_features_of')}</p>
          <button class="btn" type="button" onclick={onrestarttour}>{t('app.restart_the_tour')}</button>
        {/if}
        <h3>{t('app.two_factor_authentication')}</h3>
        <TwoFactorSettings auth={container.auth} />
        <PasskeySettings auth={container.auth} />
        <h3>{t('app.active_sessions')}</h3>
        <SessionsList auth={container.auth} />
        <h3>{t('app.change_password')}</h3>
        <ChangePasswordForm auth={container.auth} />
        {#if !container.auth.admin && !container.auth.kiosk}
          <h3>{t('auth.reset_email_title')}</h3>
          {#await import('../features/auth/ui/ResetEmailSettings.svelte') then { default: ResetEmailSettings }}
            <ResetEmailSettings auth={container.auth} />
          {/await}
        {/if}
      {/if}
    </section>
  </div>
</Modal>

<style>
  /* Two columns: the navigation stays, the section scrolls on its own (#261). */
  .layout {
    display: grid;
    grid-template-columns: minmax(180px, 230px) 1fr;
    gap: var(--space-4);
    height: min(72vh, 760px);
  }
  .sections,
  .content {
    overflow-y: auto;
    min-height: 0;
    overscroll-behavior: contain;
  }
  .sections {
    border-right: 1px solid var(--color-border);
    padding-right: var(--space-2);
  }
  .content {
    min-width: 0;
    padding-right: var(--space-1);
  }
  .group {
    display: flex;
    flex-direction: column;
    margin-bottom: var(--space-3);
  }
  .group h3 {
    margin: 0 0 var(--space-1);
    padding: 0 var(--space-2);
    font-size: 0.75rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--color-text-muted);
  }
  .sections button {
    text-align: left;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    padding: 6px var(--space-2);
    border-radius: var(--radius-small);
    cursor: pointer;
  }
  .sections button:hover {
    background: var(--color-surface-2);
  }
  .sections button.current {
    background: color-mix(in srgb, var(--color-accent) 18%, transparent);
    font-weight: 600;
  }
  .sections button:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: -2px;
  }
  .back {
    display: none;
  }
  /* Phones: the list of sections, then one section with a back button. */
  @media (max-width: 640px) {
    .layout {
      display: block;
      height: auto;
    }
    .sections {
      border-right: 0;
      padding-right: 0;
    }
    .sections button {
      padding: 10px var(--space-2);
    }
    .layout .content,
    .layout.showing .sections {
      display: none;
    }
    .layout.showing .content {
      display: block;
    }
    .back {
      display: inline-flex;
      margin-bottom: var(--space-3);
    }
  }
</style>
