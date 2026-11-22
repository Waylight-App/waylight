import type { ServerMessageKey } from '../../../api-client/generated/server-messages/speedtest';

/**
 * German translations of the backend error messages of `speedtest` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // An access token with the results:read scope is required.
  srv_9149e8b922088f89: 'Ein Zugangstoken mit dem Bereich results:read ist erforderlich.',
  // Speedtest Tracker is not configured yet.
  srv_6f32001a15a15c4d: 'Speedtest Tracker ist noch nicht eingerichtet.',
  // Speedtest Tracker is not reachable.
  srv_cfc88a9871753350: 'Speedtest Tracker ist nicht erreichbar.',
  // Thresholds must be between 0 and 100000 Mbit/s.
  srv_eb4a741553ab9379: 'Schwellenwerte müssen zwischen 0 und 100000 Mbit/s liegen.',
  // No results yet
  srv_a61c7e4a9008bf06: 'Noch keine Ergebnisse',
  // Ping {ping} ms
  srv_62fe3060b9aa0039: 'Ping {ping} ms',
  // Speedtest Tracker is not reachable
  srv_91cee62ab90d375e: 'Speedtest Tracker ist nicht erreichbar',
  // {down} ↓ · {up} ↑ Mbit/s
  srv_4a2ff384785102fb: '{down} ↓ · {up} ↑ Mbit/s',
  // Choose Connect in the widget and enter the address of your Speedtest Tracker and its access key.
  srv_34deaa4e80aaf9b7:
    'Wähle Verbinden im Widget und gib die Adresse deines Speedtest Tracker und seinen Zugangsschlüssel ein.',
  // Your internet speed over time, from Speedtest Tracker.
  srv_44125c07ce6c59fa: 'Deine Internetgeschwindigkeit im Verlauf, aus Speedtest Tracker.',
};
