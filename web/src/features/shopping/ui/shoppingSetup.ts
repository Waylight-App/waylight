import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import { CALDAV } from '../domain/shopping';
import type { ShoppingUseCases } from '../usecase/shoppingUseCases';

/** The Nextcloud address of a list address below /remote.php. */
export function nextcloudOf(address: string): string | null {
  const i = address.indexOf('/remote.php/');
  return i > 0 ? address.slice(0, i) : null;
}

/** The Nextcloud security page from a list address below /remote.php. */
function securityPage(address: string): string | null {
  const base = nextcloudOf(address);
  return base ? `${base}/settings/user/security` : null;
}

/** The setup assistant for a CalDAV task list in the shopping widget (#618), as data. */
export function shoppingSetup(shopping: ShoppingUseCases): SetupDescriptor {
  const preview = async () => {
    await shopping.setCurrent(CALDAV);
    const items = await shopping.source(CALDAV).items();
    return { lines: [t('shopping.setup_found', { n: items.filter((i) => !i.done).length })] };
  };
  return {
    service: 'CalDAV',
    fields: [
      {
        key: 'listUrl',
        label: 'shopping.setup_list_url',
        stage: 'address',
        type: 'url',
        placeholder: 'https://cloud.example.org/remote.php/dav/calendars/anna/shopping/',
      },
      { key: 'username', label: 'shopping.setup_username', stage: 'credential', type: 'text' },
      { key: 'password', label: 'shopping.setup_password', stage: 'credential', type: 'password' },
    ],
    steps: ['shopping.setup_step_security', 'shopping.setup_step_create', 'shopping.setup_step_copy'],
    credentialPage: securityPage,
    // Login Flow v2 (#654) at the Nextcloud the list lies on; other CalDAV
    // servers keep the pasted password.
    signIn: {
      service: 'nextcloud',
      label: 'Nextcloud',
      address: (values) => nextcloudOf((values['listUrl'] ?? '').trim()),
      async check(values, signIn) {
        await shopping.saveSignIn(values['listUrl'] ?? '', signIn);
        return preview();
      },
    },
    async check(values) {
      await shopping.saveSettings(
        values['listUrl'] ?? '',
        values['username'] ?? '',
        values['password'] ?? '',
      );
      return preview();
    },
  };
}
