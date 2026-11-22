import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The warranty and documents widget (#688). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'documents',
      category: 'everyday',
      keywords: [
        'documents',
        'warranty',
        'receipt',
        'invoice',
        'guarantee',
        'dokumente',
        'garantie',
        'beleg',
        'quittung',
        'rechnung',
      ],
      get title() {
        return t('widget.documents.title');
      },
      get description() {
        return t('widget.documents.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'archive',
    compact: () => 'documents',
    compactExample: () => t('compact.example.documents'),
    view: lazy(
      () => import('../../features/documents/ui/DocumentsWidget.svelte'),
      (_config, c) => ({ documents: c.documents }),
    ),
  },
];
