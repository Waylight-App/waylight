import type { ServerMessageKey } from '../../../api-client/generated/server-messages/diskhealth';

/**
 * German translations of the backend error messages of `diskhealth` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // No S.M.A.R.T. export found. Run scripts/smart-export.sh on the host (see docs/host-integrations.md).
  srv_6bf218ac119ae629:
    'Kein S.M.A.R.T.-Export gefunden. Bitte scripts/smart-export.sh auf dem Host ausführen (siehe docs/host-integrations.md).',
  // No drive reports yet
  srv_f724ce57caa5bee3: 'Noch keine Laufwerksberichte',
  // Some reports are out of date
  srv_a8b97fb3f9f0cd6b: 'Einige Berichte sind veraltet',
  // {count} of {total} drives failing
  srv_26f374c0a67c8c36: '{count} von {total} Laufwerken fallen aus',
  // {count} of {total} drives need a look
  srv_2ac49ee2d1d4b8bc: '{count} von {total} Laufwerken brauchen einen Blick',
  // {device} has warnings
  srv_ccd2d42e76e2629c: '{device} hat Warnungen',
  // {device} is failing
  srv_fdd6ac3006cff821: '{device} fällt aus',
  // {healthy} of {total} drives healthy
  srv_4886a197fd6d3ced: '{healthy} von {total} Laufwerken gesund',
  // {total} drives healthy
  srv_4590f8e90f51cb7d: '{total} Laufwerke gesund',
};
