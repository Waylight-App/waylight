import type { en } from './en';

/** German UI texts of the grafana feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.grafana-panel.title': 'Grafana-Panel',
  'widget.grafana-panel.description':
    'Ein Panel eines Grafana-Dashboards als Bild oder ohne Image Renderer mit seinen aktuellen Werten.',
  'widget.grafana-alerts.title': 'Grafana-Alarme',
  'widget.grafana-alerts.description': 'Aktive und ausstehende Alarme von Grafana Alerting.',
  'settings.section.grafana': 'Grafana',
  'grafana.connect_under_settings': 'Verbinde Grafana unter Einstellungen → Grafana.',
  'grafana.url': 'Grafana-URL',
  'grafana.token': 'Token des Dienstkontos {value}',
  'grafana.token_hint':
    'Lege in Grafana unter Administration → Users and access → Service accounts ein Dienstkonto mit der Rolle Viewer an und erstelle dafür ein Token. Das Token wird verschlüsselt gespeichert; Panels und Alarme lädt Waylight.',
  'grafana.choose_panel': 'Wähle in den Widget-Einstellungen ein Dashboard und ein Panel.',
  'grafana.open_in_grafana': 'In Grafana öffnen',
  'grafana.panel_alt': 'Grafana-Panel {title}',
  'grafana.dashboard': 'Dashboard',
  'grafana.search_dashboards': 'Dashboards suchen',
  'grafana.panel': 'Panel',
  'grafana.range': 'Zeitraum (z. B. 6h, 7d)',
  'grafana.refresh_minutes': 'Alle … Minuten neu laden',
  'grafana.no_dashboards': 'Keine Dashboards gefunden.',
  'grafana.all_quiet': 'Kein Alarm ist aktiv.',
  'grafana.pending': 'ausstehend',
  'grafana.loading_values': 'Werte werden geladen…',
  'grafana.values_over': '{title} in den letzten {range}',
  'grafana.alert_list': 'Alarmliste',
};
