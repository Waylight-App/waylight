import type { ServerMessageKey } from '../../../api-client/generated/server-messages/calendar';

/**
 * German translations of the backend error messages of `calendar` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The account has no calendar at this address.
  srv_6f0585956c2b489f: 'Das Konto hat unter dieser Adresse keinen Kalender.',
  // At most 10 calendars are supported.
  srv_326bb72e415244d3: 'Es werden höchstens 10 Kalender unterstützt.',
  // Events can be at most one year long.
  srv_38ba62864162fcec: 'Termine können höchstens ein Jahr lang sein.',
  // No calendar is configured yet.
  srv_92fe9be68e4fbeb4: 'Es ist noch kein Kalender eingerichtet.',
  // Please choose one of the configured calendars.
  srv_f5627673a11b56bd: 'Bitte wähle einen der eingerichteten Kalender.',
  // Please enter a valid start and end.
  srv_b71f32ebd27a0b6a: 'Bitte gib einen gültigen Beginn und ein gültiges Ende ein.',
  // Please enter at least one calendar link.
  srv_014d310b9d3ee071: 'Bitte mindestens eine Kalenderlink eingeben.',
  // The calendar server is not reachable.
  srv_3e73df212f6df359: 'Der Kalenderserver ist nicht erreichbar.',
  // The event must end after it starts.
  srv_db999c150db53756: 'Der Termin muss nach seinem Beginn enden.',
  // The location must be at most 200 characters long.
  srv_eeec5bbc9fb36341: 'Der Ort darf höchstens 200 Zeichen lang sein.',
  // The period must be positive and at most 42 days long.
  srv_066ea148faee11ae: 'Der Zeitraum muss positiv und höchstens 42 Tage lang sein.',
  // The title must be between 1 and 200 characters long.
  srv_f79c303515b1250c: 'Der Titel muss zwischen 1 und 200 Zeichen lang sein.',
  // from and to must be RFC 3339 timestamps.
  srv_ac92ff8f414a83b9: 'from und to müssen RFC-3339-Zeitstempel sein.',
  // At most {max} calendar subscriptions are supported.
  srv_f74740ac20c25101: 'Höchstens {max} Kalender-Abos werden unterstützt.',
  // Could not load the ICS calendar: {detail}
  srv_97f22fc39ea1794a: 'Der ICS-Kalender konnte nicht geladen werden: {detail}',
  // The calendar subscription does not exist.
  srv_210a34361dc2d298: 'Das Kalender-Abo existiert nicht.',
  // The colour must look like #3b6ef5.
  srv_5fd12f163f5132ff: 'Die Farbe muss wie #3b6ef5 aussehen.',
  // The refresh interval must be between 15 minutes and 24 hours.
  srv_99e5a87eee94b355: 'Das Aktualisierungsintervall muss zwischen 15 Minuten und 24 Stunden liegen.',
  // At most 5 accounts can be connected.
  srv_4668b7fe22c08fca: 'Es können höchstens 5 Konten verbunden werden.',
  // Choose Google or Microsoft.
  srv_2fff1508d4ce6e8a: 'Wähle Google oder Microsoft.',
  // The account is not connected (any more).
  srv_0756151be3234993: 'Das Konto ist nicht (mehr) verbunden.',
  // The account needs to be signed in again.
  srv_124d16b34844185b: 'Das Konto muss neu angemeldet werden.',
  // No upcoming events
  srv_840e712c8a9ce8f9: 'Keine anstehenden Termine',
  // Now
  srv_3763f519c347a9f7: 'Jetzt',
  // On {date}
  srv_6e7f27cd5f252230: 'Am {date}',
  // On {date} at {time}
  srv_28a1ef3fa295aae0: 'Am {date} um {time}',
  // Today
  srv_63d040e37887f17e: 'Heute',
  // Today at {time}
  srv_3ae5f0c83c31dbca: 'Heute um {time}',
  // Tomorrow
  srv_44ef1526bf1d8106: 'Morgen',
  // Tomorrow at {time}
  srv_289edcea27d3bdd2: 'Morgen um {time}',
  // {title}
  srv_981efd4748b03a23: '{title}',
  // No calendar set up
  srv_fabbc2a837e8964d: 'Kein Kalender eingerichtet',
  // Choose Connect in the widget to connect Nextcloud or another calendar server, or add a Google or Microsoft account or a calendar link in Settings under Calendar.
  srv_63b92e4feacf176a:
    'Wähle Verbinden im Widget, um Nextcloud oder einen anderen Kalenderserver zu verbinden, oder füge in den Einstellungen unter Kalender ein Google- oder Microsoft-Konto oder einen Kalender-Link hinzu.',
  // Your appointments as an agenda, a month or both side by side.
  srv_f66c0a14f5ab7a89: 'Deine Termine als Liste, als Monat oder beides nebeneinander.',
};
