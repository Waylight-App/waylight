import { intlLocale, t } from '../shared/i18n';
import type { SettingsArea } from './settingsNav';

/**
 * The one-line status of an area tile on the settings start page (#600),
 * e.g. "3 channels on" or "Backup today 03:00". "warn" asks for a look,
 * "ok" confirms that something works, "info" just describes.
 */
export interface AreaStatus {
  text: string;
  tone: 'ok' | 'warn' | 'info';
}

/**
 * What the statuses read. Each source is optional: the shell passes only
 * what the viewer may read, so a member never asks for the owner's
 * backups and gets a 403.
 */
export interface StatusSources {
  twoFactor?: () => Promise<{ enabled: boolean }>;
  channels?: () => Promise<readonly { enabled: boolean }[]>;
  backups?: () => Promise<readonly { createdAt: Date }[]>;
  version?: () => Promise<{ current: string; updateAvailable: boolean; latest: { version: string } | null }>;
  /** Visible connection sections; a count until connections report health (#592). */
  connections?: number;
}

type Loader = (s: StatusSources, now: Date) => Promise<AreaStatus | null> | AreaStatus | null;

const LOADERS: Partial<Record<SettingsArea, Loader>> = {
  account: async (s) => {
    if (!s.twoFactor) return null;
    const { enabled } = await s.twoFactor();
    return enabled
      ? { text: t('settings.status.two_factor_on'), tone: 'ok' }
      : { text: t('settings.status.two_factor_off'), tone: 'warn' };
  },
  notifications: async (s) => {
    if (!s.channels) return null;
    const on = (await s.channels()).filter((c) => c.enabled).length;
    return on === 0
      ? { text: t('settings.status.no_channels'), tone: 'info' }
      : {
          text: t(on === 1 ? 'settings.status.channels_on_one' : 'settings.status.channels_on_other', {
            count: on,
          }),
          tone: 'ok',
        };
  },
  connections: ({ connections: n }) =>
    n
      ? {
          text: t(n === 1 ? 'settings.status.connections_one' : 'settings.status.connections_other', {
            count: n,
          }),
          tone: 'info',
        }
      : null,
  data: async (s, now) => {
    if (!s.backups) return null;
    const latest = newest(await s.backups());
    if (!latest) return { text: t('settings.status.no_backup'), tone: 'warn' };
    return { text: backupText(latest, now), tone: ageDays(latest, now) > STALE_BACKUP_DAYS ? 'warn' : 'ok' };
  },
  system: async (s) => {
    if (!s.version) return null;
    const v = await s.version();
    return v.updateAvailable && v.latest
      ? { text: t('settings.status.update_available', { version: v.latest.version }), tone: 'warn' }
      : { text: t('settings.status.version', { version: v.current }), tone: 'info' };
  },
};

/**
 * A nightly backup older than this means the schedule stopped; two days
 * leave room for one missed night (server restart, full disk warning).
 */
const STALE_BACKUP_DAYS = 2;
const DAY_MS = 24 * 60 * 60 * 1000;

function newest(files: readonly { createdAt: Date }[]): Date | null {
  let best: Date | null = null;
  for (const f of files) if (!best || f.createdAt > best) best = f.createdAt;
  return best;
}

const ageDays = (d: Date, now: Date) => (now.getTime() - d.getTime()) / DAY_MS;

function backupText(d: Date, now: Date): string {
  const time = new Intl.DateTimeFormat(intlLocale(), { hour: '2-digit', minute: '2-digit' }).format(d);
  if (d.toDateString() === now.toDateString()) return t('settings.status.backup_today', { time });
  const date = new Intl.DateTimeFormat(intlLocale(), { day: 'numeric', month: 'short' }).format(d);
  return t('settings.status.backup_on', { date, time });
}

/**
 * Loads the status of every given area. A source that fails (offline,
 * missing permission) leaves its tile without a status instead of
 * breaking the start page.
 */
export async function loadAreaStatuses(
  areas: readonly SettingsArea[],
  sources: StatusSources,
  now = new Date(),
): Promise<Partial<Record<SettingsArea, AreaStatus>>> {
  const entries = await Promise.all(
    areas.map(async (a) => {
      try {
        return [a, (await LOADERS[a]?.(sources, now)) ?? null] as const;
      } catch {
        return [a, null] as const;
      }
    }),
  );
  const out: Partial<Record<SettingsArea, AreaStatus>> = {};
  for (const [a, status] of entries) if (status) out[a] = status;
  return out;
}
