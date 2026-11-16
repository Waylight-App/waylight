import type { ServerMessageKey } from '../../../api-client/generated/server-messages/markets';

/** German translations of the backend messages of `markets` (#284). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Currency rates are switched off on this server.
  srv_1fd9959d0eaf237d: 'Wechselkurse sind auf diesem Server abgeschaltet.',
  // The currency rates could not be fetched; try again later.
  srv_02db4d3c7e0abaac: 'Die Wechselkurse konnten nicht geladen werden; versuche es später noch einmal.',
  // At most 12 entries per widget.
  srv_0d35c0616158d25f: 'Höchstens 12 Einträge pro Widget.',
  // Use fx:USD for currencies or stooq:aapl.us for shares, indices and crypto.
  srv_6588144872adfafb: 'Nutze fx:USD für Währungen oder stooq:aapl.us für Aktien, Indizes und Krypto.',
};
