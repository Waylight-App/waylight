import { describe, expect, it, vi } from 'vitest';
import { type GrafanaGateway, GrafanaUseCases } from './grafanaUseCases';

describe('GrafanaUseCases', () => {
  it('trims the connection and the search', async () => {
    const gateway: GrafanaGateway = {
      settings: vi.fn(),
      saveSettings: vi.fn().mockResolvedValue({ configured: true, url: 'http://g', hasToken: true }),
      alerts: vi.fn(),
      dashboards: vi.fn().mockResolvedValue([]),
      panels: vi.fn(),
    };
    const uc = new GrafanaUseCases(gateway);
    await uc.saveSettings(' http://g ', ' glsa_1 ');
    await uc.dashboards(' node ');
    expect(gateway.saveSettings).toHaveBeenCalledWith('http://g', 'glsa_1');
    expect(gateway.dashboards).toHaveBeenCalledWith('node');
  });
});
