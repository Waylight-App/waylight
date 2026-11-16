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
import { texts as brandingTexts } from "./branding/texts";
import LicenceSection from "./licence/LicenceSection.svelte";
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
      id: "branding",
      feature: "branding",
      adminOnly: true,
      label: brandingTexts.label ?? { en: "Branding" },
      description: brandingTexts.description ?? { en: "" },
      view: lazyMountable(() => import("./branding/BrandingSection.svelte")),
    },
  ],
};
