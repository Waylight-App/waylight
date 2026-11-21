import { describe, expect, it } from 'vitest';
import { formatPanelValue, hasPanel, latest, panelLink, parsePanelConfig, renderUrl } from './grafana';

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

  it('formats panel values with their unit (#579)', () => {
    expect(formatPanelValue(42.123, 'percent', 1, 'en')).toBe('42.1 %');
    expect(formatPanelValue(0.5, 'percentunit', undefined, 'en')).toBe('50 %');
    expect(formatPanelValue(1536, 'bytes', 1, 'en')).toBe('1.5 KB');
    expect(formatPanelValue(21.5, 'celsius', undefined, 'de')).toBe('21,5 °C');
    expect(formatPanelValue(3, 'suffix: req', undefined, 'en')).toBe('3 req');
    expect(formatPanelValue(7, 'prefix:€', undefined, 'en')).toBe('€7');
    expect(formatPanelValue(1234.5678, 'unknown-unit', undefined, 'en')).toBe('1,234.6');
    expect(formatPanelValue(0.0123, '', undefined, 'en')).toBe('0.01');
  });

  it('takes the newest value of a series', () => {
    expect(
      latest({
        name: 'a',
        points: [
          { t: null, v: 1 },
          { t: null, v: 2 },
        ],
      }),
    ).toBe(2);
    expect(latest({ name: 'a', points: [] })).toBeNull();
    expect(latest(undefined)).toBeNull();
  });
});
