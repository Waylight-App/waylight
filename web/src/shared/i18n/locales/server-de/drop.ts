import type { ServerMessageKey } from '../../../api-client/generated/server-messages/drop';

/**
 * German translations of the backend error messages of `drop` (#101). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose 1 hour, 24 hours or 7 days.
  srv_763ad0ee1e861149: 'Wähle 1 Stunde, 24 Stunden oder 7 Tage.',
  // Enter some text or a link.
  srv_992a3c76b72d1b0a: 'Gib einen Text oder einen Link ein.',
  // The drop zone is full (200 MB, 100 items). Delete items first.
  srv_b7ede1495726d216: 'Die Ablage ist voll (200 MB, 100 Einträge). Lösche zuerst etwas.',
  // The file must be at most 25 MB.
  srv_7fe59f42ecfd9d12: 'Die Datei darf höchstens 25 MB groß sein.',
  // The item does not exist or has expired.
  srv_05f4e0f511133638: 'Den Eintrag gibt es nicht oder er ist abgelaufen.',
  // The text must be at most 100 KB.
  srv_c4d9d288aedb26d3: 'Der Text darf höchstens 100 KB lang sein.',
  // Drop something into the widget on one device and open it on another. There is nothing to set up.
  srv_699407f4dd40994f:
    'Leg etwas auf einem Gerät ins Widget und öffne es auf einem anderen. Es ist nichts einzurichten.',
  // Text, links and files you send between your devices; they disappear after a while.
  srv_6c1dd18a11888a9a:
    'Texte, Links und Dateien, die du zwischen deinen Geräten schickst; sie verschwinden nach einer Weile.',
};
