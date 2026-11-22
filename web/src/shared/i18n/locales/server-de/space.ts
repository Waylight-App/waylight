import type { ServerMessageKey } from '../../../api-client/generated/server-messages/space';

/**
 * German translations of the backend error messages of `space` (#64). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The NASA access key looks wrong.
  srv_0843f18316a55817: 'Der NASA-Zugangsschlüssel sieht falsch aus.',
  // The picture is not hosted by NASA.
  srv_374c6d77e754cb26: 'Das Bild liegt nicht bei der NASA.',
  // The space data is not available right now.
  srv_21a524286ae6e520: 'Die Raumfahrtdaten sind gerade nicht verfügbar.',
  // This space source is switched off in the settings.
  srv_1b5d999910a8dd39: 'Diese Raumfahrt-Quelle ist in den Einstellungen ausgeschaltet.',
  // Today's picture is not an image.
  srv_9825ec9eef300ecf: 'Das heutige Bild ist kein Bild (z. B. ein Video).',
  // The next rocket launches and NASA's astronomy picture of the day.
  srv_7f04b43fbc8ab518: 'Die nächsten Raketenstarts und das Astronomiebild des Tages der NASA.',
  // Turn on launches or the picture in Settings under Space; nothing is fetched until then.
  srv_d4b141b4193f3bcc:
    'Schalte Starts oder das Bild in den Einstellungen unter Raumfahrt ein; bis dahin wird nichts abgerufen.',
};
