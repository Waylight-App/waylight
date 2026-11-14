/** An IMAP account of the mail widget (#283); the password is never sent back. */
export interface MailAccount {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  folder: string;
  webmailUrl: string;
}

export interface MailMessage {
  account: string;
  from: string;
  subject: string;
  date: Date | null;
}

export interface Inbox {
  unread: number;
  messages: MailMessage[];
  webmailUrl: string;
  failed: string[];
}

/** Widget configuration: one account, or "" for all accounts. */
export function parseConfig(raw: Record<string, unknown>): { account: string } {
  const a = raw['account'];
  return { account: typeof a === 'string' && /^[A-Za-z0-9_-]{1,64}$/.test(a) ? a : '' };
}

/** "5 min", "3 h", "2 d": age of a message. */
export function age(date: Date | null, now: Date): string {
  if (!date) return '';
  const min = Math.max(0, Math.round((now.getTime() - date.getTime()) / 60_000));
  if (min < 60) return `${min} min`;
  const h = Math.round(min / 60);
  return h < 48 ? `${h} h` : `${Math.round(h / 24)} d`;
}
