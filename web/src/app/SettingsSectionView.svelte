<script lang="ts">
  import CaptureBookmarklet from '../features/capture/ui/CaptureBookmarklet.svelte';
  import { TimelineSettings } from '../features/timeline';
  import { BriefingSettings } from '../features/briefing';
  import { AboutSettings, ForecastSettings } from '../features/system';
  import { t } from '../shared/i18n';
  import { registeredSections } from './settingsSections';
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
  import type { Container } from './container';

  /**
   * The content of one settings section. The built-in sections are wired
   * here; features add theirs from src/app/settings (#193) and are loaded
   * on demand. Kept apart from the dialog so navigation and content change
   * independently.
   */
  interface Props {
    id: string;
    container: Container;
    modules: Set<string>;
    appearance: Appearance;
    onappearance: (a: Appearance) => void;
    onrestarttour?: () => void;
    /** Whether advanced fields are shown (#602). */
    advanced: boolean;
  }
  let { id, container, modules, appearance, onappearance, onrestarttour, advanced }: Props = $props();

  const registered = $derived(registeredSections.find((s) => s.id === id));
  let guestTabs = $state<{ id: string; name: string }[]>([]);
  // The wall tablet assistant (#680); the kiosk list is rebuilt after it created a display.
  let tabletAssistant = $state(false);
  let kioskList = $state(0);
  $effect(() => {
    if (!container.auth.guest) container.dashboard.load().then((tabs) => (guestTabs = tabs));
  });
</script>

{#if id === 'appearance'}
  <AppearanceSettings
    appearance={container.appearance}
    current={appearance}
    onchange={onappearance}
    {advanced}
  />
{:else if id === 'search'}
  <SearchSettings search={container.search} keywordTargets={() => keywordTargets(container, modules)} />
  <CaptureBookmarklet />
  <SearchHistorySettings usage={searchUsage} />
  <UnansweredSettings queries={unansweredQueries} />
{:else if id === 'services'}
  <ServicesManager services={container.services} />
{:else if id === 'feeds'}
  <FeedsManager rss={container.rss} />
{:else if id === 'folders'}
  <FoldersManager bookmarks={container.bookmarks} />
  {#if container.auth.admin}<BookmarkLinkSettings bookmarks={container.bookmarks} />{/if}
{:else if id === 'calendar'}
  <CalendarSettings calendar={container.calendar} />
{:else if id === 'transfer'}
  <TransferSettings transfer={container.transfer} />
  {#if modules.has('bookmarks')}<BookmarkImport bookmarks={container.bookmarks} />{/if}
{:else if id === 'security-log'}
  <SecurityLog auth={container.auth} />
{:else if id === 'tokens'}
  <ApiTokensSettings auth={container.auth} {modules} />
{:else if id === 'guests'}
  <GuestsSettings auth={container.auth} tabs={guestTabs} />
{:else if id === 'members'}
  {#await import('../features/auth/ui/MembersSettings.svelte') then { default: MembersSettings }}
    <MembersSettings auth={container.auth} />
  {/await}
{:else if id === 'full-version'}
  {#await import('./local/FullVersionHint.svelte') then { default: FullVersionHint }}
    <FullVersionHint variant="full" />
  {/await}
{:else if id === 'invites'}
  {#await import('../features/auth/ui/InvitesSettings.svelte') then { default: InvitesSettings }}
    <InvitesSettings auth={container.auth} tabs={guestTabs} />
  {/await}
{:else if id === 'kiosks'}
  {#await import('../features/auth/ui/KiosksSettings.svelte') then { default: KiosksSettings }}
    {#key kioskList}
      <KiosksSettings auth={container.auth} tabs={guestTabs} onassistant={() => (tabletAssistant = true)} />
    {/key}
  {/await}
  {#if tabletAssistant}
    {#await import('./kiosk/WallTabletAssistant.svelte') then { default: WallTabletAssistant }}
      <WallTabletAssistant
        {container}
        tabs={guestTabs}
        searchPlaces={modules.has('weather')
          ? async (q) =>
              (await container.weather.searchPlaces(q)).map((p) => ({
                label: [p.name, p.region, p.country].filter(Boolean).join(', '),
                lat: p.lat,
                lon: p.lon,
              }))
          : undefined}
        oncreated={() => {
          kioskList++;
          void container.dashboard.load().then((tabs) => (guestTabs = tabs));
        }}
        onclose={() => (tabletAssistant = false)}
      />
    {/await}
  {/if}
{:else if id === 'timeline'}
  <TimelineSettings timeline={container.timeline} />
{:else if id === 'briefing'}
  <BriefingSettings
    briefing={container.briefing}
    canPush={modules.has('notifications') && container.auth.admin}
  />
{:else if id === 'forecast'}
  <ForecastSettings system={container.system} />
{:else if id === 'about'}
  <AboutSettings system={container.system} />
{:else if id === 'parcels'}
  <ParcelsSettings parcels={container.parcels} />
{:else if id === 'shopping'}
  <ShoppingSettings
    shopping={container.shopping}
    suggest={async () => {
      const c = await container.calendar.settings();
      return c.configured ? { listUrl: suggestListUrl(c.calendarUrls[0] ?? ''), username: c.username } : null;
    }}
  />
{:else if id === 'waste'}
  <WasteSettings waste={container.waste} />
{:else if id === 'wireguard'}
  <WireGuardSettings wireguard={container.wireguard} />
{:else if id === 'fritzbox'}
  <FritzBoxSettings fritzbox={container.fritzbox} />
{:else if id === 'nextcloud'}
  <NextcloudSettings nextcloud={container.nextcloud} />
{:else if id === 'docker-updates'}
  <ImageUpdateSettings docker={container.docker} />
{:else if id === 'speedtest'}
  <SpeedtestSettings speedtest={container.speedtest} />
{:else if id === 'adguard'}
  <AdGuardSettings adguard={container.adguard} />
{:else if registered}
  {#await registered.load() then Section}
    <Section {...registered.props(container, { modules, advanced })} />
  {/await}
{:else if id === 'notifications'}
  <NotificationSettings notifications={container.notifications} />
{:else if id === 'backupmonitor'}
  <BackupSourcesManager backupMonitor={container.backupMonitor} />
{:else if id === 'backups'}
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
  {#if !container.auth.kiosk && !container.localMode}
    {#await import('../features/auth/ui/LinkedSignInSettings.svelte') then { default: LinkedSignInSettings }}
      <LinkedSignInSettings auth={container.auth} />
    {/await}
  {/if}
  <h3>{t('app.active_sessions')}</h3>
  <SessionsList auth={container.auth} />
  {#if !container.auth.kiosk && !container.localMode}
    <h3>{t('app.mobile_app')}</h3>
    {#await import('../features/auth/ui/AppPairing.svelte') then { default: AppPairing }}
      <AppPairing auth={container.auth} />
    {/await}
  {/if}
  <h3>{t('app.change_password')}</h3>
  <ChangePasswordForm auth={container.auth} />
  {#if !container.auth.admin && !container.auth.kiosk}
    <h3>{t('auth.reset_email_title')}</h3>
    {#await import('../features/auth/ui/ResetEmailSettings.svelte') then { default: ResetEmailSettings }}
      <ResetEmailSettings auth={container.auth} />
    {/await}
  {/if}
{/if}
