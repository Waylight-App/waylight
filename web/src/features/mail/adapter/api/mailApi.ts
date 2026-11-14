import type { HttpClient } from '../../../../shared/api-client/http';
import type { AccountDTO, InboxDTO } from '../../../../shared/api-client/generated/mail';
import type { Inbox, MailAccount } from '../../domain/mail';
import type { AccountInput, MailGateway } from '../../usecase/mailUseCases';

const ACCOUNTS = '/api/mail/settings/accounts';

/** MailGateway backed by the REST API. */
export class MailApi implements MailGateway {
  constructor(private readonly http: HttpClient) {}

  async inbox(account: string): Promise<Inbox> {
    const d = await this.http.get<InboxDTO>(
      `/api/mail/inbox${account ? `?account=${encodeURIComponent(account)}` : ''}`,
    );
    return {
      unread: d.unread,
      messages: d.messages.map((m) => ({ ...m, date: m.date ? new Date(m.date) : null })),
      webmailUrl: d.webmailUrl,
      failed: d.failed ?? [],
    };
  }
  async accounts(): Promise<MailAccount[]> {
    return (await this.http.get<AccountDTO[]>(ACCOUNTS)).map((a) => ({ ...a }));
  }
  async saveAccount(id: string, input: AccountInput): Promise<MailAccount> {
    return {
      ...(id
        ? await this.http.put<AccountDTO>(`${ACCOUNTS}/${encodeURIComponent(id)}`, input)
        : await this.http.post<AccountDTO>(ACCOUNTS, input)),
    };
  }
  deleteAccount(id: string): Promise<void> {
    return this.http.delete(`${ACCOUNTS}/${encodeURIComponent(id)}`);
  }
}
