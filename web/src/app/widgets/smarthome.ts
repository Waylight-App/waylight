import { t } from '../../shared/i18n';
import { connectionParams } from '../../shared/connection/connection';
import { compactSource } from '../../features/compact/domain/compact';
import { parseWidgetConfig } from '../../features/smarthome/domain/smarthome';
import { choicesOf, lazy, type WidgetDefinition } from '../widgetKit';

/** The smarthome widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'smarthome',
      category: 'network',
      keywords: ['home assistant', 'switch', 'sensor', 'iot', 'smart home', 'hausautomation'],
      get title() {
        return t('widget.smarthome.title');
      },
      get description() {
        return t('widget.smarthome.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'activity',
    // A tile with one device shows it with On / Off (#614); a list of
    // devices has no single state to show.
    compact: (config) => {
      const { entities } = parseWidgetConfig(config);
      return entities.length === 1 && entities[0]
        ? compactSource('smarthome', { entity: entities[0], ...connectionParams(config) })
        : undefined;
    },
    compactExample: () => t('compact.example.smarthome'),
    view: lazy(
      () => import('../../features/smarthome/ui/SmartHomeWidget.svelte'),
      (config, c) => ({
        config,
        smarthome: c.smarthome,
        discover: () => c.discovery.find('smarthome'),
        connections: choicesOf(c),
      }),
    ),
    settings: lazy(
      () => import('../../features/smarthome/ui/SmartHomeWidgetSettings.svelte'),
      (config, c) => ({ config, smarthome: c.smarthome, connections: choicesOf(c) }),
    ),
  },
];
