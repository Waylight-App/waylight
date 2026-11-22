import { t } from '../../shared/i18n';
import { connectionParams } from '../../shared/connection/connection';
import { compactSource } from '../../features/compact/domain/compact';
import { choicesOf, lazy, type WidgetDefinition } from '../widgetKit';

/** The proxmox widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'proxmox',
      category: 'server',
      keywords: ['pve', 'vm', 'lxc', 'virtualisation', 'hypervisor'],
      get title() {
        return t('widget.proxmox.title');
      },
      get description() {
        return t('widget.proxmox.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'server',
    compact: (config) => compactSource('proxmox', connectionParams(config)),
    compactExample: () => t('compact.example.proxmox'),
    view: lazy(
      () => import('../../features/proxmox/ui/ProxmoxWidget.svelte'),
      (config, c) => ({ config, proxmox: c.proxmox, connections: choicesOf(c) }),
    ),
    settings: lazy(
      () => import('../../features/proxmox/ui/ProxmoxWidgetSettings.svelte'),
      (config, c) => ({ config, connections: choicesOf(c) }),
    ),
  },
];
