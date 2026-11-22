import { intlLocale, t } from '../../../shared/i18n';
import type { FoundService, SetupDescriptor, SetupPreview } from '../../../shared/setup/setup';
import type { CalendarConnection, CalendarUseCases } from '../usecase/calendarUseCases';

/**
 * The setup assistant of the calendar widget for Nextcloud (#588), as
 * data: the server address instead of each calendar's CalDAV URL.
 */
export function calendarSetup(
  calendar: CalendarUseCases,
  discover?: () => Promise<FoundService[]>,
): SetupDescriptor {
  return {
    discover,
    service: t('calendar.setup_service'),
    fields: [
      {
        key: 'server',
        label: 'calendar.setup_server',
        stage: 'address',
        type: 'url',
        placeholder: 'https://cloud.example.com',
      },
      {
        key: 'username',
        hint: 'setup.hint_username',
        label: 'calendar.setup_username',
        stage: 'credential',
        type: 'text',
        autocomplete: 'username',
      },
      { key: 'password', label: 'calendar.setup_password', stage: 'credential', type: 'password' },
    ],
    steps: [
      'calendar.setup_step_settings',
      'calendar.setup_step_security',
      'calendar.setup_step_create',
      'calendar.setup_step_copy',
    ],
    credentialPage: (address) => (address ? `${address}/index.php/settings/user/security` : null),
    // Login Flow v2 (#654): Nextcloud creates the app password itself.
    signIn: {
      service: 'nextcloud',
      label: 'Nextcloud',
      check: async (_values, signIn) => preview(await calendar.connectSignIn(signIn)),
    },
    async check(values) {
      const c = await calendar.connect(
        values['server'] ?? '',
        values['username'] ?? '',
        values['password'] ?? '',
      );
      return preview(c);
    },
  };
}

/** The preview of a connection: calendars found and the next events. */
function preview(c: CalendarConnection): SetupPreview {
  const when = new Intl.DateTimeFormat(intlLocale(), {
    weekday: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });
  return {
    lines: [
      t('calendar.setup_calendars', { list: c.calendars.map((x) => x.name).join(', ') }),
      ...c.next.map((e) => `${when.format(e.start)} ${e.summary}`),
    ],
  };
}
