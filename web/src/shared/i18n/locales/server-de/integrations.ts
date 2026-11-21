import type { ServerMessageKey } from '../../../api-client/generated/server-messages/integrations';

/**
 * German translations of the backend error messages of `integrations` (#102, #610).
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
  // Fill in every required credential field.
  srv_99419ae88ff40fe8: 'Bitte alle nötigen Felder der Zugangsdaten ausfüllen.',
  // The connection does not exist.
  srv_7441a46151ee9363: 'Die Verbindung gibt es nicht.',
  // The name must be between 1 and {max} characters long.
  srv_08f7a91881c6888d: 'Der Name muss zwischen 1 und {max} Zeichen lang sein.',
  // The service does not expect this credential field.
  srv_6647ccd9d57ef103: 'Dieses Feld der Zugangsdaten kennt der Dienst nicht.',
  // The service is not valid.
  srv_7d55a5e68a7c5904: 'Der Dienst ist ungültig.',
  // This connection cannot be checked while its module is switched off.
  srv_7733b67aac993bdb: 'Die Verbindung lässt sich nicht prüfen, solange ihr Modul ausgeschaltet ist.',
  // This service cannot be connected.
  srv_1dd91a0989e27e29: 'Mit diesem Dienst ist keine Verbindung möglich.',
  // Widgets still use this connection. Remove it anyway?
  srv_afd218c3b96ddb04: 'Widgets nutzen diese Verbindung noch. Trotzdem entfernen?',
};
