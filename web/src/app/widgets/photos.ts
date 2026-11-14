import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The photo frame widget (#286). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'photo-frame',
      category: 'everyday',
      keywords: ['photo', 'picture', 'slideshow', 'foto', 'bild', 'bilderrahmen', 'diashow', 'album'],
      get title() {
        return t('widget.photo_frame.title');
      },
      get description() {
        return t('widget.photo_frame.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'archive',
    view: lazy(
      () => import('../../features/photos/ui/PhotoFrameWidget.svelte'),
      (config, c) => ({ config, photos: c.photos }),
    ),
    settings: lazy(
      () => import('../../features/photos/ui/PhotoFrameSettings.svelte'),
      (config, c) => ({ config, photos: c.photos }),
    ),
  },
];
