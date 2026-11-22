import type { ServerMessageKey } from '../../../api-client/generated/server-messages/notes';

/**
 * German translations of the backend error messages of `notes` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A note needs a title or some text.
  srv_cf47637ff80f03dc: 'Eine Notiz braucht einen Titel oder etwas Text.',
  // The note must be at most 100 KiB long.
  srv_4afce48c7ea72d13: 'Die Notiz darf höchstens 100 KiB lang sein.',
  // The version does not exist.
  srv_192c651a50b6f91f: 'Die Version existiert nicht.',
  // The note was changed elsewhere in the meantime.
  srv_4fc5f1e95d5f8139: 'Die Notiz wurde inzwischen an anderer Stelle geändert.',
  // One note or your pinned notes, written in Markdown.
  srv_521ab760c6e901b7: 'Eine Notiz oder deine angehefteten Notizen, geschrieben in Markdown.',
  // Type "n" and some text into the search to add a note, then pin it. Open the widget settings to choose which note it shows.
  srv_c90584c7159a5524:
    'Tippe „n“ und einen Text in die Suche, um eine Notiz anzulegen, und hefte sie an. In den Widget-Einstellungen wählst du, welche Notiz es zeigt.',
};
