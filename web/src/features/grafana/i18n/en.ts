/** English UI texts of the grafana feature (#574). */
export const en = {
  'widget.grafana-panel.title': 'Grafana panel',
  'widget.grafana-panel.description':
    'One panel of a Grafana dashboard as an image, or its current values without the image renderer.',
  'widget.grafana-alerts.title': 'Grafana alerts',
  'widget.grafana-alerts.description': 'Firing and pending alerts of Grafana Alerting.',
  'settings.section.grafana': 'Grafana',
  'grafana.connect_under_settings': 'Connect Grafana under Settings → Grafana.',
  'grafana.url': 'Grafana URL',
  'grafana.token': 'Service account token {value}',
  'grafana.token_hint':
    'In Grafana, add a service account with the role Viewer under Administration → Users and access → Service accounts, and give it a token. The token is stored encrypted; panels and alerts are loaded through Waylight.',
  'grafana.choose_panel': 'Choose a dashboard and a panel in the widget settings.',
  'grafana.open_in_grafana': 'Open in Grafana',
  'grafana.panel_alt': 'Grafana panel {title}',
  'grafana.dashboard': 'Dashboard',
  'grafana.search_dashboards': 'Search dashboards',
  'grafana.panel': 'Panel',
  'grafana.range': 'Time range (e.g. 6h, 7d)',
  'grafana.refresh_minutes': 'Refresh every … minutes',
  'grafana.no_dashboards': 'No dashboards found.',
  'grafana.all_quiet': 'No alert is firing.',
  'grafana.pending': 'pending',
  'grafana.loading_values': 'Loading values…',
  'grafana.values_over': '{title} over the last {range}',
  'grafana.alert_list': 'Alert list',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
