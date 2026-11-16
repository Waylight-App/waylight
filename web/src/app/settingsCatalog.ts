import type { MessageKey } from '../shared/i18n';
import type { SettingsGroup } from './settingsNav';

/**
 * The sections of the settings dialog and who sees which (#261, #294).
 * Kept outside the dialog so the search can offer a deep link to every
 * section the account may open (#476) without a second list.
 */
export interface CatalogSection {
  id: string;
  label: MessageKey;
  module?: string;
  group?: SettingsGroup;
  homelab?: boolean;
  /** A plugin's own label (ADR 0021). */
  title?: () => string;
  adminOnly?: boolean;
  plugin?: boolean;
}

// The sections built into the dialog, with their topic (#261). Features
// add theirs from src/app/settings (#193) and choose the group there.
export const OWN_SECTIONS: CatalogSection[] = [
  { id: 'appearance', label: 'settings.section.appearance', module: 'appearance', group: 'general' },
  { id: 'search', label: 'settings.section.search', module: 'search', group: 'general' },
  {
    id: 'notifications',
    label: 'settings.section.notifications',
    module: 'notifications',
    group: 'general',
  },
  { id: 'forecast', label: 'settings.section.forecast', module: 'system', group: 'general', homelab: true },
  { id: 'services', label: 'settings.section.services', module: 'services', group: 'content' },
  { id: 'feeds', label: 'settings.section.feeds', module: 'rss', group: 'content' },
  { id: 'folders', label: 'settings.section.bookmark_folders', module: 'bookmarks', group: 'content' },
  { id: 'calendar', label: 'settings.section.calendar', module: 'calendar', group: 'content' },
  { id: 'briefing', label: 'settings.section.briefing', module: 'briefing', group: 'content' },
  { id: 'timeline', label: 'settings.section.timeline', module: 'timeline', group: 'content' },
  { id: 'shopping', label: 'settings.section.shopping', module: 'shopping', group: 'content' },
  { id: 'waste', label: 'settings.section.waste', module: 'waste', group: 'content' },
  { id: 'parcels', label: 'settings.section.parcels', module: 'parcels', group: 'content' },
  { id: 'transfer', label: 'settings.section.transfer', module: 'transfer', group: 'content' },
  { id: 'adguard', label: 'settings.section.adguard', module: 'adguard', homelab: true },
  { id: 'speedtest', label: 'settings.section.speedtest', module: 'speedtest', homelab: true },
  { id: 'docker-updates', label: 'settings.section.docker_updates', module: 'docker', homelab: true },
  { id: 'nextcloud', label: 'settings.section.nextcloud', module: 'nextcloud', homelab: true },
  { id: 'fritzbox', label: 'settings.section.fritzbox', module: 'fritzbox', homelab: true },
  { id: 'wireguard', label: 'settings.section.wireguard', module: 'wireguard', homelab: true },
  { id: 'backups', label: 'settings.section.backups', module: 'backup', group: 'data' },
  {
    id: 'backupmonitor',
    label: 'settings.section.backupmonitor',
    module: 'backupmonitor',
    group: 'data',
    homelab: true,
  },
  { id: 'account', label: 'settings.section.account', group: 'account' },
  { id: 'tokens', label: 'settings.section.tokens', group: 'account' },
  { id: 'guests', label: 'settings.section.guests', group: 'account' },
  { id: 'members', label: 'settings.section.members', group: 'account' },
  { id: 'invites', label: 'settings.section.invites', group: 'account' },
  { id: 'kiosks', label: 'settings.section.kiosks', group: 'account' },
  { id: 'security-log', label: 'settings.section.security_log', group: 'account' },
  { id: 'about', label: 'settings.section.about', module: 'system', group: 'about' },
  { id: 'full-version', label: 'settings.section.full_version', group: 'about' },
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
      !(s.homelab && v.profile === 'everyday'),
  );
}
