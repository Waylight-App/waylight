/**
 * Entry of the ee/ frontend (ADR 0021): what the web app registers when the
 * server offers this bundle. Features without a licence stay listed in the
 * settings; their widgets are hidden by the server's module list.
 */
import { PLUGIN_API_VERSION, type TrissePlugin } from '../../../web/src/app/plugins/api';
import { texts as announcementTexts } from './announcements/texts';
import { texts as brandingTexts } from './branding/texts';
import { texts as connectionTexts } from './connections/texts';
import { texts as teamCalendarTexts } from './teamcalendar/texts';
import LicenceSection from './licence/LicenceSection.svelte';
import TeamTabsSection from './teamtabs/TeamTabsSection.svelte';
import { texts as teamTabTexts } from './teamtabs/texts';
import { lazyMountable, mountable } from './lib/mountable';

export const plugin: TrissePlugin = {
  apiVersion: PLUGIN_API_VERSION,
  name: 'ee',
  widgets: [
    {
      kind: 'announcements',
      feature: 'announcements',
      category: 'everyday',
      title: announcementTexts.title ?? { en: 'Announcements' },
      description: announcementTexts.description ?? { en: '' },
      keywords: ['announcement', 'notice', 'news', 'ankündigung', 'hinweis', 'aushang'],
      defaultSize: { w: 4, h: 5 },
      icon: 'rss',
      view: lazyMountable(() => import('./announcements/AnnouncementsWidget.svelte')),
    },
  ],
  settings: [
    {
      id: 'licence',
      label: {
        en: 'Features for organisations',
        de: 'Funktionen für Organisationen',
      },
      description: { en: '', de: '' },
      view: mountable(LicenceSection),
    },
    {
      id: 'team-tabs',
      feature: 'team-tabs',
      adminOnly: true,
      label: teamTabTexts.label ?? { en: 'Team tabs' },
      description: teamTabTexts.description ?? { en: '' },
      view: mountable(TeamTabsSection),
    },
    {
      id: 'branding',
      feature: 'branding',
      adminOnly: true,
      label: brandingTexts.label ?? { en: 'Branding' },
      description: brandingTexts.description ?? { en: '' },
      view: lazyMountable(() => import('./branding/BrandingSection.svelte')),
    },
    {
      id: 'connections',
      feature: 'connected-accounts',
      label: connectionTexts.label ?? { en: 'Connected accounts' },
      description: connectionTexts.description ?? { en: '' },
      view: lazyMountable(() => import('./connections/ConnectionsSection.svelte')),
    },
    {
      id: 'team-calendar',
      feature: 'absences',
      label: teamCalendarTexts.label ?? { en: 'Team calendar' },
      description: teamCalendarTexts.description ?? { en: '' },
      view: lazyMountable(() => import('./teamcalendar/TeamCalendarSection.svelte')),
    },
  ],
};
