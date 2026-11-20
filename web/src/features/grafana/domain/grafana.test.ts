import { describe, expect, it } from 'vitest';
import { hasPanel, panelLink, parsePanelConfig, renderUrl } from './grafana';

describe('grafana domain', () => {
  it('reads the panel settings with defaults', () => {
    const c = parsePanelConfig({ dashboardUid: 'node', panelId: 2, range: 'now-6h', refreshMinutes: 500 });
    expect(c).toMatchObject({ dashboardUid: 'node', panelId: 2, range: '6h', refreshMinutes: 60 });
    expect(hasPanel(c)).toBe(true);
    expect(hasPanel(parsePanelConfig({}))).toBe(false);
  });

  it('builds the render address and the link to Grafana', () => {
    const c = parsePanelConfig({
      dashboardUid: 'no de',
      panelId: 3,
      range: '7d',
      dashboardUrl: 'http://g/d/node/x',
    });
    expect(renderUrl(c, { width: 400.4, height: 200 }, true, 9)).toBe(
      '/api/grafana/render/no%20de/3?width=400&height=200&range=7d&theme=dark&t=9',
    );
    expect(panelLink(c)).toBe('http://g/d/node/x?viewPanel=3&from=now-7d&to=now');
    expect(panelLink(parsePanelConfig({ panelId: 3 }))).toBe('');
  });
});
