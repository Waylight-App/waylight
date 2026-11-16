import type { Catalogue } from "../lib/texts";

/** Texts of the team tabs section (#295). */
export const texts: Catalogue = {
  label: { en: "Team tabs", de: "Team-Tabs" },
  description: {
    en: "Share one of your dashboard tabs with a group: its members see it next to their own tabs and cannot change it.",
    de: "Teile einen deiner Dashboard-Tabs mit einer Gruppe: Ihre Mitglieder sehen ihn neben ihren eigenen Tabs und können ihn nicht ändern.",
  },
  hint: {
    en: "Build the tab on your own dashboard first, then choose who else sees it.",
    de: "Richte den Tab zuerst auf deinem eigenen Dashboard ein und wähle dann, wer ihn noch sieht.",
  },
  only_me: { en: "Only me", de: "Nur ich" },
  everyone: { en: "Everyone", de: "Alle" },
  group: { en: "Group: {name}", de: "Gruppe: {name}" },
  defaults_title: {
    en: "Default tab per group",
    de: "Standard-Tab pro Gruppe",
  },
  defaults_hint: {
    en: "Opens first for members of the group, unless their own tab rules pick another one.",
    de: "Öffnet sich für Mitglieder der Gruppe zuerst, sofern ihre eigenen Tab-Regeln keinen anderen wählen.",
  },
  no_default: { en: "No default", de: "Kein Standard" },
  saved: { en: "Saved.", de: "Gespeichert." },
  no_tabs: { en: "You have no tabs yet.", de: "Du hast noch keine Tabs." },
  read_only: {
    en: "The licence has expired: team tabs stay visible, but sharing cannot be changed.",
    de: "Die Lizenz ist abgelaufen: Team-Tabs bleiben sichtbar, die Freigabe lässt sich aber nicht ändern.",
  },
};
