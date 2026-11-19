import { t } from '../../../shared/i18n';
/** An active login session as shown in the session overview. */
export interface SessionInfo {
  id: string;
  /** Device name of an app session (#559); empty for browsers. */
  device: string;
  userAgent: string;
  clientIp: string;
  createdAt: Date;
  lastSeenAt: Date;
  current: boolean;
}

const browsers: [RegExp, string][] = [
  [/Edg\//, 'Edge'],
  [/OPR\//, 'Opera'],
  [/Firefox\//, 'Firefox'],
  [/Chrome\//, 'Chrome'],
  [/Safari\//, 'Safari'],
  [/curl\//, 'curl'],
];
const systems: [RegExp, string][] = [
  [/Android/, 'Android'],
  [/iPhone|iPad|iOS/, 'iOS'],
  [/Windows/, 'Windows'],
  [/Mac OS X|Macintosh/, 'macOS'],
  [/Linux/, 'Linux'],
];

/** Turns a user agent into a short label such as "Firefox on Linux". */
export function describeUserAgent(ua: string): string {
  const browser = browsers.find(([re]) => re.test(ua))?.[1];
  const os = systems.find(([re]) => re.test(ua))?.[1];
  if (browser && os) return t('auth.browser_on_os', { browser, os });
  return browser ?? os ?? (ua ? ua.slice(0, 40) : t('auth.unknown_device'));
}

/** Names a session: the device of an app session, else its browser. */
export function describeSession(s: Pick<SessionInfo, 'device' | 'userAgent'>): string {
  return s.device ? t('auth.app_on_device', { device: s.device }) : describeUserAgent(s.userAgent);
}
