import type { ServerMessageKey } from '../../../api-client/generated/server-messages/nextcloud';

/**
 * German translations of the backend error messages of `nextcloud` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Nextcloud is not configured yet.
  srv_852391e96f8a11b2: 'Nextcloud ist noch nicht eingerichtet.',
  // Choose Connect in the widget and sign in with Nextcloud, or enter its address and an app password.
  srv_8916b9019aa8cb6c:
    'Wähle Verbinden im Widget und melde dich mit Nextcloud an oder gib die Adresse und ein App-Passwort ein.',
  // Your Nextcloud: version, your storage, and for admins free space, users and app updates.
  srv_94a13471047e5554:
    'Deine Nextcloud: Version, dein Speicher und für Admins freier Platz, Benutzer und App-Updates.',
};
