import type { ServerMessageKey } from '../../../api-client/generated/server-messages/photos';

/** German translations of the backend messages of `photos` (#286). */
export const serverDe: Record<ServerMessageKey, string> = {
  // A picture must be at most 25 MB.
  srv_4959ed52f3185d86: 'Ein Bild darf höchstens 25 MB groß sein.',
  // The album name must have between 1 and 40 characters.
  srv_869ef2e39243a811: 'Der Albumname muss 1 bis 40 Zeichen lang sein.',
  // The order must list every picture of the album exactly once.
  srv_3a266f56f76134ba: 'Die Reihenfolge muss jedes Bild des Albums genau einmal enthalten.',
  // The photo storage is full (300 MB or 2000 pictures). Delete pictures first.
  srv_98c9fc1eaba891b7: 'Der Fotospeicher ist voll (300 MB oder 2000 Bilder). Lösche zuerst Bilder.',
  // The picture does not exist (any more).
  srv_164729afbb572b99: 'Das Bild gibt es nicht (mehr).',
  // Open the widget settings to upload pictures, choose the album and how long each picture stays.
  srv_63e5853951784541:
    'In den Widget-Einstellungen lädst du Bilder hoch und wählst das Album und wie lange jedes Bild bleibt.',
  // Your own pictures, one after another.
  srv_64f1f22cfb8682ac: 'Deine eigenen Bilder, eins nach dem anderen.',
};
