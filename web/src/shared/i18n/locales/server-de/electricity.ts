import type { ServerMessageKey } from '../../../api-client/generated/server-messages/electricity';

/** German translations of the backend messages of `electricity` (#288). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose a supported bidding zone (DE-LU or AT).
  srv_0f6238acbb3eb2f8: 'Wähle eine unterstützte Gebotszone (DE-LU oder AT).',
  // No electricity prices were published for today yet.
  srv_0e1d3194c692037c: 'Für heute wurden noch keine Strompreise veröffentlicht.',
};
