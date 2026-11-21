import type { en } from './en';

/** German UI texts of the attention feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'settings.section.attention': 'Benötigt deine Aufmerksamkeit',
  'attention.title': 'Benötigt deine Aufmerksamkeit',
  'attention.badge': '{n} Dinge brauchen deine Aufmerksamkeit',
  'attention.badge_one': '1 Sache braucht deine Aufmerksamkeit',
  'attention.intro':
    'Widgets und Verbindungen, die nicht mehr funktionieren oder noch nicht eingerichtet sind.',
  'attention.failing': 'Funktioniert nicht mehr',
  'attention.setup': 'Noch nicht eingerichtet',
  'attention.empty': 'Alles funktioniert. Nichts braucht deine Aufmerksamkeit.',
  'attention.loading': 'Wird geprüft …',
  'attention.since': 'seit {value}',
  'attention.check_all': 'Alles erneut prüfen',
  'attention.action.signIn': 'Neu anmelden',
  'attention.action.check': 'Erneut prüfen',
  'attention.action.fix': 'Beheben',
  'attention.action.choose': 'Einrichten',
  'attention.action.connect': 'Verbinden',
  'attention.on_tab': 'Auf dem Tab {value}',
  'attention.instance': 'Für die ganze Instanz',
};
