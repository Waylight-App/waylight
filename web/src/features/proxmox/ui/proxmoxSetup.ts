import { t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import type { ProxmoxUseCases } from '../usecase/proxmoxUseCases';

/** The setup assistant of the Proxmox widget (#618), as data. */
export function proxmoxSetup(proxmox: ProxmoxUseCases): SetupDescriptor {
  return {
    service: 'Proxmox VE',
    fields: [
      {
        key: 'url',
        label: 'setup.address',
        stage: 'address',
        type: 'url',
        placeholder: 'https://192.168.178.10:8006',
      },
      {
        key: 'tokenId',
        label: 'proxmox.setup_token_id',
        stage: 'credential',
        type: 'text',
        placeholder: 'waylight@pve!dashboard',
      },
      { key: 'secret', label: 'proxmox.setup_secret', stage: 'credential', type: 'password' },
    ],
    steps: [
      'proxmox.setup_step_open',
      'proxmox.setup_step_add',
      'proxmox.setup_step_role',
      'proxmox.setup_step_copy',
    ],
    credentialPage: (address) => (address ? `${address}/#v1:0:18:4:::::::` : null),
    async check(values) {
      // A self-signed certificate is pinned in the settings page, which
      // stays the place for it; the assistant connects without a pin.
      await proxmox.saveSettings({
        url: values['url'] ?? '',
        tokenId: values['tokenId'] ?? '',
        secret: values['secret'] ?? '',
        fingerprint: '',
      });
      const s = await proxmox.status();
      return {
        lines: [t('proxmox.setup_found', { nodes: s.nodes.length, guests: s.guests.length })],
      };
    },
  };
}
