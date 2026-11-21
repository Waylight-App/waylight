import type { ServerMessageKey } from '../../../api-client/generated/server-messages/fritzbox';

/**
 * German translations of the backend error messages of `fritzbox` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Please enter the password of the FRITZ!Box user.
  srv_eabd506dfb319855: 'Bitte gib das Passwort des FRITZ!Box-Benutzers ein.',
  // The FRITZ!Box is not configured yet.
  srv_6f0fbfbef4b05402: 'Die FRITZ!Box ist noch nicht eingerichtet.',
  // The URL must look like http://192.168.178.1:49000.
  srv_d14b64d005d328ef: 'Die URL muss wie http://192.168.178.1:49000 aussehen.',
  // Show between {min} and {max} calls.
  srv_3de8440f30f2cb18: 'Zeige zwischen {min} und {max} Anrufen.',
  // The filter must be all, missed, incoming or outgoing.
  srv_ab808509e7ff3cb7: 'Der Filter muss alle, verpasst, eingehend oder ausgehend sein.',
  // Choose Set up in the widget and enter the address of your FRITZ!Box and a user name and password of the box.
  srv_d74ab61cf0423be5:
    'Wähle Einrichten im Widget und gib die Adresse deiner FRITZ!Box sowie einen Benutzernamen und ein Passwort der Box ein.',
  // It uses the FRITZ!Box you set up for the internet widget. Open the widget settings to choose which calls are shown.
  srv_30aa95cedabf4c8e:
    'Es nutzt die FRITZ!Box, die du für das Internet-Widget eingerichtet hast. In den Widget-Einstellungen wählst du, welche Anrufe angezeigt werden.',
  // Recent calls of your FRITZ!Box: missed, incoming and outgoing.
  srv_34b961f29d48bdbc: 'Die letzten Anrufe deiner FRITZ!Box: verpasst, angenommen und ausgehend.',
  // Your internet connection as the FRITZ!Box sees it: online status, speed and how long it has been up.
  srv_3fb2a0557bebb033:
    'Deine Internetverbindung aus Sicht der FRITZ!Box: Online-Status, Geschwindigkeit und wie lange sie schon steht.',
};
