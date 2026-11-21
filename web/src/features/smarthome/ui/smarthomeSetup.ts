import { t } from '../../../shared/i18n';
import type { FoundService, SetupDescriptor } from '../../../shared/setup/setup';
import type { SmarthomeUseCases } from '../usecase/smarthomeUseCases';

/** The setup assistant of the Home Assistant widget (#588), as data. */
export function smarthomeSetup(
  smarthome: SmarthomeUseCases,
  discover?: () => Promise<FoundService[]>,
): SetupDescriptor {
  const preview = async () => {
    const entities = await smarthome.entities();
    return {
      lines: [
        t('smarthome.setup_entities', { n: entities.length }),
        ...entities.slice(0, 3).map((e) => `${e.name}: ${e.state}${e.unit ? ` ${e.unit}` : ''}`),
      ],
    };
  };
  return {
    discover,
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
    // OAuth with Waylight's address as client ID (#654); needs PUBLIC_URL,
    // otherwise the server says so and the assistant asks for a token.
    signIn: {
      service: 'homeassistant',
      async check(_values, signIn) {
        await smarthome.saveSignIn(signIn);
        return preview();
      },
    },
    async check(values) {
      await smarthome.saveSettings(values['url'] ?? '', values['token'] ?? '');
      return preview();
    },
  };
}
