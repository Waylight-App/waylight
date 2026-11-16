import type { Catalogue } from "../lib/texts";

/** Texts of the connected accounts section (#326). */
export const texts: Catalogue = {
  label: { en: "Connected accounts", de: "Verbundene Konten" },
  description: {
    en: "Link your own account at a service, so search and chat apps can act for you with your permissions.",
    de: "Verbinde dein eigenes Konto bei einem Dienst, damit Suche und Chat-Apps mit deinen Rechten für dich handeln können.",
  },
  connect: { en: "Connect", de: "Verbinden" },
  reconnect: { en: "Connect again", de: "Erneut verbinden" },
  disconnect: { en: "Disconnect", de: "Trennen" },
  connected_since: {
    en: "Connected since {date}",
    de: "Verbunden seit {date}",
  },
  expired: {
    en: "The connection has expired.",
    de: "Die Verbindung ist abgelaufen.",
  },
  not_connected: { en: "Not connected", de: "Nicht verbunden" },
  not_configured: {
    en: "Not set up by an admin yet",
    de: "Noch nicht von einem Admin eingerichtet",
  },
  done: { en: "{name} is connected.", de: "{name} ist verbunden." },
  failed: {
    en: "Connecting {name} did not work. Try again.",
    de: "Die Verbindung mit {name} hat nicht geklappt. Versuche es erneut.",
  },
  admin_title: {
    en: "Set up services (admins)",
    de: "Dienste einrichten (Admins)",
  },
  admin_hint: {
    en: "Register this instance as an app at each service and enter its client ID and secret. Use this redirect address there:",
    de: "Registriere diese Instanz bei jedem Dienst als App und trage Client-ID und Secret ein. Gib dort diese Weiterleitungsadresse an:",
  },
  client_id: { en: "Client ID", de: "Client-ID" },
  client_secret: { en: "Client secret", de: "Client-Secret" },
  secret_stored: {
    en: "Stored; leave empty to keep it",
    de: "Gespeichert; leer lassen, um es zu behalten",
  },
  tenant: { en: "Tenant (Microsoft only)", de: "Tenant (nur Microsoft)" },
  save: { en: "Save", de: "Speichern" },
  saved: { en: "Saved.", de: "Gespeichert." },
};
