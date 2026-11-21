import type { ServerMessageKey } from '../../../api-client/generated/server-messages/releases';

/**
 * German translations of the backend error messages of `releases` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most {max} artists can be followed.
  srv_987624bba2ac796c: 'Du kannst höchstens {max} Künstlern folgen.',
  // MusicBrainz is not reachable.
  srv_f9e4cfbeb07400ea: 'MusicBrainz ist nicht erreichbar.',
  // The artist ID is not a MusicBrainz ID.
  srv_cfe2240227cb188a: 'Die Künstler-ID ist keine MusicBrainz-ID.',
  // The artist name must be between 1 and 200 characters long.
  srv_d3b1108c6fe0b7c0: 'Der Künstlername muss zwischen 1 und 200 Zeichen lang sein.',
  // The cover does not exist.
  srv_6e79db37847b991c: 'Das Cover existiert nicht.',
  // Follow artists in Settings under New releases.
  srv_44ca48d69e755ae0: 'Folge Künstlern in den Einstellungen unter Neuerscheinungen.',
  // New albums and singles of the artists you follow.
  srv_750f3131063df0b9: 'Neue Alben und Singles der Künstler, denen du folgst.',
};
