import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import type { NextcloudUseCases } from '../usecase/nextcloudUseCases';

/** The setup assistant of the Nextcloud widget (#618), as data. */
export function nextcloudSetup(nextcloud: NextcloudUseCases): SetupDescriptor {
  const preview = async () => {
    const s = await nextcloud.status();
    return {
      lines: [t('nextcloud.setup_version', { version: s.version, percent: Math.round(s.quota.percent) })],
    };
  };
  return {
    service: 'Nextcloud',
    fields: [
      {
        key: 'url',
        label: 'setup.address',
        stage: 'address',
        type: 'url',
        placeholder: 'https://cloud.example.com',
      },
      {
        key: 'username',
        label: 'nextcloud.setup_username',
        stage: 'credential',
        type: 'text',
        autocomplete: 'username',
      },
      { key: 'password', label: 'nextcloud.setup_password', stage: 'credential', type: 'password' },
    ],
    steps: [
      'nextcloud.setup_step_settings',
      'nextcloud.setup_step_security',
      'nextcloud.setup_step_create',
      'nextcloud.setup_step_copy',
    ],
    credentialPage: (address) => (address ? `${address}/index.php/settings/user/security` : null),
    // Login Flow v2 (#654): Nextcloud creates the app password itself.
    signIn: {
      service: 'nextcloud',
      async check(_values, signIn) {
        await nextcloud.saveSignIn(signIn);
        return preview();
      },
    },
    async check(values) {
      await nextcloud.saveSettings(values['url'] ?? '', values['username'] ?? '', values['password'] ?? '');
      return preview();
    },
  };
}
