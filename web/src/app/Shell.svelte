<script lang="ts">
  import WaylightMark from '../shared/ui/WaylightMark.svelte';
  import { DOCS_URL, PRODUCT_NAME } from '../shared/product';
  import { setupWalk } from '../shared/setup/walk.svelte';
  import { rememberLook } from './branding/lastLook';
  import type { CommandHandlers } from './appCommands';
  import { hasStoredTimers, timerAlerts } from './timerAlerts.svelte';
  import { liveAlerts } from './liveAlerts.svelte';
  import LiveRegion from '../shared/a11y/LiveRegion.svelte';
  import SyncStatus from './offline/SyncStatus.svelte';
  import MaintenanceBanner from '../features/notifications/ui/MaintenanceBanner.svelte';
  import { setLanguage, t } from '../shared/i18n';
  import { tick } from 'svelte';
  import {
    type Appearance,
    accessibilityAttributes,
    cssVariables,
    defaultAppearance,
  } from '../features/appearance';
  import { errorMessage } from '../shared/api-client/http';
  import { applyTheme } from '../shared/theme/theme';
  import type { Container } from './container';
  import { provideRefresh } from '../shared/refresh';
  import DashboardPage from './DashboardPage.svelte';
  import { branding } from './branding/branding.svelte';
  import { loadPlugins, type OnboardingStep } from './plugins/loader';
  import type { Personal } from '../features/onboarding/domain/profile';
  import { stageOf } from '../features/onboarding/domain/stage';
  import { definitionFor, widgetDefinitions } from './widgets';
  import { searchActions } from './searchActions';
  import { registeredSections } from './settingsSections';
  import { OWN_SECTIONS } from './settingsCatalog';
  import type { AttentionState } from '../features/attention/ui/attentionState.svelte';
  import { provideOpenSettings } from '../shared/ui/settingsLink';
  import { provideSetupRights } from '../shared/setup/setup';
  import { providePaste } from '../shared/setup/paste';
  import { provideSignIn } from '../shared/setup/signin';
  import { attachShortcuts } from '../shared/shortcuts';
  import Icon from '../shared/ui/icons/Icon.svelte';
  import AccountMenu from './AccountMenu.svelte';
  import Hero from './Hero.svelte';
  import TopInfo from './TopInfo.svelte';
  import { launcherSources } from './launcherSources';
  import { LauncherUseCases } from '../features/launcher/usecase/launcherUseCases';
  import { type Segment, statusSummary } from '../features/launcher/domain/launcher';
  import { firstWidgetConfig, type Tab } from '../features/dashboard';
  import { parseWeatherConfig } from '../features/weather/domain/weather';

  /** Props of the authenticated application shell. */
  interface Props {
    container: Container;
    /** Called after logout so the app can show the login screen. */
    onlogout: () => void;
  }
  let { container, onlogout }: Props = $props();
  // Components outside widgets (e.g. the maintenance banner) poll through it too.
  provideRefresh({
    get scheduler() {
      return container.refresh;
    },
    track: () => () => {},
  });

  let settingsOpen = $state(false);
  /** Section the settings open with (#261), e.g. Appearance from the launcher. */
  let settingsSection = $state<string | undefined>();
  // Widgets that need setting up open their settings section (#512).
  provideOpenSettings((section) => {
    settingsSection = section;
    settingsOpen = true;
  });
  // Widgets that are not set up offer the setup assistant in place (#588)
  // to those who may save the integration; the local mode has no server.
  // svelte-ignore state_referenced_locally
  if (!container.localMode) {
    provideSetupRights({
      admin: container.auth.admin,
      guest: container.auth.guest || container.auth.kiosk,
    });
    provideSignIn(container.setupSignin);
    providePaste(container.setupPaste);
  }
  let editing = $state(false);
  let error = $state('');
  let appearance = $state<Appearance>(defaultAppearance);
  // True once the account's own appearance has loaded, not the placeholder.
  let appearanceLoaded = $state(false);
  // The sign-in page shows this look next time on this browser (#526).
  $effect(() => {
    if (appearanceLoaded && !container.localMode)
      rememberLook({ css: appearance.customCss, theme: appearance.theme });
  });
  let modules = $state<Set<string> | null>(null);
  let widgetKinds = $state<Set<string>>(new Set());
  let dataKinds = $state<Record<string, string>>({});
  let touring = $state(false);
  let shortcutHelp = $state(false);
  let dashboard = $state<{
    selectTabAt: (index: number) => void;
    selectTab: (id: string) => void;
    openWidgetSettings: (tabId: string, widgetId: string) => void;
    openAddWidget: (query?: string) => void;
    addTab: (tab: Tab) => void;
    reload: (selectId?: string) => Promise<void> | undefined;
    undo: () => void;
    redo: () => void;
  }>();
  let hero = $state<{ focus: () => void }>();

  // ?focus=search (the browser extension's new tab, #65): put the cursor into
  // the search field once it exists, then drop the parameter from the address.
  let focusRequested = new URLSearchParams(location.search).get('focus') === 'search';
  $effect(() => {
    if (!focusRequested || !hero) return;
    focusRequested = false;
    hero.focus();
    const url = new URL(location.href);
    url.searchParams.delete('focus');
    history.replaceState(history.state, '', url.pathname + url.search + url.hash);
  });

  // ?settings=<section> (#291): the browser extension sends someone who moves
  // from its local mode to their server straight to Settings → Transfer. The
  // dialog falls back to its first section for an unknown or hidden one.
  let settingsRequested = new URLSearchParams(location.search).get('settings');
  $effect(() => {
    if (!settingsRequested || !modules) return;
    settingsSection = settingsRequested;
    settingsOpen = true;
    settingsRequested = null;
    const url = new URL(location.href);
    url.searchParams.delete('settings');
    history.replaceState(history.state, '', url.pathname + url.search + url.hash);
  });
  /**
   * Opens the settings at a section. An open dialog keeps the section it
   * opened with, so it is closed and opened again.
   */
  async function openSettingsAt(section: string) {
    if (settingsOpen) {
      settingsOpen = false;
      await tick();
    }
    settingsSection = section;
    settingsOpen = true;
  }

  // The attention list (#592): the dot in the top bar and Settings → Needs
  // your attention. Guests, kiosk displays and the local mode can fix
  // nothing, so they never ask.
  const attentionOn = $derived(
    !!modules?.has('attention') && !container.auth.guest && !container.auth.kiosk && !container.localMode,
  );
  // Loaded only where it is used, so the extension's local mode and guests
  // do not carry it (ADR 0011).
  let attention = $state<AttentionState | null>(null);
  $effect(() => {
    if (!attentionOn) return;
    let cancelled = false;
    let stop = () => {};
    void import('./attention.svelte').then((m) => {
      if (cancelled) return;
      m.setAttentionHandlers(attentionHandlers);
      const state = m.attention;
      state.attach(container.attention);
      attention = state;
      void state.refresh();
      const poll = () => {
        if (!document.hidden) void state.refresh();
      };
      const id = setInterval(poll, 60_000);
      document.addEventListener('visibilitychange', poll);
      stop = () => {
        clearInterval(id);
        document.removeEventListener('visibilitychange', poll);
        state.attach(null);
      };
    });
    return () => {
      cancelled = true;
      stop();
      attention = null;
    };
  });
  const attentionHandlers = {
    openWidget: (tabId: string, widgetId: string) => {
      settingsOpen = false;
      dashboard?.openWidgetSettings(tabId, widgetId);
    },
    openSettings: (section: string) => void openSettingsAt(section),
    sectionOf: (integration: string) =>
      [...OWN_SECTIONS, ...registeredSections].find((s) => s.module === integration && s.id !== 'attention')
        ?.id,
    nameOf: (integration: string) => {
      const s = [...OWN_SECTIONS, ...registeredSections].find(
        (x) => x.module === integration && x.id !== 'attention',
      );
      if (s) return s.title?.() ?? t(s.label);
      return definitionFor(integration)?.type.title ?? integration;
    },
    tabName: (tabId: string) => dashboardState.tabs.find((tab) => tab.id === tabId)?.name,
  } satisfies import('./attention.svelte').AttentionHandlers;
  // A changed widget may have got its missing setting: look again shortly
  // after the dashboard changed, without asking the services.
  $effect(() => {
    void dashboardState.tabs;
    const state = attention;
    if (!state) return;
    const id = setTimeout(() => void state.refresh(), 1_000);
    return () => clearTimeout(id);
  });

  function closeSettings() {
    settingsOpen = false;
    // A connection may have been fixed in the settings: ask its service
    // again, but only when something was broken.
    if (attention && attention.count > 0) void attention.check();
  }

  let launcherOpen = $state<Segment | null>(null);
  let launcherTrigger = $state<HTMLButtonElement>();
  let serviceStatus = $state<ReturnType<typeof statusSummary>>({ kind: 'none', offline: 0 });
  const launcher = $derived(modules ? new LauncherUseCases(launcherSources(container, modules)) : null);

  // Status pill (#189): all online, or how many services are offline.
  $effect(() => {
    const l = launcher;
    if (!l?.hasServices) return;
    const poll = () =>
      l
        .services()
        .then((s) => (serviceStatus = statusSummary(s)))
        .catch(() => {});
    poll();
    const id = setInterval(poll, 60_000);
    return () => clearInterval(id);
  });

  function closeLauncher() {
    launcherOpen = null;
    launcherTrigger?.focus();
  }

  const accountInitial = $derived(appearance.greetingName.trim().charAt(0).toUpperCase());
  let dashboardState = $state<{ tabs: Tab[]; activeId: string }>({ tabs: [], activeId: '' });
  // CONFIG_FILE can own the dashboard; then the layout is read-only (#88).
  let dashboardManaged = $state(false);
  // The simple view has no edit controls (#681).
  const simpleView = $derived(!!appearance.simpleView);
  $effect(() => {
    if (simpleView) editing = false;
  });
  $effect(() => {
    // Config as code and the appliance are the admin's business (ADR 0018).
    if (!modules?.has('transfer') || !container.auth.admin) return;
    container.transfer
      .managedSections()
      .then((m) => (dashboardManaged = m.includes('dashboard')))
      .catch(() => undefined);
  });
  // Appliances ask once how the device is used (#140).
  let deviceSetup = $state(false);
  $effect(() => {
    if (!modules?.has('device') || !container.auth.admin) return;
    container.device
      .state()
      .then((s) => (deviceSetup = s.requested === ''))
      .catch(() => undefined);
  });
  // Guests and kiosk displays cannot act; their action buttons are hidden (#50).
  $effect(() => {
    document.documentElement.classList.toggle('read-only', container.auth.guest);
    return () => document.documentElement.classList.remove('read-only');
  });

  // The top bar shows the weather of the first weather widget (#167, #189).
  const weatherLocation = $derived.by(() => {
    const cfg = firstWidgetConfig(dashboardState.tabs, dashboardState.activeId, 'weather');
    return cfg ? parseWeatherConfig(cfg) : null;
  });

  // Undo/redo shortcuts only exist in edit mode (#93).
  $effect(() => {
    if (!editing) return;
    const offs = [
      container.shortcuts.register({
        id: 'edit.undo',
        combos: ['Mod+Z'],
        display: [t('shortcuts.ctrl_z')],
        label: 'shortcuts.undo',
        run: () => dashboard?.undo(),
      }),
      container.shortcuts.register({
        id: 'edit.redo',
        combos: ['Mod+Shift+Z', 'Mod+Y'],
        display: [t('shortcuts.ctrl_shift_z'), t('shortcuts.ctrl_y')],
        label: 'shortcuts.redo',
        run: () => dashboard?.redo(),
      }),
      // Esc finishes editing; open menus and pickers handle Esc themselves
      // (dialogs are excluded by the registry already) (#271).
      container.shortcuts.register({
        id: 'edit.exit',
        combos: ['Escape'],
        display: ['Esc'],
        label: 'shortcuts.finish_edit',
        run: () => {
          if (!document.querySelector('[role="menu"]')) editing = false;
        },
      }),
    ];
    return () => offs.forEach((off) => off());
  });

  // Global shortcuts (#95): one registry, one listener.
  $effect(() => {
    const r = container.shortcuts;
    const offs = [
      r.register({
        id: 'search.open',
        combos: ['Mod+K', '/'],
        display: [t('app.ctrl_k'), '/'],
        label: 'shortcuts.search',
        typingCombos: ['Mod+K'],
        run: () => hero?.focus(), // the hero field is the search (#222)
      }),
      r.register({
        id: 'tabs.select',
        combos: ['1', '2', '3', '4', '5', '6', '7', '8', '9'],
        display: ['1–9'],
        label: 'shortcuts.tabs',
        run: (combo) => dashboard?.selectTabAt(Number(combo) - 1),
      }),
      ...(container.auth.guest
        ? []
        : [
            r.register({
              id: 'edit.toggle',
              combos: ['e'],
              display: ['E'],
              label: 'shortcuts.edit',
              run: () => {
                if (!dashboardManaged && !simpleView) editing = !editing;
              },
            }),
          ]),
      r.register({
        id: 'shortcuts.help',
        combos: ['?'],
        display: ['?'],
        label: 'shortcuts.help',
        run: () => (shortcutHelp = true),
      }),
    ];
    const detach = attachShortcuts(r);
    return () => {
      detach();
      offs.forEach((off) => off());
    };
  });
  // Timers started in the search survive a reload (#474): load them again,
  // but only when some are stored, so other pages do not pay for the code.
  // svelte-ignore state_referenced_locally
  if (!container.auth.kiosk && hasStoredTimers()) void import('./timers');

  // Live notifications of the browser channel (#583): admins only, and
  // only when such a channel is on, so other accounts open no stream.
  $effect(() => {
    if (!modules?.has('notifications') || !container.auth.admin || container.auth.kiosk) return;
    let stop: (() => void) | undefined;
    let cancelled = false;
    container.notifications
      .channels()
      .then(async (channels) => {
        if (cancelled || !channels.some((c) => c.kind === 'browser' && c.enabled)) return;
        const [{ startLiveNotifications }, { addLiveAlert }] = await Promise.all([
          import('../features/notifications/adapter/live/liveStream'),
          import('./liveAlerts.svelte'),
        ]);
        if (cancelled) return;
        stop = startLiveNotifications((n, leader) => {
          addLiveAlert(n);
          // One system notification per notification: from the tab that
          // holds the stream, and only while no Waylight tab is in view.
          if (
            leader &&
            document.hidden &&
            typeof Notification !== 'undefined' &&
            Notification.permission === 'granted'
          ) {
            new Notification(n.title, { body: n.body, tag: `waylight-${n.at}` });
          }
        });
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
      stop?.();
    };
  });

  let offline = $state(!navigator.onLine);

  // The first-start flow previews language and look before they are saved (#671).
  let lookPreview = $state<Partial<Personal> | null>(null);
  const shownLook = $derived(lookPreview ? { ...appearance, ...lookPreview } : appearance);

  $effect(() => applyTheme(shownLook.theme));
  $effect(() => setLanguage(shownLook.language));

  // Accessibility options (#681): text size, high contrast and the simple
  // view are attributes of the root, so tokens.css and base.css apply them.
  $effect(() => {
    const root = document.documentElement;
    const attrs = accessibilityAttributes(appearance);
    for (const [name, value] of Object.entries(attrs)) {
      if (value === undefined) root.removeAttribute(name);
      else root.setAttribute(name, value);
    }
    return () => Object.keys(attrs).forEach((name) => root.removeAttribute(name));
  });

  $effect(() => {
    const root = document.documentElement;
    const vars = cssVariables(shownLook);
    for (const [k, v] of Object.entries(vars)) root.style.setProperty(k, v);
    return () => Object.keys(vars).forEach((k) => root.style.removeProperty(k));
  });

  $effect(() => {
    // textContent (not innerHTML) so user CSS can never inject markup.
    const style = document.createElement('style');
    style.dataset['source'] = 'custom-css';
    style.textContent = appearance.customCss;
    document.head.append(style);
    return () => style.remove();
  });

  // Plugins (the ee/ frontend, ADR 0021) are loaded once per page, before
  // the modules are read, so their widgets and settings are known when the
  // dashboard renders. Local mode has no server to offer any.
  let pluginsLoaded: Promise<unknown> | undefined;
  /** First-start steps of the plugins, e.g. the instance name (#671). */
  const onboardingSteps: OnboardingStep[] = [];

  async function init() {
    try {
      if (!container.localMode) {
        pluginsLoaded ??= loadPlugins(container.http, container.auth, {
          widgets: widgetDefinitions,
          sections: registeredSections,
          searchActions,
          onboardingSteps,
        });
        await pluginsLoaded;
      }
      const info = await container.enabledModules();
      modules = info.modules;
      widgetKinds = info.widgetKinds;
      dataKinds = info.dataKinds;
      if (info.modules.has('appearance')) {
        appearance = await container.appearance.load();
        appearanceLoaded = true;
      }
      if (!container.auth.guest && info.modules.has('onboarding')) {
        const status = await container.onboarding.status();
        const stage = stageOf(status);
        // The public demo shares one account and keeps its showcase, and
        // the local mode has no server to ask: both go straight to the tour.
        if (stage !== 'finished' && (container.localMode || container.auth.demo)) setTimeout(startTour, 600);
        // The wizard comes first (#739), the tour after it.
        else if (stage === 'wizard') wizardStep = status.step;
        // Wait a moment so the dashboard is rendered before the tour measures it.
        else if (stage === 'tour') setTimeout(startTour, 600);
      }
    } catch (err) {
      error = errorMessage(err);
    } finally {
      onboardingChecked = true;
    }
  }

  /** The dashboard waits until it is known whether the wizard comes first (#739). */
  let onboardingChecked = $state(false);
  /** The step the wizard opens at (empty before it started); null while it is not shown. */
  let wizardStep = $state<string | null>(null);

  /** Saves personal settings from the wizard (#671). */
  async function savePersonal(patch: Partial<Personal>) {
    appearance = await container.appearance.save({ ...appearance, ...patch });
  }

  /** Applies a design from the wizard and shows it. */
  async function applyDesign(id: string) {
    appearance = await container.appearance.applyTheme(id);
    return appearance;
  }

  /** The wizard is finished or skipped: the dashboard opens, then the tour. */
  async function wizardDone() {
    wizardStep = null;
    lookPreview = null;
    try {
      await container.onboarding.finishWizard();
    } catch (err) {
      error = errorMessage(err);
    }
    // A setup walk started on the done page comes first; the tour follows it.
    if (setupWalk.active) setupWalk.whenFinished(startTour);
    else setTimeout(startTour, 600);
  }

  // The widget kind whose help is open (#678): from a widget's "?" or a
  // help result of the search. Kiosk displays offer neither.
  let helpKind = $state<string | null>(null);
  const showHelp = $derived(container.auth.kiosk ? undefined : (kind: string) => (helpKind = kind));

  function startTour() {
    editing = false;
    touring = true;
  }

  function finishTour() {
    touring = false;
    container.onboarding.finish().catch((err) => (error = errorMessage(err)));
  }

  async function restartTour() {
    settingsOpen = false;
    try {
      await container.onboarding.restart();
      setTimeout(startTour, 200);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  /** Changes the theme from the search (#476); undone by the opposite command. */
  async function setTheme(mode: 'light' | 'dark' | 'system') {
    try {
      appearance = await container.appearance.save({ ...appearance, theme: mode });
    } catch (err) {
      error = errorMessage(err);
    }
  }

  // App commands of the search (#476): the handlers live here, in the
  // composition root, so the search imports no other feature. A missing
  // handler means the command is not offered.
  const commandHandlers: CommandHandlers = {
    get theme() {
      return modules?.has('appearance') && !container.auth.guest && !appearance.themeLocked
        ? setTheme
        : undefined;
    },
    settings: (section) => ((settingsSection = section), (settingsOpen = true)),
    get addWidget() {
      return container.auth.guest || dashboardManaged || simpleView
        ? undefined
        : (query: string) => ((editing = true), dashboard?.openAddWidget(query));
    },
    get edit() {
      return container.auth.guest || dashboardManaged || simpleView ? undefined : () => (editing = true);
    },
    shortcuts: () => (shortcutHelp = true),
    get widgetHelp() {
      return showHelp;
    },
    get tour() {
      return modules?.has('onboarding') && !container.auth.guest ? restartTour : undefined;
    },
    logout: () => void logout(),
    tab: (id) => dashboard?.selectTab(id),
    tabs: () => dashboardState.tabs,
    sectionViewer: () => ({
      modules: modules ?? new Set(),
      dataKinds,
      profile: appearance.profile,
      localMode: container.localMode,
      guest: container.auth.guest,
      member: container.auth.member,
      admin: container.auth.admin,
      owner: container.auth.owner,
      advanced: appearance.showAdvanced,
    }),
  };

  async function logout() {
    // Waiting offline changes are only on this device (#80).
    if (
      container.offline.pending > 0 &&
      !confirm(t('offline.logout_pending', { n: container.offline.pending }))
    )
      return;
    try {
      await container.auth.logout();
      await container.offline.clear();
      onlogout();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  // After offline changes reached the server, widgets show the server's state (#80).
  // svelte-ignore state_referenced_locally
  let seenSync = container.offline.revision; // the container never changes
  $effect(() => {
    const r = container.offline.revision;
    if (r === seenSync) return;
    seenSync = r;
    container.refresh.refreshAll();
  });
  // Replay what was changed offline in an earlier session.
  // svelte-ignore state_referenced_locally
  void container.offline.replay();

  init();
</script>

<svelte:window ononline={() => (offline = false)} onoffline={() => (offline = true)} />

<!-- The app frame (#849) holds top bar, header and grid; kiosk displays stay full screen. -->
<div class="app-frame" class:framed={!container.auth.kiosk && wizardStep === null}>
  <!-- Top bar and hero (#189): brand · date, weather, to-dos · status, launcher, account. -->
  {#if wizardStep === null}
    <header class="topbar">
      <a class="brand" href="/">
        {#if branding().logo}
          <!-- An organisation's logo (#305); the dark variant follows the theme. -->
          <img class="brand-logo light" src={branding().logo} alt="" height="28" />
          <img class="brand-logo dark" src={branding().logoDark || branding().logo} alt="" height="28" />
        {:else}
          <span class="logo"><WaylightMark size={24} /></span>
        {/if}
        <span class:wordmark={branding().name === PRODUCT_NAME}>{branding().name}</span>
      </a>
      <div class="center">
        <TopInfo
          weather={modules?.has('weather') ? container.weather : null}
          location={weatherLocation}
          todos={modules?.has('todos') ? container.todos : null}
        />
      </div>
      <div class="right">
        {#if attention && attention.count > 0}
          {@const state = attention}
          {#await import('../features/attention/ui/AttentionBadge.svelte') then { default: AttentionBadge }}
            <AttentionBadge {state} onopen={() => void openSettingsAt('attention')} />
          {/await}
        {/if}
        {#if serviceStatus.kind !== 'none'}
          <button
            type="button"
            class="status"
            data-kind={serviceStatus.kind}
            data-launcher-trigger
            onclick={() => (launcherOpen = launcherOpen ? null : 'services')}
          >
            <span class="dot" aria-hidden="true"></span>
            <span class="long">
              {serviceStatus.kind === 'ok'
                ? t('launcher.all_online')
                : t('launcher.some_offline', { n: serviceStatus.offline })}
            </span>
            <span class="short">
              {serviceStatus.kind === 'ok'
                ? t('launcher.online_short')
                : t('launcher.some_offline', { n: serviceStatus.offline })}
            </span>
          </button>
        {/if}
        {#if launcher}
          <button
            bind:this={launcherTrigger}
            type="button"
            class="icon-btn"
            data-launcher-trigger
            aria-haspopup="dialog"
            aria-expanded={launcherOpen !== null}
            aria-label={t('launcher.title')}
            title={t('launcher.title')}
            onclick={() => (launcherOpen = launcherOpen ? null : 'all')}
            ><Icon name="grid" size={16} /></button
          >
        {/if}
        {#if !container.auth.guest && !simpleView && !dashboardManaged}
          <button
            type="button"
            class="icon-btn edit"
            aria-pressed={editing}
            aria-label={t('app.edit_layout')}
            title={t('app.edit_layout')}
            onclick={() => (editing = !editing)}><Icon name={editing ? 'done' : 'pencil'} size={16} /></button
          >
        {/if}
        <AccountMenu
          initial={accountInitial}
          {editing}
          canEdit={!container.auth.guest && !simpleView}
          editLocked={dashboardManaged ? t('app.edit_managed') : ''}
          onedit={() => (editing = !editing)}
          onsettings={() => ((settingsSection = undefined), (settingsOpen = true))}
          onlogout={logout}
        />
      </div>
      {#if launcherOpen && launcher}
        {#await import('../features/launcher/ui/LauncherPanel.svelte') then { default: LauncherPanel }}
          <LauncherPanel
            {launcher}
            content={appearance.launcher}
            initialSegment={launcherOpen}
            onclose={closeLauncher}
            onsettings={() => (
              (launcherOpen = null),
              (settingsSection = 'appearance'),
              (settingsOpen = true)
            )}
          />
        {/await}
      {/if}
    </header>
    <Hero
      bind:this={hero}
      name={appearance.greetingName}
      moment={appearance.greetingMoment}
      tabsId="header-tabs"
      {container}
      {modules}
      {weatherLocation}
      widgetConfig={(kind) => firstWidgetConfig(dashboardState.tabs, dashboardState.activeId, kind)}
      {commandHandlers}
    />
  {/if}

  {#if deviceSetup}
    <!-- The setup step of an appliance (#140): asked once, until a mode is chosen. -->
    {#await import('../features/device/ui/DeviceSetup.svelte') then { default: DeviceSetup }}
      <DeviceSetup device={container.device} onclose={() => (deviceSetup = false)} />
    {/await}
  {/if}
  {#if container.auth.kiosk}
    {#await import('./kiosk/KioskMode.svelte') then { default: KioskMode }}
      <KioskMode
        loadOptions={() => container.auth.kioskOptions()}
        tabCount={dashboardState.tabs.length}
        activeIndex={Math.max(
          0,
          dashboardState.tabs.findIndex((tab) => tab.id === dashboardState.activeId),
        )}
        selectTabAt={(i) => dashboard?.selectTabAt(i)}
      />
    {/await}
  {/if}
  <LiveRegion />
  {#if liveAlerts.list.length}
    {#await import('./LiveToasts.svelte') then { default: LiveToasts }}<LiveToasts />{/await}
  {/if}
  {#if timerAlerts.list.length}
    {#await import('./TimerToasts.svelte') then { default: TimerToasts }}<TimerToasts />{/await}
  {/if}
  <main class="content">
    {#if offline}<p class="offline" role="status">
        {t('app.you_are_offline_showing_the_last')}
      </p>{/if}
    <SyncStatus sync={container.offline} />
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if modules && onboardingChecked && wizardStep === null}
      {#if modules.has('notifications') && container.auth.admin}<MaintenanceBanner
          notifications={container.notifications}
        />{/if}
      <DashboardPage
        bind:this={dashboard}
        {container}
        {editing}
        {widgetKinds}
        profile={appearance.profile}
        tabBarTarget="header-tabs"
        onstate={(tabs, activeId) => (dashboardState = { tabs, activeId })}
        onedit={container.auth.guest || dashboardManaged || simpleView ? undefined : () => (editing = true)}
        {simpleView}
        globalBackground={appearance.backgroundImage}
        onappearance={(a) => (appearance = a)}
        ondone={() => (editing = false)}
        onhelp={showHelp}
      />
    {/if}
  </main>
</div>

{#if settingsOpen && modules}
  <!-- Loaded on first use: the settings are not needed for the first paint (ADR 0011). -->
  {#await import('./SettingsDialog.svelte') then { default: SettingsDialog }}
    <SettingsDialog
      {container}
      {modules}
      {dataKinds}
      {appearance}
      onappearance={(a) => (appearance = a)}
      onrestarttour={modules.has('onboarding') ? restartTour : undefined}
      section={settingsSection}
      onclose={closeSettings}
    />
  {/await}
{/if}

{#if shortcutHelp}
  {#await import('./ShortcutsDialog.svelte') then { default: ShortcutsDialog }}
    <ShortcutsDialog shortcuts={container.shortcuts.all()} onclose={() => (shortcutHelp = false)} />
  {/await}
{/if}

{#if wizardStep !== null && modules}
  {#await import('./OnboardingWizard.svelte') then { default: OnboardingWizard }}
    <OnboardingWizard
      {container}
      {modules}
      {widgetKinds}
      {appearance}
      {onboardingSteps}
      stored={wizardStep}
      {savePersonal}
      {applyDesign}
      preview={(patch) => (lookPreview = patch)}
      ondone={() => void wizardDone()}
    />
  {/await}
{/if}

{#if helpKind}
  {@const kind = helpKind}
  {#await import('../features/help/ui/WidgetHelpDialog.svelte') then { default: WidgetHelpDialog }}
    <WidgetHelpDialog
      name={definitionFor(kind)?.type.title ?? kind}
      load={() => container.help.widget(kind)}
      fallback={definitionFor(kind)?.type.description ?? ''}
      docsUrl={DOCS_URL}
      onclose={() => (helpKind = null)}
    />
  {/await}
{/if}

{#if touring}
  {#await import('../features/onboarding/ui/Tour.svelte') then { default: Tour }}
    <Tour onfinish={finishTour} />
  {/await}
{/if}

<style>
  .offline {
    margin: 0 0 var(--space-3);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-sm);
    background: color-mix(in srgb, var(--color-warning) 22%, transparent);
  }
  /* Top bar (#189): 64 px, brand left, information centred, actions right. */
  .topbar {
    position: relative;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: var(--space-4);
    height: 64px;
    padding: 0 var(--space-6);
    max-width: 1600px;
    margin: 0 auto;
  }
  .brand-logo {
    height: 28px;
    width: auto;
    max-width: 10rem;
    object-fit: contain;
  }
  .brand-logo.dark,
  :global(html[data-theme='dark']) .brand-logo.light {
    display: none;
  }
  :global(html[data-theme='dark']) .brand-logo.dark {
    display: inline;
  }
  .brand {
    /* Only the logo and the name are the link, not the whole first column. */
    justify-self: start;
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-weight: 600;
    color: inherit;
    text-decoration: none;
    white-space: nowrap;
  }
  .wordmark {
    font-family: var(--font-wordmark);
    font-size: var(--fs-sm);
    line-height: var(--lh-sm);
    font-weight: 600;
  }
  .logo {
    display: inline-grid;
    place-items: center;
    color: var(--color-accent);
  }
  .center {
    min-width: 0;
  }
  .right {
    justify-self: end;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    height: 32px;
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: none;
    color: var(--color-text-muted);
    font: inherit;
    font-size: var(--fs-xs);
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
  }
  .status .short {
    display: none;
  }
  .status .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-ok);
  }
  .status[data-kind='offline'] {
    color: var(--color-danger);
  }
  .status[data-kind='offline'] .dot {
    background: var(--color-danger);
  }
  /* Icon buttons of the bar: 32 px, subtle until hovered. */
  .icon-btn {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    display: grid;
    place-items: center;
    border: 0;
    background: none;
    color: var(--color-text-subtle);
    cursor: pointer;
    transition:
      color 0.15s,
      background-color 0.15s;
  }
  .icon-btn:hover,
  .icon-btn[aria-expanded='true'],
  .icon-btn[aria-pressed='true'] {
    color: var(--color-text);
    background: var(--color-surface-2);
  }
  /* The glance line needs the width; below 1100 px it goes. */
  @media (max-width: 1100px) {
    .topbar {
      grid-template-columns: 1fr auto;
    }
    .center {
      display: none;
    }
  }
  /* Phones: 56 px bar with logo, a short status pill and the avatar. */
  @media (max-width: 639px) {
    .topbar {
      height: 56px;
      padding: 0 var(--space-4);
    }
    .icon-btn.edit {
      display: none;
    }
    .status .long {
      display: none;
    }
    .status .short {
      display: inline;
    }
  }
  .content {
    /* Leaves room for the bottom navigation of phones (#771). */
    padding: var(--space-6) var(--space-6) calc(var(--space-6) + var(--bottom-nav-h, 0px));
    max-width: 1600px;
    margin: 0 auto;
  }
  @media (max-width: 639px) {
    .content {
      padding-inline: var(--space-4);
      padding-top: var(--space-4);
    }
  }
  /* App frame (#849): a hairline box up to 1440 px, centred on the page
     background. Theme tokens only, so every theme draws it in its own border
     colour, high contrast with its stronger one. It has no background, so
     page and tab backgrounds show through, and it does not clip, so menus
     and panels can open past its edge. Phones keep the full width. */
  @media (min-width: 640px) {
    .app-frame.framed {
      width: min(1440px, 100% - 2 * var(--space-6));
      margin: var(--space-6) auto;
      border: 1px solid var(--color-border);
      border-radius: var(--radius-md);
    }
  }
</style>
