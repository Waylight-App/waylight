import type { ServerMessageKey } from '../../../api-client/generated/server-messages/adguard';

/**
 * German translations of the backend error messages of `adguard` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Username and password are required.
  srv_ed3a053401c9b56a: 'Benutzername und Passwort sind erforderlich.',
  // The DNS filter is not reachable.
  srv_4a327952321d8a3f: 'Der DNS-Filter ist nicht erreichbar.',
  // This DNS filter is not configured yet.
  srv_fcbb5bbe636e4559: 'Dieser DNS-Filter ist noch nicht eingerichtet.',
  // Unknown DNS filter, use adguard or pihole.
  srv_1a87cc7977ed4993: 'Unbekannter DNS-Filter, verwende adguard oder pihole.',
  // Pause for 5, 30 or 60 minutes.
  srv_0c0b1a4d7e09267c: 'Pausiere für 5, 30 oder 60 Minuten.',
  // Pausing is switched off for this DNS filter. Allow it with DNS_PAUSE_PROVIDERS on the server.
  srv_6af65a6080fe6b52:
    'Pausieren ist für diesen DNS-Filter ausgeschaltet. Erlaube es mit DNS_PAUSE_PROVIDERS auf dem Server.',
  // Blocking is off
  srv_71fcd40efaa8ff27: 'Blockieren ist aus',
  // Blocking paused
  srv_43efb7d828a46638: 'Blockieren pausiert',
  // The DNS filter is not reachable
  srv_c24326f62f32176b: 'Der DNS-Filter ist nicht erreichbar',
  // {count} of {total} queries
  srv_15a7ef0c4caf3d1d: '{count} von {total} Anfragen',
  // {percent} % blocked
  srv_f388457dcac53d65: '{percent} % blockiert',
  // Choose Set up in the widget and enter the address and password of your AdGuard Home or Pi-hole.
  srv_9403d5143205448f:
    'Wähle Einrichten im Widget und gib Adresse und Passwort deines AdGuard Home oder Pi-hole ein.',
  // How many requests AdGuard Home or Pi-hole blocked today, and whether blocking is on.
  srv_b49765dc7af314cb:
    'Wie viele Anfragen AdGuard Home oder Pi-hole heute blockiert hat und ob das Blockieren an ist.',
};
