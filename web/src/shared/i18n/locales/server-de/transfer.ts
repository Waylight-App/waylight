import type { ServerMessageKey } from '../../../api-client/generated/server-messages/transfer';

/**
 * German translations of the backend error messages of `transfer` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Export version {version} is not supported by this version of trisse-home.
  srv_43fe335aacf98b3c: 'Exportversion {version} wird von dieser trisse-home-Version nicht unterstützt.',
  // The data could not be imported: {detail}
  srv_5715790273b7248a: 'Die Daten konnten nicht importiert werden: {detail}',
  // The file is larger than 16 MB.
  srv_e7cb725c138134cc: 'Die Datei ist größer als 16 MB.',
  // The file is not valid JSON.
  srv_b8c878135ee917b1: 'Die Datei ist kein gültiges JSON.',
  // This is not a trisse-home export file.
  srv_d9a79f89c625afc7: 'Das ist keine trisse-home-Exportdatei.',
  // Config version {version} is not supported by this version of trisse-home.
  srv_44877fa04a91ddb0:
    'Die Konfigurationsversion {version} wird von dieser trisse-home-Version nicht unterstützt.',
  // Section {name}: {reason}
  srv_a8a0f7eee6c42ca3: 'Abschnitt {name}: {reason}',
  // The environment variable {name} is not set.
  srv_79e20ea037a8c0dd: 'Die Umgebungsvariable {name} ist nicht gesetzt.',
  // The file is larger than 2 MB.
  srv_3949050e0fe11ded: 'Die Datei ist größer als 2 MB.',
  // The file is not valid: {reason}
  srv_14afd68d009021d7: 'Die Datei ist ungültig: {reason}',
  // The file must be a YAML mapping with version and sections.
  srv_09e549de4d9f81f3: 'Die Datei muss eine YAML-Zuordnung mit version und Abschnitten sein.',
  // The section {name} appears twice.
  srv_5df61e32f2bec534: 'Der Abschnitt {name} kommt zweimal vor.',
  // The version must be a number.
  srv_1a0dd18b884c11a8: 'Die version muss eine Zahl sein.',
  // Unknown section {name} (or its module is disabled).
  srv_40b2f70963afe6eb: 'Unbekannter Abschnitt {name} (oder sein Modul ist deaktiviert).',
  // YAML aliases (*name) are not supported; write the value out.
  srv_131e8662bd896c29: 'YAML-Aliase (*name) werden nicht unterstützt; schreibe den Wert aus.',
  // Choose at least one file.
  srv_d4e32079da1f8a6c: 'Wähle mindestens eine Datei.',
  // The files are larger than 2 MB together.
  srv_b080b45242740908: 'Die Dateien sind zusammen größer als 2 MB.',
  // This is not a Homarr board file. Homarr 1.x exports cannot be read; use the board JSON of Homarr 0.x.
  srv_d39ee18784bab42a:
    'Das ist keine Homarr-Board-Datei. Exporte von Homarr 1.x lassen sich nicht lesen; nimm die Board-JSON von Homarr 0.x.',
  // Unknown import format.
  srv_91da20f0a1aee001: 'Unbekanntes Importformat.',
  // Upload one Homarr board file (JSON).
  srv_bf8a861f5dbe0e42: 'Lade eine Homarr-Board-Datei (JSON) hoch.',
  // Upload services.yaml, bookmarks.yaml or widgets.yaml from Homepage.
  srv_b395afa091b77fe4: 'Lade services.yaml, bookmarks.yaml oder widgets.yaml aus Homepage hoch.',
  // {name} is not valid YAML: {reason}
  srv_9f0de21ec1d3ede8: '{name} ist kein gültiges YAML: {reason}',
};
