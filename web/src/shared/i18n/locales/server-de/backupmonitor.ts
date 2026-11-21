import type { ServerMessageKey } from '../../../api-client/generated/server-messages/backupmonitor';

/**
 * German translations of the backend error messages of `backupmonitor` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The backup source does not exist.
  srv_24356b769e2a8258: 'Die Backup-Quelle existiert nicht.',
  // The grace period must be between 0 minutes and 7 days.
  srv_dcd48aa7c1edb891: 'Die Karenzzeit muss zwischen 0 Minuten und 7 Tagen liegen.',
  // The kind must be backup or check.
  srv_0dbc446d2a829ab3: 'Die Art muss backup oder check sein.',
  // The maximum age must be between 1 hour and 60 days.
  srv_761ff31914abb39c: 'Das Höchstalter muss zwischen 1 Stunde und 60 Tagen liegen.',
  // Use a cron expression with five fields, e.g. 0 3 * * * (daily at 03:00).
  srv_564c3b1fecfb1093:
    'Bitte einen Cron-Ausdruck mit fünf Feldern verwenden, z. B. 0 3 * * * (täglich um 03:00).',
  // Backup OK
  srv_54022920e535da37: 'Sicherung OK',
  // Last success {time}
  srv_dca29003ab21344d: 'Zuletzt erfolgreich {time}',
  // No backups yet
  srv_4efa152402f99eaf: 'Noch keine Sicherungen',
  // Waiting for the first run
  srv_122bbf8d15074197: 'Wartet auf den ersten Lauf',
  // {count} of {total} failed
  srv_690e8b1d63e37e2e: '{count} von {total} fehlgeschlagen',
  // {count} of {total} overdue
  srv_8ede15696ca60c31: '{count} von {total} überfällig',
  // {name} failed
  srv_44460b43e14f12f1: '{name} ist fehlgeschlagen',
  // {name} is overdue
  srv_654703b9e7d6cc6c: '{name} ist überfällig',
  // {total} backups OK
  srv_5533d9eb788ce382: '{total} Sicherungen OK',
  // In Settings under Backup monitoring, add a job and let it call its secret link after each run.
  srv_8706d83e48cb5271:
    'Lege in den Einstellungen unter Backup-Überwachung eine Aufgabe an und lass sie nach jedem Lauf ihren geheimen Link aufrufen.',
  // Whether your backups and scheduled jobs ran on time and without errors.
  srv_4dc4f6b65d55e8ad: 'Ob deine Backups und geplanten Aufgaben pünktlich und ohne Fehler gelaufen sind.',
};
