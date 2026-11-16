/**
 * Entry of the ee/ frontend (ADR 0021): what the web app registers when the
 * server offers this bundle. Features without a licence stay listed in the
 * settings; their widgets are hidden by the server's module list.
 */
import {
  PLUGIN_API_VERSION,
  type TrissePlugin,
} from "../../../web/src/app/plugins/api";
import LicenceSection from "./licence/LicenceSection.svelte";
import TeamTabsSection from "./teamtabs/TeamTabsSection.svelte";
import { texts as teamTabTexts } from "./teamtabs/texts";
import { mountable } from "./lib/mountable";

export const plugin: TrissePlugin = {
  apiVersion: PLUGIN_API_VERSION,
  name: "ee",
  widgets: [],
  settings: [
    {
      id: "licence",
      label: {
        en: "Features for organisations",
        de: "Funktionen für Organisationen",
      },
      description: { en: "", de: "" },
      view: mountable(LicenceSection),
    },
    {
      id: "team-tabs",
      feature: "team-tabs",
      adminOnly: true,
      label: teamTabTexts.label ?? { en: "Team tabs" },
      description: teamTabTexts.description ?? { en: "" },
      view: mountable(TeamTabsSection),
    },
  ],
};
