/** English texts of the Slack and Teams channels (#324, #325). */
export const en = {
  'notifications.chat_webhook_url': 'Webhook URL',
  'notifications.slack_hint':
    'In Slack, add the app “Incoming Webhooks” to a channel (or create a Slack app with incoming webhooks) and paste the webhook URL here. It is stored encrypted and never shown again.',
  'notifications.teams_hint':
    'In Teams, open the channel’s Workflows and choose “Post to a channel when a webhook request is received”. Paste the address of the workflow here. It is stored encrypted and never shown again.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
