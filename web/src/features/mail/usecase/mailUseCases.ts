import { t } from '../../../shared/i18n';
import type { Inbox, MailAccount } from '../domain/mail';

/** An account to add or change; an empty password keeps the stored one. */
export interface AccountInput {
  name: string;
  host: string;
  port: number;
  username: string;
  password: string;
  folder: string;
  webmailUrl: string;
}

/** Output port: the mail API. */
export interface MailGateway {
  inbox(account: string): Promise<Inbox>;
  accounts(): Promise<MailAccount[]>;
  saveAccount(id: string, input: AccountInput): Promise<MailAccount>;
  deleteAccount(id: string): Promise<void>;
}

/** Mail widget and its accounts (#283). */
export class MailUseCases {
  constructor(private readonly gateway: MailGateway) {}

  inbox(account: string): Promise<Inbox> {
    return this.gateway.inbox(account);
  }
  /** The accounts; guests (no access) get none. */
  accounts(): Promise<MailAccount[]> {
    return this.gateway.accounts().catch(() => []);
  }
  saveAccount(id: string, input: AccountInput): Promise<MailAccount> {
    if (!input.name.trim() || !input.host.trim() || !input.username.trim())
      throw new Error(t('mail.fill_in'));
    if (!id && !input.password) throw new Error(t('mail.password_needed'));
    return this.gateway.saveAccount(id, { ...input, name: input.name.trim(), host: input.host.trim() });
  }
  deleteAccount(id: string): Promise<void> {
    return this.gateway.deleteAccount(id);
  }
}
