import type { en } from './en';

/** German texts of the Slack, Teams and app channels; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'notifications.chat_webhook_url': 'Webhook-URL',
  'notifications.slack_hint':
    'Füge in Slack einem Kanal die App „Incoming Webhooks“ hinzu (oder lege eine Slack-App mit eingehenden Webhooks an) und füge hier die Webhook-URL ein. Sie wird verschlüsselt gespeichert und nicht wieder angezeigt.',
  'notifications.teams_hint':
    'Öffne in Teams die Workflows des Kanals und wähle „In einem Kanal posten, wenn eine Webhookanforderung empfangen wird“. Füge hier die Adresse des Workflows ein. Sie wird verschlüsselt gespeichert und nicht wieder angezeigt.',
  'notifications.app': 'Waylight-App',
  'notifications.app_hint':
    'Sendet an die Handys unten. Ein Handy erscheint, sobald jemand mit Adminrechten in der Waylight-App die Benachrichtigungen einschaltet. Die Nachrichten sind auf dem Weg über Google oder Apple Ende-zu-Ende-verschlüsselt.',
  'notifications.app_devices': 'Handys mit der Waylight-App',
  'notifications.app_devices_empty': 'Noch kein Handy hat Benachrichtigungen eingeschaltet.',
  'notifications.app_device_remove': 'Entfernen',
};
