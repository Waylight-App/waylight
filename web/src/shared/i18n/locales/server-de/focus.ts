import type { ServerMessageKey } from '../../../api-client/generated/server-messages/focus';

/**
 * German translations of the backend error messages of `focus` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A session lasts between 1 and 180 minutes.
  srv_e9ad8cf134079d79: 'Eine Einheit dauert zwischen 1 und 180 Minuten.',
  // The session must have ended within the last day.
  srv_addb603b2e8d10d3: 'Die Einheit muss innerhalb des letzten Tages geendet haben.',
  // {count} sessions
  srv_74b44643f778ca81: '{count} Einheiten',
  // {minutes} min focus today
  srv_cf0e9452be96913b: '{minutes} Min. Fokus heute',
  // A long break follows every 1 to 12 sessions.
  srv_888252e2db1c2fca: 'Eine lange Pause folgt nach 1 bis 12 Einheiten.',
  // A phase lasts between 1 and 180 minutes.
  srv_0f85885a0017dc08: 'Eine Phase dauert zwischen 1 und 180 Minuten.',
  // Focus · {time}
  srv_562970061d64cba1: 'Fokus · {time}',
  // Long break · {time}
  srv_34384acbdc4b9f9a: 'Lange Pause · {time}',
  // Only a signed-in account has its own focus timer.
  srv_372b1075533617de: 'Nur ein angemeldetes Konto hat einen eigenen Fokus-Timer.',
  // Pause
  srv_061cc20e3432dead: 'Pause',
  // Paused · {time}
  srv_f896339e1166a799: 'Pausiert · {time}',
  // Short break · {time}
  srv_c9c9fdc4f2ddf844: 'Kurze Pause · {time}',
  // Start
  srv_ae1c7f13b7c0c47f: 'Start',
  // Unknown timer action.
  srv_4aeb14eb94ebdd1c: 'Unbekannte Timer-Aktion.',
  // A focus timer: work for a while, then take a short break.
  srv_68bf687ab1701e2d: 'Ein Fokus-Timer: eine Weile arbeiten, dann eine kurze Pause.',
  // Press start. Open the widget settings to change the work and break lengths.
  srv_44b22cd219681124:
    'Drück auf Start. In den Widget-Einstellungen änderst du die Länge von Arbeit und Pause.',
};
