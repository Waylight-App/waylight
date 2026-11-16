import type { ServerMessageKey } from '../../../api-client/generated/server-messages/search';

/**
 * German translations of the backend error messages of `search` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A shortcut with this keyword already exists.
  srv_76a5625367cbe997: 'Es gibt bereits ein Kürzel mit diesem Schlüsselwort.',
  // The URL must contain {query}.
  srv_1605df454fade766: 'Die URL muss {query} enthalten.',
  // The keyword must consist of 1 to 16 letters or digits.
  srv_d6e09729bc0d1f3b: 'Das Schlüsselwort muss aus 1 bis 16 Buchstaben oder Ziffern bestehen.',
  // The search query is too long.
  srv_657058c57c5faf46: 'Die Suchanfrage ist zu lang.',
  // The search query must not be empty.
  srv_b828d9f84cec079a: 'Die Suchanfrage darf nicht leer sein.',
  // The shortcut does not exist.
  srv_2b29e30433c7ad15: 'Das Kürzel existiert nicht.',
  // Unknown search provider.
  srv_dea670a71d2295ca: 'Unbekannte Suchmaschine.',
  // A prefix is 1 to 4 letters or digits.
  srv_74215ccd79a3ff66: 'Ein Präfix besteht aus 1 bis 4 Buchstaben oder Ziffern.',
  // Each prefix can be used only once.
  srv_fd6c518295d812e7: 'Jedes Präfix darf nur einmal vorkommen.',
  // Each shortcut keyword can be used only once.
  srv_ff5b12b63bbd2fcc: 'Jedes Kürzel darf nur einmal vorkommen.',
  // A keyword is 1 to 24 letters, digits, - or _, starting with a letter or digit.
  srv_7ac1113700bd84e6:
    'Ein Stichwort besteht aus 1 bis 24 Buchstaben, Ziffern, - oder _ und beginnt mit einem Buchstaben oder einer Ziffer.',
  // Each own keyword can be used only once.
  srv_7751ccb712b43df6: 'Jedes eigene Stichwort darf nur einmal vorkommen.',
  // The keyword does not exist.
  srv_fb77bf30efe31cac: 'Das Stichwort existiert nicht.',
  // The label can be at most 100 characters long.
  srv_a4db379e15a68e9e: 'Die Bezeichnung darf höchstens 100 Zeichen lang sein.',
  // The target is missing or too long.
  srv_3d9c86149fa849ea: 'Das Ziel fehlt oder ist zu lang.',
  // The target must be a service, bookmark, widget, web address or search.
  srv_201b9e864fd6c6c6:
    'Das Ziel muss ein Dienst, Lesezeichen, Widget, eine Webadresse oder eine Suche sein.',
  // The web address must start with http:// or https://.
  srv_8bd56c7034353aa9: 'Die Webadresse muss mit http:// oder https:// beginnen.',
  // Own keywords need a signed-in account.
  srv_2ce07ca7ab64ce07: 'Eigene Stichwörter brauchen ein angemeldetes Konto.',
  // This keyword is already in use.
  srv_315ea6f2515e5a80: 'Dieses Stichwort ist schon vergeben.',
  // “{word}” is one of your own keywords; choose another prefix or delete the keyword.
  srv_1f5c1d6f6e0c2771:
    '„{word}“ ist eines deiner eigenen Stichwörter; wähle ein anderes Präfix oder lösche das Stichwort.',
  // “{word}” is the quick capture prefix for bookmarks; choose another keyword or change the prefix.
  srv_c79d63d26819dd1a:
    '„{word}“ ist das Präfix zum schnellen Erfassen von Lesezeichen; wähle ein anderes Stichwort oder ändere das Präfix.',
  // “{word}” is the quick capture prefix for notes; choose another keyword or change the prefix.
  srv_8b4b0288428b14ce:
    '„{word}“ ist das Präfix zum schnellen Erfassen von Notizen; wähle ein anderes Stichwort oder ändere das Präfix.',
  // “{word}” is the quick capture prefix for to-dos; choose another keyword or change the prefix.
  srv_284a8e8993530e8b:
    '„{word}“ ist das Präfix zum schnellen Erfassen von Aufgaben; wähle ein anderes Stichwort oder ändere das Präfix.',
};
