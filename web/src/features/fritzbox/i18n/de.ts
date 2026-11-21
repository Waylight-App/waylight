import type { en } from './en';

/** German UI texts of the fritzbox feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.fritzbox_calls.title': 'FRITZ!Box-Anrufe',
  'widget.fritzbox_calls.description':
    'Letzte Anrufe der FRITZ!Box: alle, verpasste, eingehende oder ausgehende.',
  'fritzbox.show_uptime': 'Verbindungsdauer anzeigen',
  'fritzbox.show_ip': 'Externe IP-Adresse anzeigen',
  'fritzbox.show_rates': 'Aktuellen Down- und Upload anzeigen',
  'fritzbox.calls_filter': 'Anrufe',
  'fritzbox.filter.all': 'Alle Anrufe',
  'fritzbox.filter.missed': 'Verpasste',
  'fritzbox.filter.incoming': 'Eingehende',
  'fritzbox.filter.outgoing': 'Ausgehende',
  'fritzbox.calls_count': 'Anzahl der Anrufe (5–50)',
  'fritzbox.show_numbers': 'Nummern von Anrufenden ohne Namen anzeigen',
  'fritzbox.save': 'Speichern',
  'fritzbox.no_calls': 'Keine Anrufe.',
  'fritzbox.call.incoming': 'eingehend',
  'fritzbox.call.missed': 'verpasst',
  'fritzbox.call.outgoing': 'ausgehend',
  'fritzbox.call.rejected': 'abgewiesen',
  'fritzbox.call.active': 'aktiv',
  'fritzbox.setup_username': 'FRITZ!Box-Benutzer',
  'fritzbox.setup_password': 'Passwort',
  'fritzbox.setup_step_network':
    'Öffne die FRITZ!Box, gehe zu Heimnetz → Netzwerk → Netzwerkeinstellungen und setze die Haken bei Zugriff für Anwendungen zulassen und Statusinformationen über UPnP übertragen.',
  'fritzbox.setup_step_user':
    'Für Leitungsgeschwindigkeit und Anrufliste: Lege unter System → FRITZ!Box-Benutzer einen Benutzer Waylight mit dem Recht FRITZ!Box Einstellungen an.',
  'fritzbox.setup_step_user_calls':
    'Lege unter System → FRITZ!Box-Benutzer einen Benutzer Waylight mit den Rechten FRITZ!Box Einstellungen und Sprachnachrichten, Faxnachrichten, FRITZ!App Fon und Anrufliste an.',
  'fritzbox.setup_step_enter_optional': 'Trage Benutzer und Passwort unten ein oder lass beides leer.',
  'fritzbox.setup_step_enter': 'Trage Benutzer und Passwort unten ein.',
  'fritzbox.setup_online': 'Online, externe Adresse {ip}',
  'fritzbox.setup_offline': 'Die FRITZ!Box antwortet; die Internetverbindung ist getrennt.',
  'fritzbox.setup_calls': '{n} letzte Anrufe',
};
