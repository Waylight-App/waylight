/** A calendar service connected by signing in (#282). */
export type CalendarProvider = 'google' | 'microsoft';

export const CALENDAR_PROVIDERS: readonly CalendarProvider[] = ['google', 'microsoft'];

/** One calendar of a connected account. */
export interface AccountCalendar {
  id: string;
  name: string;
  color: string;
  writable: boolean;
  enabled: boolean;
}

/** A connected Google or Microsoft account; tokens stay on the server. */
export interface CalendarAccount {
  id: string;
  provider: CalendarProvider;
  name: string;
  calendars: AccountCalendar[];
}

/** The connected accounts and the redirect address to register ('' = unavailable). */
export interface CalendarAccounts {
  redirectUrl: string;
  accounts: CalendarAccount[];
}

/** A calendar that accepts new events (CalDAV or an account calendar). */
export interface CalendarChoice {
  id: string;
  name: string;
}

export type SignInStatus = 'pending' | 'ready' | 'failed';

/** Where people create the app registration for a provider. */
export const PROVIDER_CONSOLE: Record<CalendarProvider, string> = {
  google: 'https://console.cloud.google.com/apis/credentials',
  microsoft: 'https://entra.microsoft.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade',
};
