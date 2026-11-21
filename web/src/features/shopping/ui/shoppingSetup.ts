import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import { CALDAV } from '../domain/shopping';
import type { ShoppingUseCases } from '../usecase/shoppingUseCases';

/** The Nextcloud security page from a list address below /remote.php. */
function securityPage(address: string): string | null {
  const i = address.indexOf('/remote.php/');
  return i > 0 ? `${address.slice(0, i)}/settings/user/security` : null;
}

/** The setup assistant for a CalDAV task list in the shopping widget (#618), as data. */
export function shoppingSetup(shopping: ShoppingUseCases): SetupDescriptor {
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
    async check(values) {
      await shopping.saveSettings(
        values['listUrl'] ?? '',
        values['username'] ?? '',
        values['password'] ?? '',
      );
      await shopping.setCurrent(CALDAV);
      const items = await shopping.source(CALDAV).items();
      return { lines: [t('shopping.setup_found', { n: items.filter((i) => !i.done).length })] };
    },
  };
}
