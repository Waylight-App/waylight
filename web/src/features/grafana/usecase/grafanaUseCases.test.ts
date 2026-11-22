import { describe, expect, it, vi } from 'vitest';
import { parsePanelConfig } from '../domain/grafana';
import { type GrafanaGateway, GrafanaUseCases } from './grafanaUseCases';

describe('GrafanaUseCases', () => {
  it('trims the connection and the search', async () => {
    const gateway: GrafanaGateway = {
      settings: vi.fn(),
      saveSettings: vi.fn().mockResolvedValue({ configured: true, url: 'http://g', hasToken: true }),
      alerts: vi.fn(),
      dashboards: vi.fn().mockResolvedValue([]),
      panels: vi.fn(),
      panelValues: vi.fn(),
    };
    const uc = new GrafanaUseCases(gateway);
    await uc.saveSettings(' http://g ', ' glsa_1 ');
    await uc.dashboards(' node ');
    expect(gateway.saveSettings).toHaveBeenCalledWith('http://g', 'glsa_1');
    expect(gateway.dashboards).toHaveBeenCalledWith('node', undefined);
  });

  it('asks for the values of the chosen panel (#579)', async () => {
    const values = { title: 'CPU', unit: 'percent', series: [] };
    const gateway = { panelValues: vi.fn().mockResolvedValue(values) } as unknown as GrafanaGateway;
    const uc = new GrafanaUseCases(gateway);
    expect(await uc.panelValues(parsePanelConfig({ dashboardUid: 'node', panelId: 7, range: '1d' }))).toBe(
      values,
    );
    expect(gateway.panelValues).toHaveBeenCalledWith('/api/grafana/data/node/7?range=1d');
  });
});
