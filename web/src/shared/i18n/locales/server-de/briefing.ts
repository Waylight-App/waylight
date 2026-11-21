import type { ServerMessageKey } from '../../../api-client/generated/server-messages/briefing';

/**
 * German translations of the backend error messages of `briefing` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Latitude must be between -90 and 90, longitude between -180 and 180.
  srv_7ef0897648a1c97c:
    'Der Breitengrad muss zwischen -90 und 90 liegen, der Längengrad zwischen -180 und 180.',
  // Notifications are not enabled.
  srv_777fe9a31de38b1d: 'Benachrichtigungen sind nicht aktiviert.',
  // A short morning summary: weather warnings, problems overnight, due to-dos, appointments and waste collection.
  srv_6f6a25b9fc67ac66:
    'Eine kurze Morgenübersicht: Wetterwarnungen, Probleme der Nacht, fällige Aufgaben, Termine und Müllabfuhr.',
  // It collects what the other widgets know. In Settings under Daily briefing you can have it sent to you each morning.
  srv_08b4856a2d3da6d6:
    'Es sammelt, was die anderen Widgets wissen. In den Einstellungen unter Tages-Briefing kannst du es dir jeden Morgen schicken lassen.',
};
