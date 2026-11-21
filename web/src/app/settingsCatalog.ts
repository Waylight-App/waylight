import type { MessageKey } from '../shared/i18n';
import type { SettingsArea } from './settingsNav';

/**
 * The sections of the settings dialog and who sees which (#261, #294).
 * Kept outside the dialog so the search can offer a deep link to every
 * section the account may open (#476) without a second list.
 */
export interface CatalogSection {
  id: string;
  label: MessageKey;
  module?: string;
  /** Task area on the settings start page (#600); default "connections". */
  area?: SettingsArea;
  homelab?: boolean;
  /** A plugin's own label (ADR 0021). */
  title?: () => string;
  adminOnly?: boolean;
  plugin?: boolean;
  /** Other words people use for it, in English and German (#601). */
  keywords?: readonly string[];
  /** Labels of fields or headings inside the section the search finds (#601). */
  fields?: readonly MessageKey[];
  /**
   * For developers and power users (#602): listed only while the account
   * shows advanced settings, or when a link opens it directly.
   */
  advanced?: boolean;
  /** Fields of the section that count as advanced; searchable only while shown (#602). */
  advancedFields?: readonly MessageKey[];
}

// The sections built into the dialog. Each belongs to exactly one task area
// of the start page (#600); sections without one are connections. Features
// add theirs from src/app/settings (#193) and choose the area there.
export const OWN_SECTIONS: CatalogSection[] = [
  // My account: sign-in, second factor, sessions.
  {
    id: 'account',
    label: 'settings.section.account',
    area: 'account',
    keywords: [
      'password',
      'login',
      'sign in',
      'security',
      'passwort',
      'anmeldung',
      'sicherheit',
      '2fa',
      'totp',
    ],
    fields: [
      'app.two_factor_authentication',
      'auth.passkeys',
      'app.active_sessions',
      'app.mobile_app',
      'app.change_password',
      'auth.reset_email_title',
      'app.guided_tour',
    ],
  },
  // Appearance & language.
  {
    id: 'appearance',
    label: 'settings.section.appearance',
    module: 'appearance',
    area: 'appearance',
    keywords: ['dark', 'light', 'theme', 'colour', 'color', 'dunkel', 'hell', 'design', 'farbe', 'sprache'],
    fields: [
      'appearance.theme',
      'appearance.gallery',
      'appearance.accent_color',
      'appearance.background_image',
      'appearance.language',
      'appearance.greeting_name',
      'appearance.launcher',
      'appearance.profile',
    ],
    advancedFields: ['appearance.custom_css_override_variables_such_as'],
  },
  // Notifications: every channel, the phone and the web interface.
  {
    id: 'notifications',
    label: 'settings.section.notifications',
    module: 'notifications',
    area: 'notifications',
    keywords: [
      'push',
      'alert',
      'ntfy',
      'gotify',
      'e-mail',
      'webhook',
      'benachrichtigung',
      'warnung',
      'mitteilung',
    ],
    fields: [
      'notifications.channels',
      'notifications.notify_me_when',
      'notifications.quiet_hours',
      'notifications.maintenance',
      'notifications.app_devices',
    ],
  },
  // Dashboard & widgets: what the start page shows.
  {
    id: 'search',
    label: 'settings.section.search',
    module: 'search',
    area: 'dashboard',
    keywords: ['search engine', 'bang', 'suchmaschine', 'suche'],
    fields: ['search.search_field', 'search.own_keywords', 'quick.prefixes', 'search.history_heading'],
  },
  { id: 'feeds', label: 'settings.section.feeds', module: 'rss', area: 'dashboard' },
  { id: 'folders', label: 'settings.section.bookmark_folders', module: 'bookmarks', area: 'dashboard' },
  { id: 'calendar', label: 'settings.section.calendar', module: 'calendar', area: 'dashboard' },
  { id: 'briefing', label: 'settings.section.briefing', module: 'briefing', area: 'dashboard' },
  { id: 'timeline', label: 'settings.section.timeline', module: 'timeline', area: 'dashboard' },
  { id: 'shopping', label: 'settings.section.shopping', module: 'shopping', area: 'dashboard' },
  { id: 'waste', label: 'settings.section.waste', module: 'waste', area: 'dashboard' },
  { id: 'parcels', label: 'settings.section.parcels', module: 'parcels', area: 'dashboard' },
  // Connections & services: the self-hosted services and every integration.
  { id: 'services', label: 'settings.section.services', module: 'services', area: 'connections' },
  { id: 'adguard', label: 'settings.section.adguard', module: 'adguard', homelab: true },
  { id: 'speedtest', label: 'settings.section.speedtest', module: 'speedtest', homelab: true },
  { id: 'docker-updates', label: 'settings.section.docker_updates', module: 'docker', homelab: true },
  { id: 'nextcloud', label: 'settings.section.nextcloud', module: 'nextcloud', homelab: true },
  { id: 'fritzbox', label: 'settings.section.fritzbox', module: 'fritzbox', homelab: true },
  { id: 'wireguard', label: 'settings.section.wireguard', module: 'wireguard', homelab: true },
  {
    id: 'tokens',
    label: 'settings.section.tokens',
    area: 'connections',
    keywords: ['api', 'token', 'key', 'schlüssel'],
    fields: ['auth.new_token'],
    advanced: true,
  },
  // People & access: who else uses this instance and how.
  {
    id: 'members',
    label: 'settings.section.members',
    area: 'people',
    keywords: ['users', 'accounts', 'family', 'benutzer', 'konten', 'familie'],
    fields: ['auth.new_member', 'auth.groups'],
  },
  { id: 'invites', label: 'settings.section.invites', area: 'people', fields: ['auth.new_invite'] },
  { id: 'guests', label: 'settings.section.guests', area: 'people', fields: ['auth.new_guest'] },
  {
    id: 'kiosks',
    label: 'settings.section.kiosks',
    area: 'people',
    keywords: ['kiosk', 'wall', 'tablet', 'display', 'wand', 'anzeige'],
    fields: ['auth.kiosk_new'],
  },
  // Data & backup.
  {
    id: 'backups',
    label: 'settings.section.backups',
    module: 'backup',
    area: 'data',
    keywords: ['backup', 'restore', 'webdav', 'sicherung', 'wiederherstellen'],
    fields: ['backup.offsite', 'backup.restore'],
  },
  {
    id: 'backupmonitor',
    label: 'settings.section.backupmonitor',
    module: 'backupmonitor',
    area: 'data',
    homelab: true,
  },
  {
    id: 'transfer',
    label: 'settings.section.transfer',
    module: 'transfer',
    area: 'data',
    keywords: ['export', 'import', 'json', 'yaml', 'homepage', 'homarr'],
  },
  // System & info.
  {
    id: 'about',
    label: 'settings.section.about',
    module: 'system',
    area: 'system',
    keywords: ['version', 'update', 'release', 'aktualisierung'],
    fields: ['system.update_check'],
  },
  {
    id: 'forecast',
    label: 'settings.section.forecast',
    module: 'system',
    area: 'system',
    homelab: true,
    advanced: true,
  },
  { id: 'security-log', label: 'settings.section.security_log', area: 'system' },
  { id: 'full-version', label: 'settings.section.full_version', area: 'system' },
];

