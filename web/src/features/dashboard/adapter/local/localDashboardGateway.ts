import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import type { Dashboard, Tab, TabRule, TabTemplate, Widget, WidgetConfig } from '../../domain/dashboard';
import type { LayoutItem, Rect } from '../../domain/layout';
import type { WidgetSize } from '../../domain/sizes';
import { deriveMobile, resolveMobile, type MobileItem } from '../../domain/mobile';
import { defaultTabId } from '../../domain/schedule';
import type { DashboardGateway } from '../../usecase/ports';

/** A tab as stored, in the shape of the server's tabs table. */
interface TabRow {
  id: string;
  name: string;
  position: number;
  background: string;
  rules: TabRule[];
  mobile: MobileItem[];
  mobileCustom: boolean;
}

/** A widget as stored, in the shape of the server's widgets table. */
interface WidgetRow extends Rect {
  id: string;
  tabId: string;
  kind: string;
  config: WidgetConfig;
  size?: WidgetSize;
}

interface TemplateRow {
  id: string;
  name: string;
  widgets: { kind: string; rect: Rect; config: WidgetConfig }[];
}

/** The dashboard document of the local mode. */
export interface DashboardDoc {
  tabs: TabRow[];
  widgets: WidgetRow[];
  templates: TemplateRow[];
}

/**
 * Widgets of the first tab, as the server's sample layout for the start page
 * (internal/modules/dashboard/usecase/sample.go), limited to the kinds that
 * work without a server. Weather is left out because it needs a location.
 */
const firstTabWidgets: { kind: string; rect: Rect; config: WidgetConfig }[] = [
  { kind: 'clock', rect: { x: 0, y: 0, w: 3, h: 3 }, config: {} },
  { kind: 'todos', rect: { x: 3, y: 0, w: 5, h: 4 }, config: {} },
  { kind: 'notes', rect: { x: 8, y: 0, w: 4, h: 6 }, config: { scope: 'pinned' } },
  { kind: 'bookmarks', rect: { x: 3, y: 4, w: 5, h: 5 }, config: {} },
];

/** Name of the tab created on first use; the server uses the same (DefaultTabName). */
const FIRST_TAB = 'Home';

class NotFoundError extends Error {
  override name = 'NotFoundError';
}

/**
 * Dashboard gateway of the browser extension's local mode (ADR 0019): tabs,
 * widgets, layouts and templates in one document of the DocumentStore.
 * Validation stays in the use cases, as for the API gateway.
 */
