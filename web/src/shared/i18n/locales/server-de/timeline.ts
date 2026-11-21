import type { ServerMessageKey } from '../../../api-client/generated/server-messages/timeline';

/**
 * German translations of the backend error messages of `timeline` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The retention must be between 1 and 365 days.
  srv_2367e29c6edb5841: 'Die Aufbewahrungsdauer muss zwischen 1 und 365 Tagen liegen.',
  // Nothing to set up; events appear as the other widgets report them.
  srv_264ed58603336ca1: 'Nichts einzurichten; Ereignisse erscheinen, sobald die anderen Widgets sie melden.',
  // What happened when: outages, recoveries, failed backups and security events.
  srv_500c79b769d06a25:
    'Was wann passiert ist: Ausfälle, Wiederherstellungen, fehlgeschlagene Backups und Sicherheitsereignisse.',
};
