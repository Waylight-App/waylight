import type { ServerMessageKey } from '../../../api-client/generated/server-messages/arr';

/**
 * German translations of the backend error messages of `arr` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A Sonarr, Radarr or Prowlarr instance does not answer.
  srv_80cd4d1600527c40: 'Eine Sonarr-, Radarr- oder Prowlarr-Instanz antwortet nicht.',
  // At most {max} instances are supported.
  srv_e087fafa20933ecf: 'Höchstens {max} Instanzen werden unterstützt.',
  // Choose Sonarr, Radarr or Prowlarr.
  srv_35314e09732db39d: 'Wähle Sonarr, Radarr oder Prowlarr.',
  // No Sonarr, Radarr or Prowlarr instance is configured yet.
  srv_e7b97389eba1653e: 'Noch keine Sonarr-, Radarr- oder Prowlarr-Instanz eingerichtet.',
  // The access key is required (Settings → General).
  srv_7cb14eb21ac56134: 'Der Zugangsschlüssel ist erforderlich (Settings → General).',
  // The address must look like http://192.168.1.2:8989.
  srv_3c6d93dc1a0cc653: 'Die Adresse muss wie http://192.168.1.2:8989 aussehen.',
  // The instance does not exist.
  srv_c8e8fb2032a453cc: 'Die Instanz existiert nicht.',
  // Choose Set up in the widget, pick the app and enter its address and key.
  srv_f1bc61893833e7cb: 'Wähle Einrichten im Widget, wähle die App und gib Adresse und Schlüssel ein.',
  // Upcoming episodes and films, downloads in progress and warnings from Sonarr, Radarr or Prowlarr.
  srv_2c28b0f9e10b3a43:
    'Kommende Folgen und Filme, laufende Downloads und Warnungen aus Sonarr, Radarr oder Prowlarr.',
};
