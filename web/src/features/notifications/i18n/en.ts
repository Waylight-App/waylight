/** English texts of the Slack, Teams and app channels (#324, #325, #564). */
export const en = {
  'notifications.chat_webhook_url': 'Webhook URL',
  'notifications.slack_hint':
    'In Slack, add the app “Incoming Webhooks” to a channel (or create a Slack app with incoming webhooks) and paste the webhook URL here. It is stored encrypted and never shown again.',
  'notifications.teams_hint':
    'In Teams, open the channel’s Workflows and choose “Post to a channel when a webhook request is received”. Paste the address of the workflow here. It is stored encrypted and never shown again.',
  'notifications.app': 'Waylight app',
  'notifications.app_hint':
    'Sends to the phones below. A phone appears once someone with admin rights turns on notifications in the Waylight app. Messages are end-to-end encrypted on the way through Google or Apple.',
  'notifications.app_devices': 'Phones with the Waylight app',
  'notifications.app_devices_empty': 'No phone has turned on notifications yet.',
  'notifications.app_device_remove': 'Remove',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
