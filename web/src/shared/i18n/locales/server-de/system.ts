import type { ServerMessageKey } from '../../../api-client/generated/server-messages/system';

/**
 * German translations of the backend error messages of `system` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Host metrics are not available. Are /proc and /sys mounted?
  srv_b854584a505d6c89: 'Host-Messwerte sind nicht verfügbar. Sind /proc und /sys eingebunden?',
  // The forecast warning must be between 1 and 365 days.
  srv_4a2cd38168dace16: 'Die Prognosewarnung muss zwischen 1 und 365 Tagen liegen.',
  // The forecast window must be between 3 and 90 days.
  srv_57fd7519663b39c9: 'Das Prognosefenster muss zwischen 3 und 90 Tagen liegen.',
  // How busy the server is: processor, memory, disks, temperature and how long it has been running.
  srv_593e88b1242b6e90:
    'Wie ausgelastet der Server ist: Prozessor, Arbeitsspeicher, Festplatten, Temperatur und wie lange er schon läuft.',
  // Nothing to set up; it shows the machine Waylight runs on.
  srv_a3f31a0950eeec3a: 'Nichts einzurichten; es zeigt den Rechner, auf dem Waylight läuft.',
};
