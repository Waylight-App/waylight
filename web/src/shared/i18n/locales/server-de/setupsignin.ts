import type { ServerMessageKey } from '../../../api-client/generated/server-messages/setupsignin';

/**
 * German translations of the backend error messages of `setupsignin` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Signing in with Home Assistant needs PUBLIC_URL, the address you open Waylight with. Create a long-lived access token instead.
  srv_182123d619d272ea:
    'Die Anmeldung bei Home Assistant braucht PUBLIC_URL, die Adresse, unter der du Waylight öffnest. Erstelle stattdessen ein langlebiges Zugriffstoken.',
  // The address must start with http:// or https://.
  // The service did not name a sign-in page. Create a key instead.
  srv_ce81cb900ee071d5: 'Der Dienst hat keine Anmeldeseite genannt. Erstelle stattdessen einen Schlüssel.',
  // The sign-in was cancelled. Sign in again or create a key instead.
  srv_d765b701fb776c4a:
    'Die Anmeldung wurde abgebrochen. Melde dich erneut an oder erstelle stattdessen einen Schlüssel.',
  // This service has no sign-in.
  srv_b4e2abb05e1e39c8: 'Für diesen Dienst gibt es keine Anmeldung.',
  // The code has expired. Show a new code and scan it again.
  srv_e56fe3c73c0a647e: 'Der Code ist abgelaufen. Zeige einen neuen Code an und scanne ihn erneut.',
  // The value must not be empty or longer than 4096 characters.
  srv_7368f0d8826fd013: 'Der Wert darf nicht leer oder länger als 4096 Zeichen sein.',
};