/** What decides whether a section is shown. */
export interface SectionViewer {
  /** Enabled backend modules. */
  modules: ReadonlySet<string>;
  /** Who owns each module's data (ADR 0018); members see personal ones. */
  dataKinds: Record<string, string>;
  /** Appearance profile; "everyday" hides home-server settings (#281). */
  profile: string;
  /** The browser extension's local mode (#291). */
  localMode: boolean;
  guest: boolean;
  member: boolean;
  admin: boolean;
  owner: boolean;
  /** Whether advanced sections are listed (#602); hidden when not set. */
  advanced?: boolean;
}

/** Guests only manage their own account (ADR 0012). */
const guestSections = new Set(['account']);
/**
 * Members (#294) manage their account and the settings of modules with
 * personal data; instance settings stay with admins. Parcels hold the
 * instance's carrier key and the admin's mailbox.
 */
const memberSections = new Set(['account']);
const memberHidden = new Set(['parcels', 'transfer']);
/** Backups, the export and API tokens stay with the owner (ADR 0022). */
const ownerSections = new Set(['backups', 'transfer', 'tokens']);

function allowed(s: CatalogSection, v: SectionViewer): boolean {
  // What a server adds (#291) only makes sense in the extension's local mode.
  if (s.id === 'full-version') return v.localMode;
  if (s.adminOnly && !v.admin) return false;
  // Plugin sections without adminOnly are for every member, e.g. their
  // own connected accounts (#326); guests manage only their account.
  if (s.plugin && !v.guest) return true;
  if (ownerSections.has(s.id) && !v.owner) return false;
  if (v.guest) return guestSections.has(s.id);
  if (v.member) {
    const kind = s.module ? v.dataKinds[s.module] : undefined;
    const personal = kind === 'personal-content' || kind === 'personal-settings';
    return s.module ? personal && !memberHidden.has(s.id) : memberSections.has(s.id);
  }
  return true;
}

/** The sections the viewer may open, in the given order. */
export function visibleSections<T extends CatalogSection>(sections: readonly T[], v: SectionViewer): T[] {
  return sections.filter(
    (s) =>
      (!s.module || v.modules.has(s.module)) &&
      allowed(s, v) &&
      // The "everyday" profile hides home-server and network settings (#281).
      !(s.homelab && v.profile === 'everyday') &&
      !(s.advanced && !v.advanced),
  );
}
