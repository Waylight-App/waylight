import type { ServerMessageKey } from '../../../api-client/generated/server-messages/grafana';

/**
 * German translations of the backend error messages of `grafana` (#574). The keys come from the
 * generated catalogue of the same name; the type makes a missing translation a compile error. The
 * English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Grafana cannot render this panel as an image; install the Grafana image renderer.
  srv_8ff75b8c58539e49:
    'Grafana kann dieses Panel nicht als Bild ausgeben; installiere den Grafana Image Renderer.',
  // Grafana returned no values for this panel; panels that use dashboard variables need the image renderer.
  srv_de8bb22138954324:
    'Grafana hat für dieses Panel keine Werte geliefert; Panels mit Dashboard-Variablen brauchen den Image Renderer.',
  // Grafana is not configured yet.
  srv_a25fcc52746e2dd0: 'Grafana ist noch nicht eingerichtet.',
  // Grafana is not reachable.
  srv_2ee1fd290ca5d367: 'Grafana ist nicht erreichbar.',
  // The URL must look like http://192.168.1.2:3000.
  srv_ff20bffaaad40a7b: 'Die URL muss wie http://192.168.1.2:3000 aussehen.',
  // The dashboard does not exist.
  srv_0016d1a6fbb13c7f: 'Das Dashboard existiert nicht.',
  // The service account token is required.
  srv_3bbe665dcdec36da: 'Das Token des Dienstkontos ist erforderlich.',
  // The time range must look like 30m, 6h, 7d or 4w.
  srv_008998dbd8384168: 'Der Zeitraum muss wie 30m, 6h, 7d oder 4w aussehen.',
};
