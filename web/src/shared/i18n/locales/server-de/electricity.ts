import type { ServerMessageKey } from '../../../api-client/generated/server-messages/electricity';

/** German translations of the backend messages of `electricity` (#288). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose a supported bidding zone (DE-LU or AT).
  srv_0f6238acbb3eb2f8: 'Wähle eine unterstützte Gebotszone (DE-LU oder AT).',
  // No electricity prices were published for today yet.
  srv_0e1d3194c692037c: 'Für heute wurden noch keine Strompreise veröffentlicht.',
  // Cheap from {time}
  srv_f44c3b908996ef0f: 'Günstig ab {time}',
  // Cheap now
  srv_5ae21e1163e76e98: 'Jetzt günstig',
  // Cheap tomorrow from {time}
  srv_8ef85f0e320a5c28: 'Morgen günstig ab {time}',
  // No price for this hour
  srv_46d62be6e252d98a: 'Kein Preis für diese Stunde',
  // Prices could not be updated
  srv_0096b59b55372fb3: 'Preise konnten nicht aktualisiert werden',
  // {price} ct now
  srv_d7cd175f61b1223d: '{price} ct jetzt',
  // Open the widget settings and choose your country. No account or key is needed.
  srv_9d42ecd817b66e11:
    'Öffne die Widget-Einstellungen und wähle dein Land. Ein Konto oder Schlüssel ist nicht nötig.',
  // The electricity market price per hour for today and tomorrow, with the cheapest hours marked.
  srv_2a91edfdb4c2047e:
    'Der Strom-Börsenpreis pro Stunde für heute und morgen, die günstigsten Stunden sind markiert.',
};
