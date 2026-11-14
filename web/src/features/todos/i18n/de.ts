import type { en } from './en';

/** German UI texts of connected to-do lists (#169). */
export const de: Record<keyof typeof en, string> = {
  'settings.section.todos': 'To-do-Listen',
  'todos.changed_elsewhere':
    'Dieses To-do wurde inzwischen in der anderen App geändert. Jetzt wird dessen Stand angezeigt.',
  'todos.sources.intro':
    'Zeige eine Aufgabenliste aus Todoist, Nextcloud Tasks, Synology, Radicale, Baïkal, Fastmail oder einem anderen CalDAV-Server in einem To-do-Widget. Änderungen gehen in beide Richtungen; das Widget prüft jede Minute auf Änderungen.',
  'todos.sources.connected': 'Verbundene Listen',
  'todos.sources.none': 'Noch keine Liste verbunden.',
  'todos.sources.add': 'Liste verbinden',
  'todos.sources.name': 'Name in trisse-home',
  'todos.sources.url': 'Adresse des Servers oder der Liste',
  'todos.sources.url_hint':
    'Die DAV-Adresse des Servers reicht, z. B. https://cloud.example.com/remote.php/dav – dann unten die Liste wählen.',
  'todos.sources.username': 'Benutzername',
  'todos.sources.password': 'App-Passwort',
  'todos.sources.find': 'Listen suchen',
  'todos.sources.found': 'Aufgabenliste',
  'todos.sources.none_found': 'Unter dieser Adresse wurden keine Aufgabenlisten gefunden.',
  'todos.sources.connect': 'Verbinden',
  'todos.sources.remove': '{name} trennen',
  'todos.sources.removed': 'Getrennt. Auf dem Server wurde nichts gelöscht.',
  'todos.sources.saved': 'Verbunden und gespeichert.',
  'todos.sources.limits':
    'Wiederholte To-dos, Erinnerungen und eigene Reihenfolge gibt es nur in der eingebauten Liste. Fälligkeiten werden ohne Uhrzeit gespeichert.',
  'todos.sources.url_needed': 'Gib die Adresse des Servers oder der Liste ein (https://…).',
  'todos.sources.login_needed': 'Benutzername und App-Passwort sind erforderlich.',
  'todos.sources.name_needed': 'Gib der Liste einen Namen.',
  'todos.sources.list': 'Liste',
  'todos.sources.kind': 'Dienst',
  'todos.sources.kind_caldav': 'CalDAV (Nextcloud, Synology, Radicale, …)',
  'todos.sources.kind_todoist': 'Todoist',
  'todos.sources.token': 'API-Token',
  'todos.sources.token_hint':
    'In Todoist: Einstellungen → Integrationen → Entwickler → API-Token kopieren. Dann unten die Projekte suchen.',
  'todos.sources.token_needed': 'Gib das Todoist-API-Token ein.',
  'todos.sources.project': 'Projekt',
  'todos.sources.project_needed': 'Suche deine Projekte und wähle eines.',
  'todos.sources.limits_todoist':
    'Todoist wird höchstens alle 15 Sekunden nach Änderungen gefragt. Wird eine wiederkehrende Todoist-Aufgabe erledigt, springt sie wie in den Todoist-Apps auf ihren nächsten Termin.',
  'todos.sources.local': 'Eingebaute Liste',
  'todos.sources.missing': 'Getrennte Liste',
  'todos.sources.widget_hint': 'Listen verbindest du unter Einstellungen → To-do-Listen.',
};
