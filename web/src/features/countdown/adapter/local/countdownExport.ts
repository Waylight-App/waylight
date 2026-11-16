import { EXPORT_OWNER, exportTime, type ExportTables } from '../../../../shared/local/exportRows';
import type { CountdownDoc } from './localCountdownGateway';

const minuteOf = (time: string): number => {
  if (time === '') return -1; // whole day
  const [h = '0', m = '0'] = time.split(':');
  return Number(h) * 60 + Number(m);
};

/** The countdowns as rows of countdown_items (migrations of internal/modules/countdown). */
export function countdownExport(doc: CountdownDoc): ExportTables {
  return {
    countdown_items: doc.items.map((c) => ({
      id: c.id,
      title: c.title,
      emoji: c.emoji,
      day: c.date,
      minute: minuteOf(c.time),
      created_at: exportTime(c.createdAt),
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
  };
}
