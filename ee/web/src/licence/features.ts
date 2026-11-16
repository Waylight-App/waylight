import type { Texts } from '../../../../web/src/app/plugins/api';

/** A licence feature as GET /api/licence/features names it. */
export interface FeatureInfo {
  id: string;
  name: Texts;
  description: Texts;
}

/**
 * Every feature of ee/ in a stable order, with its name and a short
 * description. The settings list all of them, also the ones that are off.
 * The ids match domain.KnownFeatures in ee/licence.
 */
export const features: FeatureInfo[] = [
  { id: 'sso', name: { en: 'Single sign-on for all accounts', de: 'Single Sign-on für alle Konten' }, description: { en: 'Members sign in with your identity provider; groups are mapped.', de: 'Mitglieder melden sich über euren Identity Provider an; Gruppen werden übernommen.' } },
  { id: 'directory-sync', name: { en: 'Directory sync', de: 'Verzeichnis-Abgleich' }, description: { en: 'Accounts and groups from LDAP, Active Directory or SCIM.', de: 'Konten und Gruppen aus LDAP, Active Directory oder SCIM.' } },
  { id: 'policies', name: { en: 'Admin policies', de: 'Admin-Richtlinien' }, description: { en: 'Required second factor, session length, allowed modules.', de: 'Pflicht zum zweiten Faktor, Sitzungsdauer, erlaubte Module.' } },
  { id: 'audit-log', name: { en: 'Audit log', de: 'Audit-Log' }, description: { en: 'Content changes with export and retention.', de: 'Inhaltsänderungen mit Export und Aufbewahrung.' } },
  { id: 'team-tabs', name: { en: 'Team tabs', de: 'Team-Tabs' }, description: { en: 'Dashboard tabs that admins maintain for groups.', de: 'Dashboard-Tabs, die Admins für Gruppen pflegen.' } },
  { id: 'announcements', name: { en: 'Announcements', de: 'Ankündigungen' }, description: { en: 'Notices for everyone, with read marks and expiry.', de: 'Hinweise für alle, mit Gelesen-Markierung und Ablauf.' } },
  { id: 'absences', name: { en: 'Absences and team calendar', de: 'Abwesenheiten und Teamkalender' }, description: { en: 'Who is away when, for the whole team.', de: 'Wer wann weg ist, für das ganze Team.' } },
  { id: 'presence', name: { en: 'Live presence', de: 'Anwesenheit' }, description: { en: 'Who is in the office, at home or away.', de: 'Wer im Büro, zu Hause oder unterwegs ist.' } },
  { id: 'branding', name: { en: 'Branding', de: 'Branding' }, description: { en: 'Own name, logo and colours for the instance.', de: 'Eigener Name, Logo und Farben für die Instanz.' } },
  { id: 'connected-accounts', name: { en: 'Connected accounts', de: 'Verbundene Konten' }, description: { en: 'Each member links their own account at an external service.', de: 'Jedes Mitglied verbindet sein eigenes Konto bei einem externen Dienst.' } },
  { id: 'workplace-search', name: { en: 'Workplace search', de: 'Arbeitsplatz-Suche' }, description: { en: 'Confluence, SharePoint, Notion, Slack and Teams in the search field.', de: 'Confluence, SharePoint, Notion, Slack und Teams im Suchfeld.' } },
  { id: 'chat-apps', name: { en: 'Chat apps', de: 'Chat-Apps' }, description: { en: 'The start page in Slack and Microsoft Teams.', de: 'Die Startseite in Slack und Microsoft Teams.' } },
];
