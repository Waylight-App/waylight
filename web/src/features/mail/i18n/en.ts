/** English UI texts of the mail widget (#283). */
export const en = {
  'widget.mail.title': 'Mail',
  'widget.mail.description': 'Unread count and the newest unread messages of your inboxes.',
  'settings.section.mail': 'Mail',
  'mail.unread': '{n} unread',
  'mail.none': 'No unread mail.',
  'mail.open_webmail': 'Open webmail',
  'mail.failed': 'Could not be read: {names}',
  'mail.account': 'Account',
  'mail.all_accounts': 'All accounts',
  'mail.intro':
    'Read-only: nothing is marked as read, moved or deleted, and only sender, subject and date are read, never the message text. Use an app password.',
  'mail.guests':
    'Guests who can see a tab with a mail widget see the senders and subjects in it. Put the widget on a tab of your own if that matters.',
  'mail.accounts': 'Accounts',
  'mail.no_accounts': 'No account yet.',
  'mail.add': 'Add an account',
  'mail.name': 'Name',
  'mail.host': 'IMAP server',
  'mail.port': 'Port',
  'mail.username': 'User name',
  'mail.password': 'App password',
  'mail.folder': 'Folder',
  'mail.webmail': 'Webmail address (optional)',
  'mail.save': 'Test and save',
  'mail.saved': 'Connected and saved.',
  'mail.remove': 'Remove {name}',
  'mail.fill_in': 'Enter name, IMAP server and user name.',
  'mail.password_needed': 'Enter the app password.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
