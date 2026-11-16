/**
 * Entry of the ee/ frontend (ADR 0021): what the web app registers when the
 * server offers this bundle. Features without a licence stay listed in the
 * settings; their widgets are hidden by the server's module list.
 */
import {
  PLUGIN_API_VERSION,
  type TrissePlugin,
} from "../../../web/src/app/plugins/api";
import { texts as announcementTexts } from "./announcements/texts";
import LicenceSection from "./licence/LicenceSection.svelte";
import TeamTabsSection from "./teamtabs/TeamTabsSection.svelte";
import { texts as teamTabTexts } from "./teamtabs/texts";
import PresenceSection from "./presence/PresenceSection.svelte";
import { texts as presenceTexts } from "./presence/texts";
import { lazyMountable, mountable } from "./lib/mountable";

export const plugin: TrissePlugin = {
  apiVersion: PLUGIN_API_VERSION,
  name: "ee",
  widgets: [
    {
      kind: "announcements",
      feature: "announcements",
      category: "everyday",
      title: announcementTexts.title ?? { en: "Announcements" },
      description: announcementTexts.description ?? { en: "" },
      keywords: [
        "announcement",
        "notice",
        "news",
        "ankündigung",
        "hinweis",
        "aushang",
      ],
      defaultSize: { w: 4, h: 5 },
      icon: "rss",
      view: lazyMountable(
        () => import("./announcements/AnnouncementsWidget.svelte"),
      ),
    },
    {
      kind: "presence",
      feature: "presence",
      category: "everyday",
      title: presenceTexts.title ?? { en: "Presence" },
      description: presenceTexts.description ?? { en: "" },
      keywords: [
        "presence",
        "office",
        "home office",
        "who is in",
        "anwesenheit",
        "büro",
        "homeoffice",
      ],
      defaultSize: { w: 4, h: 5 },
      icon: "activity",
      view: lazyMountable(() => import("./presence/PresenceWidget.svelte")),
      settings: lazyMountable<{
        config?: Record<string, unknown>;
        onsave: (config: Record<string, unknown>) => void;
      }>(() => import("./presence/PresenceWidgetSettings.svelte")),
    },
  ],
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
    {
      id: "presence",
      feature: "presence",
      adminOnly: true,
      label: presenceTexts.section ?? { en: "Presence" },
      description: presenceTexts.section_hint ?? { en: "" },
      view: mountable(PresenceSection),
    },
  ],
};