export class LocalDashboardGateway implements DashboardGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<DashboardDoc> {
    return (await this.store.get<DashboardDoc>('dashboard')) ?? { tabs: [], widgets: [], templates: [] };
  }

  /** Reads, changes and saves the document in one step; returns what fn returns. */
  private async change<T>(fn: (doc: DashboardDoc) => T): Promise<T> {
    const doc = await this.read();
    const result = fn(doc);
    await this.store.put('dashboard', doc);
    return result;
  }

  private static tabRow(doc: DashboardDoc, id: string): TabRow {
    const tab = doc.tabs.find((t) => t.id === id);
    if (!tab) throw new NotFoundError(`tab ${id} not found`);
    return tab;
  }

  private static widgetRow(doc: DashboardDoc, id: string): WidgetRow {
    const widget = doc.widgets.find((w) => w.id === id);
    if (!widget) throw new NotFoundError(`widget ${id} not found`);
    return widget;
  }

  /** The tab as the API returns it: widgets attached, phone layout resolved. */
  private static toTab(doc: DashboardDoc, row: TabRow): Tab {
    const widgets: Widget[] = doc.widgets
      .filter((w) => w.tabId === row.id)
      .map((w) => ({
        id: w.id,
        tabId: w.tabId,
        kind: w.kind,
        config: w.config,
        x: w.x,
        y: w.y,
        w: w.w,
        h: w.h,
        ...(w.size ? { size: w.size } : {}),
      }));
    return {
      id: row.id,
      name: row.name,
      position: row.position,
      background: row.background,
      rules: row.rules,
      widgets,
      mobile: row.mobileCustom ? resolveMobile(row.mobile, widgets) : deriveMobile(widgets),
      mobileCustom: row.mobileCustom,
      mine: true,
    };
  }

  private static newTab(doc: DashboardDoc, name: string): TabRow {
    const row: TabRow = {
      id: newId(),
      name,
      position: doc.tabs.reduce((max, t) => Math.max(max, t.position + 1), 0),
      background: '',
      rules: [],
      mobile: [],
      mobileCustom: false,
    };
    doc.tabs.push(row);
    return row;
  }

  private static place(doc: DashboardDoc, tabId: string, items: TemplateRow['widgets']): void {
    for (const { kind, rect, config } of items) {
      doc.widgets.push({ id: newId(), tabId, kind, config: structuredClone(config), ...rect });
    }
  }

  async load(): Promise<Dashboard> {
    let doc = await this.read();
    if (doc.tabs.length === 0) {
      // First start, as after the server's setup: one tab to begin with.
      doc = await this.change((d) => {
        const tab = LocalDashboardGateway.newTab(d, FIRST_TAB);
        LocalDashboardGateway.place(d, tab.id, firstTabWidgets);
        return d;
      });
    }
    const tabs = [...doc.tabs]
      .sort((a, b) => a.position - b.position)
      .map((t) => LocalDashboardGateway.toTab(doc, t));
    return { tabs, defaultTabId: defaultTabId(tabs, this.now()) };
  }

  createTab(name: string): Promise<Tab> {
    return this.change((d) => LocalDashboardGateway.toTab(d, LocalDashboardGateway.newTab(d, name)));
  }

  renameTab(id: string, name: string): Promise<Tab> {
    return this.change((d) => {
      const tab = LocalDashboardGateway.tabRow(d, id);
      tab.name = name;
      return LocalDashboardGateway.toTab(d, tab);
    });
  }

  async deleteTab(id: string): Promise<void> {
    await this.change((d) => {
      LocalDashboardGateway.tabRow(d, id);
      d.tabs = d.tabs.filter((t) => t.id !== id);
      d.widgets = d.widgets.filter((w) => w.tabId !== id);
    });
  }

  async reorderTabs(ids: string[]): Promise<void> {
    await this.change((d) => {
      ids.forEach((id, position) => {
        LocalDashboardGateway.tabRow(d, id).position = position;
      });
    });
  }

  addWidget(
    tabId: string,
    kind: string,
    rect: Rect,
    config: WidgetConfig,
    size?: WidgetSize,
  ): Promise<Widget> {
    return this.change((d) => {
      LocalDashboardGateway.tabRow(d, tabId);
      const sized = size ? { size } : {};
      const row: WidgetRow = { id: newId(), tabId, kind, config, ...rect, ...sized };
      d.widgets.push(row);
      return { id: row.id, tabId, kind, config, ...rect, ...sized };
    });
  }

  configureWidget(id: string, config: WidgetConfig): Promise<Widget> {
    return this.change((d) => {
      const w = LocalDashboardGateway.widgetRow(d, id);
      w.config = config;
      return {
        id: w.id,
        tabId: w.tabId,
        kind: w.kind,
        config: w.config,
        x: w.x,
        y: w.y,
        w: w.w,
        h: w.h,
        ...(w.size ? { size: w.size } : {}),
      };
    });
  }

  async saveLayout(tabId: string, items: LayoutItem[]): Promise<void> {
    await this.change((d) => {
      LocalDashboardGateway.tabRow(d, tabId);
      for (const item of items) {
        const w = LocalDashboardGateway.widgetRow(d, item.id);
        if (w.tabId !== tabId) throw new NotFoundError(`widget ${item.id} is not on tab ${tabId}`);
        Object.assign(w, { x: item.x, y: item.y, w: item.w, h: item.h });
        if (item.size) w.size = item.size;
        else delete w.size;
      }
    });
  }

  async deleteWidget(id: string): Promise<void> {
    await this.change((d) => {
      LocalDashboardGateway.widgetRow(d, id);
      d.widgets = d.widgets.filter((w) => w.id !== id);
    });
  }

  setBackground(tabId: string, background: string): Promise<Tab> {
    return this.change((d) => {
      const tab = LocalDashboardGateway.tabRow(d, tabId);
      tab.background = background;
      return LocalDashboardGateway.toTab(d, tab);
    });
  }

  setRules(tabId: string, rules: TabRule[]): Promise<Tab> {
    return this.change((d) => {
      const tab = LocalDashboardGateway.tabRow(d, tabId);
      tab.rules = rules;
      return LocalDashboardGateway.toTab(d, tab);
    });
  }

  saveMobileLayout(tabId: string, items: MobileItem[]): Promise<Tab> {
    return this.change((d) => {
      const tab = LocalDashboardGateway.tabRow(d, tabId);
      tab.mobile = items;
      tab.mobileCustom = true;
      return LocalDashboardGateway.toTab(d, tab);
    });
  }

  resetMobileLayout(tabId: string): Promise<Tab> {
    return this.change((d) => {
      const tab = LocalDashboardGateway.tabRow(d, tabId);
      tab.mobile = [];
      tab.mobileCustom = false;
      return LocalDashboardGateway.toTab(d, tab);
    });
  }

  async listTemplates(): Promise<TabTemplate[]> {
    const doc = await this.read();
    return doc.templates.map((t) => ({
      id: t.id,
      name: t.name,
      widgets: t.widgets.length,
      kinds: [...new Set(t.widgets.map((w) => w.kind))],
    }));
  }

  saveTemplate(tabId: string, name: string): Promise<TabTemplate> {
    return this.change((d) => {
      LocalDashboardGateway.tabRow(d, tabId);
      const widgets = d.widgets
        .filter((w) => w.tabId === tabId)
        .map((w) => ({
          kind: w.kind,
          rect: { x: w.x, y: w.y, w: w.w, h: w.h },
          config: structuredClone(w.config),
        }));
      const row: TemplateRow = { id: newId(), name, widgets };
      d.templates.push(row);
      return { id: row.id, name, widgets: widgets.length, kinds: [...new Set(widgets.map((w) => w.kind))] };
    });
  }

  async deleteTemplate(id: string): Promise<void> {
    await this.change((d) => {
      if (!d.templates.some((t) => t.id === id)) throw new NotFoundError(`template ${id} not found`);
      d.templates = d.templates.filter((t) => t.id !== id);
    });
  }

  createTabFromTemplate(templateId: string, name: string): Promise<Tab> {
    return this.change((d) => {
      const template = d.templates.find((t) => t.id === templateId);
      if (!template) throw new NotFoundError(`template ${templateId} not found`);
      const tab = LocalDashboardGateway.newTab(d, name);
      LocalDashboardGateway.place(d, tab.id, template.widgets);
      return LocalDashboardGateway.toTab(d, tab);
    });
  }
}
