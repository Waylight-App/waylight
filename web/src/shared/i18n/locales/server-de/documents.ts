import type { ServerMessageKey } from '../../../api-client/generated/server-messages/documents';

/** German translations of the backend messages of `documents` (#102, #688). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Add a document with the product, the purchase day and the end of the warranty. Attach a photo or a PDF of up to 10 MB.
  srv_1d0b7fd7ea87e530:
    'Lege ein Dokument mit Produkt, Kaufdatum und Garantieende an. Hänge ein Foto oder ein PDF mit bis zu 10 MB an.',
  // Choose a photo or a PDF.
  srv_f7b4c9a6ad2131d1: 'Wähle ein Foto oder ein PDF.',
  // No warranties running
  srv_f2930534b26e4b01: 'Keine Garantien aktiv',
  // The document does not exist.
  srv_f1d6d68443a22ffa: 'Das Dokument existiert nicht.',
  // The document has no file.
  srv_d39f6f517d6727c4: 'Das Dokument hat keine Datei.',
  // The file must be at most 10 MB.
  srv_dec1bf3d590d0762: 'Die Datei darf höchstens 10 MB groß sein.',
  // The title must be between 1 and 80 characters long.
  srv_58daba2b260935be: 'Der Titel muss zwischen 1 und 80 Zeichen lang sein.',
  // The warranty cannot end before the purchase.
  srv_8adb3cb749b21e39: 'Die Garantie kann nicht vor dem Kauf enden.',
  // Your receipts and warranty papers, with the warranty status of each product.
  srv_7594c510b871f181: 'Deine Belege und Garantieunterlagen, mit dem Garantiestatus jedes Produkts.',
  // {name}: warranty ends {date}
  srv_a03ec9a433926804: '{name}: Garantie endet am {date}',
};
