import type { ServerMessageKey } from './generated/server-messages';

/**
 * German translations of the backend messages of the ee/ modules (#632).
 * The keys come from the generated catalogue next to it; the type makes a
 * missing translation a compile error. The plugin hands them to the core,
 * which translates server errors and compact states with them. The English
 * original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A user with this userName exists.
  srv_fbe42a65e3ad4943: 'Es gibt schon einen Benutzer mit diesem userName.',
  // An absence can span at most a year.
  srv_c053d6d39505436f: 'Eine Abwesenheit kann höchstens ein Jahr dauern.',
  // An admin has to register Waylight at this service first.
  srv_8ca1f3ecca038f78: 'Ein Admin muss Waylight zuerst bei diesem Dienst registrieren.',
  // Ask for at most 62 days at once.
  srv_c8357fa440121c4e: 'Frage höchstens 62 Tage auf einmal ab.',
  // Away
  srv_ececb48b4688a29b: 'Weg',
  // Choose a default theme before keeping members on it.
  srv_dcb4a6dea2ab1da6: 'Wähle ein Standarddesign, bevor du Mitglieder darauf festlegst.',
  // Choose holiday, sick, remote or other.
  srv_993d4e46bd0e7e9f: 'Wähle Urlaub, krank, remote oder anderes.',
  // Choose office, home or away.
  srv_d867c25dd03cadf4: 'Wähle Büro, Homeoffice oder weg.',
  // Choose one of your tabs that is shared with this group or with everyone.
  srv_272397481f0f574d: 'Wähle einen deiner Tabs, der mit dieser Gruppe oder mit allen geteilt ist.',
  // Connected accounts need a valid licence key.
  srv_a8fc21a7f9032445: 'Verbundene Konten brauchen einen gültigen Lizenzschlüssel.',
  // Connecting accounts needs PUBLIC_URL, so the service can send the browser back.
  srv_4146e68bc57de7b3:
    'Zum Verbinden von Konten muss PUBLIC_URL gesetzt sein, damit der Dienst den Browser zurückschicken kann.',
  // Enter a last day as YYYY-MM-DD, not before the first.
  srv_b10bffa646d2fb1d: 'Gib einen letzten Tag als JJJJ-MM-TT ein, nicht vor dem ersten.',
  // Enter the accent colour as #rrggbb.
  srv_fbda0bc2e50c52d7: 'Gib die Akzentfarbe als #rrggbb ein.',
  // Enter the day as YYYY-MM-DD.
  srv_406e0543f9cb2f5e: 'Gib den Tag als JJJJ-MM-TT ein.',
  // Enter the first day as YYYY-MM-DD.
  srv_b60b734a3a57d4a4: 'Gib den ersten Tag als JJJJ-MM-TT ein.',
  // Enter the last day as YYYY-MM-DD.
  srv_5782d5d3e8d55376: 'Gib den letzten Tag als JJJJ-MM-TT ein.',
  // Enter the secondary colour as #rrggbb.
  srv_b8832da637636f2f: 'Gib die Zweitfarbe als #rrggbb ein.',
  // Guests, kiosk displays and API tokens cannot set a status.
  srv_f7a2af3f0d1e9e65: 'Gäste, Kiosk-Anzeigen und API-Tokens können keinen Status setzen.',
  // Home
  srv_6abb1cd87fe0114e: 'Homeoffice',
  // My status today
  srv_a954440392d2ad7a: 'Mein Status heute',
  // Office
  srv_30b4e6b54a651371: 'Büro',
  // Only admins can switch presence on or off.
  srv_841ffd577f230161: 'Nur Admins können die Anwesenheit ein- oder ausschalten.',
  // Only an admin can change the branding.
  srv_c42246af1bc37d79: 'Nur ein Admin kann das Branding ändern.',
  // Only an admin can read the audit log.
  srv_c5478c6a183f80df: 'Nur ein Admin kann das Audit-Log lesen.',
  // Only an admin can register services.
  srv_4b6c91d5e9a17e7a: 'Nur ein Admin kann Dienste registrieren.',
  // Only an admin can see or change the licence key.
  srv_46b68e3594501562: 'Nur ein Admin kann den Lizenzschlüssel sehen oder ändern.',
  // Only an admin can see or change the policies.
  srv_5355c91166b95e94: 'Nur ein Admin kann die Richtlinien sehen oder ändern.',
  // Only members and admins can mark announcements as read.
  srv_a9505e852d6c028c: 'Nur Mitglieder und Admins können Ankündigungen als gelesen markieren.',
  // Only members and admins share calendars.
  srv_fbe0ca8945677a77: 'Nur Mitglieder und Admins teilen Kalender.',
  // Posting to Slack or Teams needs a licence key that includes chat apps.
  srv_ee68ac02d831018b: 'Für Beiträge in Slack oder Teams braucht es einen Lizenzschlüssel mit Chat-Apps.',
  // Presence is switched off for this organisation.
  srv_ccd005dcc1a12e3d: 'Anwesenheit ist für diese Organisation ausgeschaltet.',
  // The SCIM token is missing or wrong.
  srv_c200acf86457cbad: 'Das SCIM-Token fehlt oder ist falsch.',
  // The absence does not exist.
  srv_aaa52af61ff9db08: 'Die Abwesenheit gibt es nicht.',
  // The announcement does not exist.
  srv_61500043207d04a0: 'Die Ankündigung gibt es nicht.',
  // The connection has expired; connect the account again.
  srv_8f47f815e520e332: 'Die Verbindung ist abgelaufen; verbinde das Konto erneut.',
  // The default theme must be system, light or dark.
  srv_8fd8e3ee9c617dc4: 'Das Standarddesign muss System, hell oder dunkel sein.',
  // The last day must not be before the first.
  srv_eafd00928d11993f: 'Der letzte Tag darf nicht vor dem ersten liegen.',
  // The licence key is set with the environment variable LICENCE_KEY and cannot be changed here.
  srv_0fa482e06d8aa270:
    'Der Lizenzschlüssel ist über die Umgebungsvariable LICENCE_KEY gesetzt und kann hier nicht geändert werden.',
  // The minimum password length must be 0 or 12 to 64, the session length 0 to 2160 hours.
  srv_1662811f9114ef0c:
    'Die Mindestlänge für Passwörter muss 0 oder 12 bis 64 sein, die Sitzungsdauer 0 bis 2160 Stunden.',
  // The name must be at most 40 characters.
  srv_6059f25194cc7f75: 'Der Name darf höchstens 40 Zeichen haben.',
  // The note must be at most 200 characters.
  srv_1313ef79de25c7d6: 'Die Notiz darf höchstens 200 Zeichen haben.',
  // The retention must be 0 (keep for good) or between 30 and 3650 days.
  srv_316f54115293eab2: 'Die Aufbewahrung muss 0 (für immer behalten) oder zwischen 30 und 3650 Tagen sein.',
  // The return day must be a day after today, as YYYY-MM-DD.
  srv_9b204365551415ba: 'Der Tag der Rückkehr muss nach heute liegen, als JJJJ-MM-TT.',
  // The service did not confirm the sign-in: {detail}
  srv_8f61d806dfae22da: 'Der Dienst hat die Anmeldung nicht bestätigt: {detail}',
  // The sign-in has expired or does not belong to you; start again.
  srv_4ffceff61b3926db: 'Die Anmeldung ist abgelaufen oder gehört nicht zu dir; fang von vorn an.',
  // The text must be at most 10,000 characters.
  srv_8acb22e220ab9430: 'Der Text darf höchstens 10.000 Zeichen haben.',
  // The title must be 1 to 120 characters.
  srv_814a240ae1d970eb: 'Der Titel muss 1 bis 120 Zeichen haben.',
  // There are already 50 current announcements; let some expire or delete them first.
  srv_3bbad27681207c66:
    'Es gibt schon 50 aktuelle Ankündigungen; lass einige ablaufen oder lösche sie zuerst.',
  // This PATCH operation is not supported.
  srv_a119215abd68edcd: 'Diese PATCH-Operation wird nicht unterstützt.',
  // This account does not exist.
  srv_e292561bef8e02f5: 'Dieses Konto gibt es nicht.',
  // This account is not connected.
  srv_0e2f844b6efca1f2: 'Dieses Konto ist nicht verbunden.',
  // This group does not exist.
  srv_7ddddd9c390cd9a1: 'Diese Gruppe gibt es nicht.',
  // This licence key is not valid. Paste it exactly as you received it.
  srv_e8c180a3d732a86d:
    'Dieser Lizenzschlüssel ist ungültig. Füge ihn genau so ein, wie du ihn bekommen hast.',
  // This service is not known.
  srv_ffec4a0bca44886e: 'Dieser Dienst ist nicht bekannt.',
  // Upload a PNG, JPEG or GIF of at most 2 MB; SVG is not accepted.
  srv_c7ca86deca0949d2: 'Lade ein PNG, JPEG oder GIF mit höchstens 2 MB hoch; SVG wird nicht angenommen.',
  // Upload a logo, a dark logo or a favicon.
  srv_6d4dbe9e11e463b2: 'Lade ein Logo, ein dunkles Logo oder ein Favicon hoch.',
  // You can only enter and change your own absences.
  srv_6f3071f3338510fa: 'Du kannst nur deine eigenen Abwesenheiten eintragen und ändern.',
  // You may not write announcements.
  srv_0786b209ec0f5bc7: 'Du darfst keine Ankündigungen schreiben.',
  // from and to must be dates (YYYY-MM-DD) or times (RFC 3339), limit a number.
  srv_feb744876c8c61f1: 'from und to müssen Daten (JJJJ-MM-TT) oder Zeiten (RFC 3339) sein, limit eine Zahl.',
  // userName is required.
  srv_2ef8b59086fad102: 'userName ist erforderlich.',
  // {count} away
  srv_02edf4cd895e64fe: '{count} weg',
  // {office} in the office · {home} at home
  srv_4a4b587b0d9b29b5: '{office} im Büro · {home} im Homeoffice',
};
