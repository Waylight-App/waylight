import { describe, expect, it } from 'vitest';
import type { MessageKey } from '../shared/i18n';
import { buildIndex, normalise, searchSettings } from './settingsSearch';
import { textMatchesLabel } from './settingsHighlight';
import { groupByArea } from './settingsNav';
import { type CatalogSection, OWN_SECTIONS, type SectionViewer, visibleSections } from './settingsCatalog';
import { registeredSections } from './settingsSections';

// Small fake catalogues: the UI speaks English, German is the other language.
const EN: Record<string, string> = {
  'settings.area.account': 'My account',
  'settings.area.connections': 'Connections & services',
  'settings.area.people': 'People & access',
  'settings.area.appearance': 'Appearance & language',
  'settings.section.account': 'Account',
  'settings.section.members': 'Members',
  'settings.section.immich': 'Immich',
  'settings.section.appearance': 'Appearance',
  'app.change_password': 'Change password',
  'app.two_factor_authentication': 'Two-factor authentication',
  'appearance.greeting_name': 'Your name in the greeting (optional)',
};
const DE: Record<string, string> = {
  'settings.area.account': 'Mein Konto',
  'settings.area.connections': 'Zugänge & Dienste',
  'settings.section.account': 'Konto',
  'app.change_password': 'Passwort ändern',
  'app.two_factor_authentication': 'Zwei-Faktor-Authentifizierung',
  'appearance.greeting_name': 'Dein Name in der Begrüßung (optional)',
};
const en = (k: MessageKey) => EN[k] ?? k;
const de = (k: MessageKey) => DE[k] ?? k;

function indexFor(viewer: Partial<SectionViewer> = {}, others = [de]) {
  const v: SectionViewer = {
    modules: new Set(['appearance', 'immich']),
    dataKinds: {},
    profile: 'homelab',
    localMode: false,
    guest: false,
    member: false,
    admin: true,
    owner: true,
    ...viewer,
  };
  const sections = visibleSections<CatalogSection>([...OWN_SECTIONS, ...registeredSections], v);
  const areas = groupByArea(sections, (s) => en(s.label));
  return buildIndex(areas, {
    ui: en,
    others,
    sectionName: (s) => en(s.label),
    sectionLabel: (s) => s.label,
  });
}
const titles = (q: string, viewer?: Partial<SectionViewer>) =>
  searchSettings(indexFor(viewer), q).map((h) => `${h.kind}:${h.title}`);

describe('settings search (#601)', () => {
  it('finds areas, sections and fields by name', () => {
    expect(titles('account')).toEqual(expect.arrayContaining(['area:My account', 'section:Account']));
    expect(titles('change pass')[0]).toBe('field:Change password');
    expect(titles('immich')).toContain('section:Immich');
  });

  it('finds through synonyms', () => {
    expect(titles('password')).toContain('section:Account');
    expect(titles('dark')).toContain('section:Appearance');
  });

  it('finds German words in an English interface and ignores accents', () => {
    expect(titles('passwort')).toEqual(expect.arrayContaining(['section:Account', 'field:Change password']));
    expect(titles('zugange')).toContain('area:Connections & services');
    expect(titles('begrüßung')).toContain('field:Your name in the greeting (optional)');
  });

  it('works in the UI language alone while the other catalogue is not loaded', () => {
    expect(searchSettings(indexFor({}, []), 'zwei').map((h) => h.title)).toEqual([]);
    expect(searchSettings(indexFor({}, []), 'two-factor')[0]?.title).toBe('Two-factor authentication');
  });

  it('describes where a field lives and which section opens', () => {
    const [hit] = searchSettings(indexFor(), 'change password');
    expect(hit).toMatchObject({
      kind: 'field',
      section: 'account',
      area: 'account',
      path: 'My account › Account',
      field: 'Change password',
    });
  });

  it('only finds sections the viewer may open', () => {
    expect(titles('members')).toContain('section:Members');
    expect(titles('members', { member: true, admin: false, owner: false })).not.toContain('section:Members');
    expect(titles('immich', { guest: true, admin: false, owner: false })).toEqual([]);
  });

  it('needs every word and ignores empty queries', () => {
    expect(titles('change zebra')).toEqual([]);
    expect(titles('   ')).toEqual([]);
  });

  it('normalises case, accents and spaces', () => {
    expect(normalise('  Zugänge   &  DIENSTE ')).toBe('zugange & dienste');
  });
});

describe('field highlight (#601)', () => {
  it('matches a label with or without a trailing hint', () => {
    expect(textMatchesLabel('Change password', 'Change password')).toBe(true);
    expect(textMatchesLabel('Language  (browser)', 'Language')).toBe(true);
    expect(textMatchesLabel('Languages', 'Language')).toBe(false);
    expect(textMatchesLabel('Anything', '')).toBe(false);
  });
});
