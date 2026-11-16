import type { Catalogue } from "../lib/texts";

/** Texts of the presence widget (#346). */
export const texts: Catalogue = {
  title: { en: "Presence", de: "Anwesenheit" },
  description: {
    en: "Who is in the office, at home or away today.",
    de: "Wer heute im Büro, im Homeoffice oder weg ist.",
  },
  office: { en: "Office", de: "Büro" },
  home: { en: "Home", de: "Homeoffice" },
  away: { en: "Away", de: "Weg" },
  none: { en: "Not set", de: "Nicht gesetzt" },
  my_status: { en: "My status today", de: "Mein Status heute" },
  return_on: { en: "Back on", de: "Zurück am" },
  back: { en: "back {date}", de: "zurück {date}" },
  from_absence: { en: "from the absences", de: "aus den Abwesenheiten" },
  read_only: {
    en: "The licence key has expired: statuses can only be read.",
    de: "Der Lizenzschlüssel ist abgelaufen: Status sind nur lesbar.",
  },
  empty: { en: "Nobody in this group.", de: "Niemand in dieser Gruppe." },
  disabled: {
    en: "Presence is switched off for this organisation.",
    de: "Anwesenheit ist für diese Organisation ausgeschaltet.",
  },
  counts: {
    en: "{office} in the office · {home} at home · {away} away · {none} not set",
    de: "{office} im Büro · {home} im Homeoffice · {away} weg · {none} nicht gesetzt",
  },
  group: { en: "Group", de: "Gruppe" },
  everyone: { en: "Everyone", de: "Alle" },
  save: { en: "Save", de: "Speichern" },
  section: { en: "Presence", de: "Anwesenheit" },
  section_hint: {
    en: "Lets everyone show whether they are in the office, at home or away today. Switching it off deletes all statuses.",
    de: "Alle zeigen, ob sie heute im Büro, im Homeoffice oder weg sind. Ausschalten löscht alle Status.",
  },
  enabled: { en: "Presence is on", de: "Anwesenheit ist an" },
};
