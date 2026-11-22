/** English UI texts of the contracts feature (#687). */
export const en = {
  'widget.contracts.title': 'Contracts',
  'widget.contracts.description':
    'Contracts with cost and cancellation deadline, with a reminder before each deadline.',
  'compact.example.contracts': 'Gym: cancel by 31 Dec',
  'contracts.new': 'New contract',
  'contracts.edit': 'Change contract',
  'contracts.name': 'Contract',
  'contracts.provider': 'Provider',
  'contracts.cost': 'Cost ({currency})',
  'contracts.period': 'Due',
  'contracts.period.monthly': 'Monthly',
  'contracts.period.quarterly': 'Quarterly',
  'contracts.period.yearly': 'Yearly',
  'contracts.per.monthly': 'a month',
  'contracts.per.quarterly': 'a quarter',
  'contracts.per.yearly': 'a year',
  'contracts.term_end': 'End of term',
  'contracts.notice_days': 'Notice period (days)',
  'contracts.renew_months': 'Renews by (months, 0 = ends)',
  'contracts.remind_days': 'Remind days before (0 = never)',
  'contracts.deadline_hint':
    'The cancellation deadline is the end of the term minus the notice period. The reminder goes to your notification channels.',
  'contracts.note': 'Note',
  'contracts.total': '{amount} a month in total',
  'contracts.cancel_by': 'Cancel by {date}',
  'contracts.ended': 'Ended on {date}',
  'contracts.empty': 'No contracts yet.',
  'contracts.edit_named': 'Change {name}',
  'contracts.remove': 'Remove {name}',
  'contracts.currency': 'Currency',
  'contracts.currency_hint': 'Three-letter code such as EUR, CHF or USD.',
  'contracts.error.name': 'Enter a name of up to 80 characters.',
  'contracts.error.cost': 'Enter the cost as a number, such as 19.99.',
  'contracts.error.term_end': 'Choose the end of the term.',
  'contracts.error.numbers': 'Check the notice period, renewal and reminder.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
