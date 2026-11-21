import { t } from '../../../shared/i18n';
import { formatCount } from '../../../shared/ui/format';
import type { FoundService, SetupDescriptor } from '../../../shared/setup/setup';
import { thumbnailUrl } from '../domain/immich';
import type { ImmichUseCases } from '../usecase/immichUseCases';

/** The setup assistant of the Immich widget (#588), as data. */
export function immichSetup(
  immich: ImmichUseCases,
  discover?: () => Promise<FoundService[]>,
): SetupDescriptor {
  return {
    discover,
    service: 'Immich',
    fields: [
      {
        key: 'url',
        label: 'setup.address',
        stage: 'address',
        type: 'url',
        placeholder: 'http://192.168.178.20:2283',
      },
      { key: 'apiKey', label: 'immich.setup_key', stage: 'credential', type: 'password' },
    ],
    steps: [
      'immich.setup_step_profile',
      'immich.setup_step_keys',
      'immich.setup_step_permissions',
      'immich.setup_step_copy',
    ],
    // Immich opens its settings with the API key panel expanded from this link.
    credentialPage: (address) => (address ? `${address}/user-settings?isOpen=api-keys` : null),
    async check(values) {
      await immich.saveSettings(values['url'] ?? '', values['apiKey'] ?? '');
      const o = await immich.overview();
      return {
        lines: [t('immich.setup_photos', { n: formatCount(o.photos), videos: formatCount(o.videos) })],
        images: o.memories
          .flatMap((m) => m.assetIds)
          .slice(0, 3)
          .map(thumbnailUrl),
      };
    },
  };
}
