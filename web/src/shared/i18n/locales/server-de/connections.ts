import type { ServerMessageKey } from '../../../api-client/generated/server-messages/connections';

/**
 * German translations of the backend error messages of `connections` (#102, #610).
 * The keys come from the generated catalogue (run `make generate` first); the
 * type makes a missing translation a compile error.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A connection of this service with this name exists already.
  srv_5ca079545b8105a1: 'Eine Verbindung dieses Dienstes mit diesem Namen gibt es schon.',
  // A credential field has an invalid name.
  srv_366d1409958e448c: 'Ein Feld der Zugangsdaten hat einen ungültigen Namen.',
  // A credential has at most {max} fields.
  srv_d960734b1f081510: 'Zugangsdaten haben höchstens {max} Felder.',
  // A credential value must be at most {max} characters long.
  srv_783587b0f09187ba: 'Ein Wert der Zugangsdaten darf höchstens {max} Zeichen lang sein.',
  // At most {max} connections can be saved.
  srv_21f5be4823deec3c: 'Höchstens {max} Verbindungen können gespeichert werden.',
  // The connection does not exist.
  srv_7441a46151ee9363: 'Die Verbindung gibt es nicht.',
  // The name must be between 1 and {max} characters long.
  srv_08f7a91881c6888d: 'Der Name muss zwischen 1 und {max} Zeichen lang sein.',
  // The service is not valid.
  srv_7d55a5e68a7c5904: 'Der Dienst ist ungültig.',
  // Widgets still use this connection. Remove it anyway?
  srv_afd218c3b96ddb04: 'Widgets nutzen diese Verbindung noch. Trotzdem entfernen?',
};
