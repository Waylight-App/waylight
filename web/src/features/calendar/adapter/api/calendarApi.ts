import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  AgendaDTO,
  ConnectDTO,
  EventDTO,
  SettingsDTO,
  SubscriptionDTO,
} from '../../../../shared/api-client/generated/calendar';
import type {
  AccountDTO,
  AccountsDTO,
  CalendarChoiceDTO,
  ConnectInputDTO,
  NewEventDTO,
  SignInStartDTO,
  SignInStatusDTO,
  SettingsInputDTO,
  SubscriptionInputDTO,
} from '../../../../shared/api-client/generated/calendar-requests';
import {
  type CalendarEvent,
  type CalendarSettings,
  type CalendarSubscription,
  parseLocalDate,
  type SubscriptionInput,
} from '../../domain/calendar';
import type {
  CalendarAccount,
  CalendarAccounts,
  CalendarChoice,
  CalendarProvider,
  SignInStatus,
} from '../../domain/accounts';
import type { Agenda, CalendarConnection, CalendarGateway } from '../../usecase/calendarUseCases';

const toAccount = (d: AccountDTO): CalendarAccount => ({
  ...d,
  provider: d.provider === 'microsoft' ? 'microsoft' : 'google',
  calendars: d.calendars.map((c) => ({ ...c })),
});

const toEvent = (d: EventDTO): CalendarEvent => ({
  uid: d.uid,
  summary: d.summary,
  location: d.location,
  description: d.description,
  url: d.url,
  calendar: d.calendar,
  color: d.color,
  allDay: d.allDay,
  // All-day dates are calendar dates, not instants: interpret them locally.
  start: d.allDay ? parseLocalDate(d.start) : new Date(d.start),
  end: d.allDay ? parseLocalDate(d.end) : new Date(d.end),
});

/** CalendarGateway backed by the REST API. */
export class CalendarApi implements CalendarGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<CalendarSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/calendar/settings')) };
  }
  async saveSettings(urls: string[], username: string, password: string): Promise<CalendarSettings> {
    const body: SettingsInputDTO = { calendarUrls: urls, username, password };
    return { ...(await this.http.put<SettingsDTO>('/api/calendar/settings', body)) };
  }
  async connect(
    server: string,
    username: string,
    password: string,
    signIn?: string,
  ): Promise<CalendarConnection> {
    const body: ConnectInputDTO = { server, username, password, signIn };
    const d = await this.http.post<ConnectDTO>('/api/calendar/connect', body);
    return { calendars: d.calendars.map((c) => ({ ...c })), next: d.next.map(toEvent) };
  }
  async events(days: number): Promise<Agenda> {
    const d = await this.http.get<AgendaDTO>(`/api/calendar/events?days=${days}`);
    return { events: d.events.map(toEvent), error: d.error };
  }
  async range(from: Date, to: Date): Promise<Agenda> {
    const q = new URLSearchParams({ from: from.toISOString(), to: to.toISOString() });
    const d = await this.http.get<AgendaDTO>(`/api/calendar/events?${q}`);
    return { events: d.events.map(toEvent), error: d.error };
  }
  async create(input: NewEventDTO): Promise<CalendarEvent> {
    return toEvent(await this.http.post<EventDTO>('/api/calendar/events', input));
  }

  async accounts(): Promise<CalendarAccounts> {
    const d = await this.http.get<AccountsDTO>('/api/calendar/accounts');
    return { redirectUrl: d.redirectUrl, accounts: d.accounts.map(toAccount) };
  }
  async startSignIn(
    provider: CalendarProvider,
    clientId: string,
    clientSecret: string,
  ): Promise<{ url: string; state: string }> {
    return {
      ...(await this.http.post<SignInStartDTO>('/api/calendar/accounts/oauth/start', {
        provider,
        clientId,
        clientSecret,
      })),
    };
  }
  async signInStatus(state: string): Promise<SignInStatus> {
    const s = (
      await this.http.get<SignInStatusDTO>(`/api/calendar/accounts/oauth/state/${encodeURIComponent(state)}`)
    ).status;
    return s === 'ready' || s === 'failed' ? s : 'pending';
  }
  async setCalendars(id: string, enabled: string[]): Promise<CalendarAccount> {
    return toAccount(
      await this.http.put<AccountDTO>(`/api/calendar/accounts/${encodeURIComponent(id)}`, { enabled }),
    );
  }
  async disconnect(id: string): Promise<void> {
    await this.http.delete(`/api/calendar/accounts/${encodeURIComponent(id)}`);
  }
  async writable(): Promise<CalendarChoice[]> {
    return (await this.http.get<CalendarChoiceDTO[]>('/api/calendar/writable')).map((c) => ({ ...c }));
  }

  async subscriptions(): Promise<CalendarSubscription[]> {
    return (await this.http.get<SubscriptionDTO[]>('/api/calendar/subscriptions')).map((d) => ({ ...d }));
  }
  async addSubscription(input: SubscriptionInput): Promise<CalendarSubscription> {
    const body: SubscriptionInputDTO = input;
    return { ...(await this.http.post<SubscriptionDTO>('/api/calendar/subscriptions', body)) };
  }
  async updateSubscription(id: string, input: SubscriptionInput): Promise<CalendarSubscription> {
    const body: SubscriptionInputDTO = input;
    return {
      ...(await this.http.put<SubscriptionDTO>(
        `/api/calendar/subscriptions/${encodeURIComponent(id)}`,
        body,
      )),
    };
  }
  deleteSubscription(id: string): Promise<void> {
    return this.http.delete(`/api/calendar/subscriptions/${encodeURIComponent(id)}`);
  }
}
