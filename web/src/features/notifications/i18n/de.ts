import type { en } from './en';

/** German texts of the Slack, Teams and app channels; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'notifications.chat_webhook_url': 'Webhook-Adresse',
  'notifications.slack_hint':
    'Füge in Slack einem Kanal die App „Incoming Webhooks“ hinzu (oder lege eine Slack-App mit eingehenden Webhooks an) und füge hier die Webhook-URL ein. Sie wird verschlüsselt gespeichert und nicht wieder angezeigt.',
  'notifications.teams_hint':
    'Öffne in Teams die Workflows des Kanals und wähle „In einem Kanal posten, wenn eine Webhookanforderung empfangen wird“. Füge hier die Adresse des Workflows ein. Sie wird verschlüsselt gespeichert und nicht wieder angezeigt.',
  'notifications.app': 'Waylight-App',
  'notifications.app_hint':
    'Sendet an die Handys unten. Ein Handy erscheint, sobald jemand mit Adminrechten in der Waylight-App die Benachrichtigungen einschaltet. Die Nachrichten sind auf dem Weg über das Push-Relay Ende-zu-Ende-verschlüsselt.',
  'notifications.app_devices': 'Handys mit der Waylight-App',
  'notifications.app_devices_empty': 'Noch kein Handy hat Benachrichtigungen eingeschaltet.',
  'notifications.app_device_remove': 'Entfernen',
  'notifications.browser': 'Weboberfläche',
  'notifications.browser_hint':
    'Zeigt Benachrichtigungen live in dieser Weboberfläche, solange sie offen ist, direkt von diesem Server. Mit Erlaubnis des Browsers erscheinen sie auch als Systembenachrichtigung.',
  'notifications.browser_allow': 'Systembenachrichtigungen erlauben',
  'notifications.browser_allowed': 'Systembenachrichtigungen sind in diesem Browser erlaubt.',
  'notifications.browser_denied':
    'Dieser Browser blockiert Benachrichtigungen für Waylight; ändere das in den Website-Einstellungen.',
  'notifications.live_dismiss': 'Schließen',
};
