import type { en } from './en';

/** German UI texts of trisse-home.yaml; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'transfer.config_title': 'Einstellungen als YAML',
  'transfer.config_intro':
    'Aussehen, Suche, deine Dienst-Kacheln, Lesezeichen und das Dashboard als eine lesbare Datei, für Git oder eine zweite Instanz. Ohne Geheimnisse; ${NAME} wird aus der Server-Umgebung gelesen.',
  'transfer.config_download': 'trisse-home.yaml herunterladen',
  'transfer.config_apply': 'trisse-home.yaml anwenden…',
  'transfer.config_confirm':
    'Jeder Abschnitt der Datei ersetzt diese Einstellungen vollständig; fehlende Abschnitte bleiben, wie sie sind. Fortfahren?',
};
