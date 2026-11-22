import type { ServerMessageKey } from '../../../api-client/generated/server-messages/smarthome';

/**
 * German translations of the backend error messages of `smarthome` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A long-lived access token is required.
  srv_084af9ca0c7d167d: 'Ein langlebiges Zugriffstoken ist erforderlich.',
  // Actions are not enabled for this entity (SMARTHOME_ACTIONS_ENTITIES).
  srv_0f4d7732f2cd4899: 'Aktionen sind für diese Entität nicht aktiviert (SMARTHOME_ACTIONS_ENTITIES).',
  // At most {max} entities per widget.
  srv_3a4c7bf887ab6487: 'Höchstens {max} Entitäten pro Widget.',
  // Home Assistant is not configured yet.
  srv_bcbabc6a08a54107: 'Home Assistant ist noch nicht eingerichtet.',
  // Home Assistant is not reachable.
  srv_dc90aa2a1e0a5f62: 'Home Assistant ist nicht erreichbar.',
  // Home Assistant rejected the action: {detail}
  srv_af1423afec138ad6: 'Home Assistant hat die Aktion abgelehnt: {detail}',
  // The address must look like http://homeassistant.local:8123.
  srv_079867a7aef69741: 'Die Adresse muss wie http://homeassistant.local:8123 aussehen.',
  // The entity does not exist.
  srv_2ef3700ca8a09eac: 'Die Entität existiert nicht.',
  // This entity has no action; only switches and buttons can be used.
  srv_a6715a41424b56f8: 'Diese Entität hat keine Aktion; nur Schalter und Taster lassen sich bedienen.',
  // Choose one entity for the compact view.
  srv_ea2866d2e8547dcb: 'Wähle eine Entität für die kompakte Ansicht.',
  // Home Assistant did not renew the sign-in. Set up Home Assistant again.
  srv_c900d4617320061f:
    'Home Assistant hat die Anmeldung nicht verlängert. Richte Home Assistant erneut ein.',
  // Off
  srv_302cbb19bf9a174a: 'Aus',
  // On
  srv_091d5807b5b33550: 'An',
  // Run
  srv_9fbbc819fe7e66aa: 'Ausführen',
  // Turn off
  srv_de736af382a428ab: 'Ausschalten',
  // Turn on
  srv_6bb329b249dd1ee7: 'Einschalten',
  // Unavailable
  srv_43e9aef7aa01fca9: 'Nicht verfügbar',
  // {name}
  srv_393ba5b6e23583b4: '{name}',
  // {state}
  srv_c7c0c20a29ee6976: '{state}',
  // {state} {unit}
  srv_4f14b24d395337a4: '{state} {unit}',
  // Choose Connect in the widget and sign in with Home Assistant, then pick the devices.
  srv_9a5cb1b37ed14c86:
    'Wähle Verbinden im Widget, melde dich mit Home Assistant an und wähle dann die Geräte.',
  // Devices from Home Assistant as tiles, with their state and, if allowed, a switch.
  srv_491401ddfbf37254:
    'Geräte aus Home Assistant als Kacheln, mit ihrem Zustand und, wenn erlaubt, einem Schalter.',
};
