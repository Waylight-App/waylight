import { EXPORT_OWNER, type ExportTables } from '../../../../shared/local/exportRows';
import type { TabRule } from '../../domain/dashboard';
import type { DashboardDoc } from './localDashboardGateway';

const minutes = (hhmm: string): number => {
  const [h = '0', m = '0'] = hhmm.split(':');
  return Number(h) * 60 + Number(m);
};

/**
 * A rule as the server stores it (adapter/gateway/rules.go): minutes since
 * midnight, and from = to for the whole day.
 */
function ruleRow(r: TabRule) {
  const whole = r.from === '' || r.from === r.to;
  return {
    days: r.days,
    from: whole ? 0 : minutes(r.from),
    to: whole ? 0 : minutes(r.to),
    networks: r.networks,
  };
}

/**
 * The dashboard document as rows of dashboard_tabs, dashboard_widgets and
 * dashboard_templates (migrations of internal/modules/dashboard). Tabs
 * belong to the owner and stay private, as on a server after setup.
 */
export function dashboardExport(doc: DashboardDoc): ExportTables {
  return {
    dashboard_tabs: doc.tabs.map((t) => ({
      id: t.id,
      name: t.name,
      position: t.position,
      background: t.background,
      rules: JSON.stringify(t.rules.map(ruleRow)),
      // NULL = the phone layout follows the desktop grid.
      mobile_layout: t.mobileCustom
        ? JSON.stringify(
            t.mobile.map((m) =>
              m.hidden ? { id: m.id, hidden: true, rows: m.rows } : { id: m.id, rows: m.rows },
            ),
          )
        : null,
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
    dashboard_widgets: doc.widgets.map((w) => ({
      id: w.id,
      tab_id: w.tabId,
      kind: w.kind,
      x: w.x,
      y: w.y,
      w: w.w,
      h: w.h,
      config: JSON.stringify(w.config),
    })),
    dashboard_templates: doc.templates.map((t) => ({
      id: t.id,
      name: t.name,
      widgets: JSON.stringify(t.widgets.map((w) => ({ kind: w.kind, ...w.rect, config: w.config }))),
    })),
  };
}
