import type { ServerMessageKey } from '../../../api-client/generated/server-messages/immich';

/**
 * German translations of the backend error messages of `immich` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Immich is not configured yet.
  srv_b1e270c3a3b1cc7b: 'Immich ist noch nicht eingerichtet.',
  // Immich is not reachable.
  srv_5f51c1adae96a94e: 'Immich ist nicht erreichbar.',
  // The access key is required.
  srv_41f4f2e49617b26a: 'Der Zugangsschlüssel ist erforderlich.',
  // The address must look like http://192.168.1.2:2283.
  srv_e7b4706fb601581c: 'Die Adresse muss wie http://192.168.1.2:2283 aussehen.',
  // The thumbnail does not exist.
  srv_6493a23b37be9677: 'Das Vorschaubild existiert nicht.',
  // Choose Connect in the widget and enter the address of Immich and a key you create in its settings.
  srv_53abe4db239639c0:
    'Wähle Verbinden im Widget und gib die Adresse von Immich und einen Schlüssel ein, den du in dessen Einstellungen erstellst.',
  // Your Immich photo library and the photos taken on this day in past years.
  srv_96f0e81eb11e30d2:
    'Deine Immich-Fotobibliothek und die Fotos, die an diesem Tag in früheren Jahren entstanden sind.',
};
