import type { ServerMessageKey } from '../../../api-client/generated/server-messages/services';

/**
 * German translations of the backend error messages of `services` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // No MAC address is stored for this service.
  srv_591001f7788fb502: 'Für diesen Dienst ist keine MAC-Adresse hinterlegt.',
  // The check address must start with http:// or https://.
  srv_aa968d1ebc0107b3: 'Die Prüfadresse muss mit http:// oder https:// beginnen.',
  // The description must be at most 200 characters long.
  srv_b6d4e2000983a0fe: 'Die Beschreibung darf höchstens 200 Zeichen lang sein.',
  // The icon is too long.
  srv_42e741b2bbfc2695: 'Das Icon ist zu lang.',
  // The new order must contain every service exactly once.
  srv_a75b491a5e89aabe: 'Die neue Reihenfolge muss jeden Dienst genau einmal enthalten.',
  // The service does not exist.
  srv_df6166e3647fbc04: 'Der Dienst existiert nicht.',
  // Docker discovery needs the docker module to be enabled.
  srv_1f3976061912504c: 'Die Docker-Autodiscovery braucht das aktivierte Docker-Modul.',
  // The interval must be between 1 and {max} minutes.
  srv_c2aa1fabd915da16: 'Das Intervall muss zwischen 1 und {max} Minuten liegen.',
  // Unknown merge rule.
  srv_c64f8cf81797f821: 'Unbekannte Abgleichsregel.',
  // Each service name can be used only once in the file.
  srv_2ad35206dae11b67: 'Jeder Dienstname darf in der Datei nur einmal vorkommen.',
  // No services yet
  srv_e40289b4499ed164: 'Noch keine Dienste',
  // {count} services down
  srv_ac66b193c5335c18: '{count} Dienste ausgefallen',
  // {name} is down
  srv_09bd02ddefc751e8: '{name} ist ausgefallen',
  // {online} of {total} running
  srv_85ec6321e77d0b4a: '{online} von {total} laufen',
  // A status page of your services: uptime over the last days and current or past outages.
  srv_65797128a11803b4:
    'Eine Statusseite deiner Dienste: Verfügbarkeit der letzten Tage und aktuelle oder vergangene Ausfälle.',
  // Add services in Settings under Services, or let Waylight find them on your network. Containers with Waylight labels appear by themselves.
  srv_8838681c99377be8:
    'Füge Dienste in den Einstellungen unter Dienste hinzu oder lass Waylight sie in deinem Netz finden. Container mit Waylight-Labels erscheinen von selbst.',
  // It uses the services you added in Settings under Services; there is nothing else to set up.
  srv_e02aa2085e4c6961:
    'Es nutzt die Dienste aus den Einstellungen unter Dienste; sonst ist nichts einzurichten.',
  // Tiles for your own services with their online status, response time and certificate warnings.
  srv_82c86b04ef996a5a:
    'Kacheln für deine eigenen Dienste mit Online-Status, Antwortzeit und Zertifikatswarnungen.',
};
