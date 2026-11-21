import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import type { SmarthomeUseCases } from '../usecase/smarthomeUseCases';

/** The setup assistant of the Home Assistant widget (#588), as data. */
export function smarthomeSetup(smarthome: SmarthomeUseCases): SetupDescriptor {
  return {
    service: 'Home Assistant',
    fields: [
      {
        key: 'url',
        label: 'setup.address',
        stage: 'address',
        type: 'url',
        placeholder: 'http://homeassistant.local:8123',
      },
      { key: 'token', label: 'smarthome.setup_token', stage: 'credential', type: 'password' },
    ],
    steps: [
      'smarthome.setup_step_profile',
      'smarthome.setup_step_security',
      'smarthome.setup_step_create',
      'smarthome.setup_step_copy',
    ],
    credentialPage: (address) => (address ? `${address}/profile/security` : null),
    async check(values) {
      await smarthome.saveSettings(values['url'] ?? '', values['token'] ?? '');
      const entities = await smarthome.entities();
      return {
        lines: [
          t('smarthome.setup_entities', { n: entities.length }),
          ...entities.slice(0, 3).map((e) => `${e.name}: ${e.state}${e.unit ? ` ${e.unit}` : ''}`),
        ],
      };
    },
  };
}
