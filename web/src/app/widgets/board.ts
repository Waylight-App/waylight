import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The family board widget (#685). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'board',
      category: 'everyday',
      keywords: [
        'board',
        'note',
        'message',
        'fridge',
        'family',
        'pinnwand',
        'notiz',
        'nachricht',
        'familie',
        'kühlschrank',
      ],
      get title() {
        return t('widget.board.title');
      },
      get description() {
        return t('widget.board.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'pencil',
    compact: () => 'board',
    compactExample: () => t('compact.example.board'),
    view: lazy(
      () => import('../../features/board/ui/BoardWidget.svelte'),
      (_config, c) => ({ board: c.board }),
    ),
  },
];
