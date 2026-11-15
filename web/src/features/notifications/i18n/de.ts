import type { en } from './en';

/** German texts of the Slack and Teams channels; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'notifications.chat_webhook_url': 'Webhook-URL',
  'notifications.slack_hint':
    'Füge in Slack einem Kanal die App „Incoming Webhooks“ hinzu (oder lege eine Slack-App mit eingehenden Webhooks an) und füge hier die Webhook-URL ein. Sie wird verschlüsselt gespeichert und nicht wieder angezeigt.',
  'notifications.teams_hint':
    'Öffne in Teams die Workflows des Kanals und wähle „In einem Kanal posten, wenn eine Webhookanforderung empfangen wird“. Füge hier die Adresse des Workflows ein. Sie wird verschlüsselt gespeichert und nicht wieder angezeigt.',
};
