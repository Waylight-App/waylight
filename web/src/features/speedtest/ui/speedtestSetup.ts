import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import { formatMbps } from '../domain/speedtest';
import type { SpeedtestUseCases } from '../usecase/speedtestUseCases';

/** The setup assistant of the speedtest widget (#618), as data. */
export function speedtestSetup(speedtest: SpeedtestUseCases): SetupDescriptor {
  return {
    service: 'Speedtest Tracker',
    fields: [
      {
        key: 'url',
        label: 'setup.address',
        stage: 'address',
        type: 'url',
        placeholder: 'http://192.168.178.20:8765',
      },
      { key: 'token', label: 'speedtest.setup_token', stage: 'credential', type: 'password' },
    ],
    steps: ['speedtest.setup_step_open', 'speedtest.setup_step_create', 'speedtest.setup_step_copy'],
    credentialPage: (address) => (address ? `${address}/admin/api-tokens` : null),
    async check(values) {
      // Thresholds stay off (0); the settings page sets them.
      await speedtest.saveSettings({
        url: values['url'] ?? '',
        token: values['token'] ?? '',
        minDownMbps: 0,
        minUpMbps: 0,
      });
      const r = await speedtest.results(24);
      const latest = r.latest;
      return {
        lines: [
          latest
            ? t('speedtest.setup_latest', {
                down: formatMbps(latest.downloadMbps),
                up: formatMbps(latest.uploadMbps),
              })
            : t('speedtest.no_results'),
        ],
      };
    },
  };
}
