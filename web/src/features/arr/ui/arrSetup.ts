import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import type { ArrApp } from '../domain/arr';
import type { ArrUseCases } from '../usecase/arrUseCases';

/** The setup assistant of the Sonarr/Radarr widget (#618), as data. */
export function arrSetup(arr: ArrUseCases): SetupDescriptor {
  return {
    service: 'Sonarr / Radarr',
    fields: [
      {
        key: 'app',
        label: 'arr.app',
        stage: 'address',
        type: 'choice',
        options: [
          { value: 'sonarr', label: 'Sonarr' },
          { value: 'radarr', label: 'Radarr' },
          { value: 'prowlarr', label: 'Prowlarr' },
        ],
      },
      {
        key: 'url',
        label: 'setup.address',
        stage: 'address',
        type: 'url',
        placeholder: 'http://192.168.178.20:8989',
      },
      { key: 'apiKey', label: 'arr.setup_api_key', stage: 'credential', type: 'password' },
    ],
    steps: ['arr.setup_step_open', 'arr.setup_step_copy'],
    // All three apps keep the key on the same page.
    credentialPage: (address) => (address ? `${address}/settings/general` : null),
    async check(values) {
      await arr.save({
        app: (values['app'] ?? 'sonarr') as ArrApp,
        name: '',
        url: values['url'] ?? '',
        apiKey: values['apiKey'] ?? '',
      });
      const o = await arr.overview();
      return {
        lines: [
          t('arr.setup_found', { upcoming: o.upcoming.length, queue: o.queue.length }),
          ...o.health.slice(0, 2).map((h) => h.message),
        ],
      };
    },
  };
}
