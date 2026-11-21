/** English UI texts of compact widget views (#595). */
export const en = {
  'compact.loading': 'Loading…',
  'compact.unavailable': 'Not available',
  'compact.level_attention': 'Needs attention',
  'compact.level_problem': 'Problem',
  'compact.example.todos': '3 open · 1 overdue',
  'compact.example.habits': '2 of 5 today',
  'compact.example.focus': '50 min focus today',
  'compact.example.services': '11 of 12 running',
  'compact.example.status-overview': '11 of 12 running',
  'compact.example.calendar': 'Dentist · Today at 14:00',
  'compact.example.backup-status': 'Backup OK',
  'compact.example.diskhealth': '3 drives healthy',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
