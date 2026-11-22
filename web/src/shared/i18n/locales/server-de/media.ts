import type { ServerMessageKey } from '../../../api-client/generated/server-messages/media';

/**
 * German translations of the backend error messages of `media` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose Jellyfin or Plex.
  srv_52395ea57eb8771f: 'Wähle Jellyfin oder Plex.',
  // The access key or token is required.
  srv_c958491d891a6a18: 'Der Zugangsschlüssel oder das Zugangstoken ist erforderlich.',
  // The address must look like http://192.168.1.2:8096.
  srv_9d6bdf2942458cde: 'Die Adresse muss wie http://192.168.1.2:8096 aussehen.',
  // The media server is not configured yet.
  srv_7b276d2193d07bc8: 'Der Medienserver ist noch nicht eingerichtet.',
  // The media server is not reachable.
  srv_4c92147e0244306f: 'Der Medienserver ist nicht erreichbar.',
  // The poster does not exist.
  srv_2b05bbdb56e575c6: 'Das Cover existiert nicht.',
  // Choose Set up in the widget, pick Jellyfin or Plex and enter its address and key.
  srv_2c5a63cfd6db3056:
    'Wähle Einrichten im Widget, wähle Jellyfin oder Plex und gib Adresse und Schlüssel ein.',
  // What is playing on Jellyfin or Plex right now and what was added recently.
  srv_b8fdd06f02bd02db: 'Was gerade auf Jellyfin oder Plex läuft und was neu hinzugekommen ist.',
};
