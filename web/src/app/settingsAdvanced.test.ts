import { describe, expect, it } from 'vitest';
import type { MessageKey } from '../shared/i18n';
import { type CatalogSection, OWN_SECTIONS, type SectionViewer, visibleSections } from './settingsCatalog';
import { registeredSections } from './settingsSections';
import { groupByArea } from './settingsNav';
import { buildIndex, searchSettings } from './settingsSearch';
import { advancedProblem } from './settingsAdvancedHint';
import type { AttentionItem } from '../features/attention/domain/attention';

const all: CatalogSection[] = [...OWN_SECTIONS, ...registeredSections];
const owner = (advanced?: boolean): SectionViewer => ({
  modules: new Set(['appearance', 'system', 'apiwidget']),
  dataKinds: {},
  profile: 'homelab',
  localMode: false,
  guest: false,
  member: false,
  admin: true,
  owner: true,
  advanced,
});
const ids = (v: SectionViewer) => visibleSections(all, v).map((s) => s.id);

describe('advanced settings (#602)', () => {
  it('marks the documented initial list', () => {
    expect(
      all
        .filter((s) => s.advanced)
        .map((s) => s.id)
        .sort(),
    ).toEqual(['apiwidget', 'forecast', 'tokens']);
    expect(all.find((s) => s.id === 'appearance')?.advancedFields).toEqual([
      'appearance.custom_css_override_variables_such_as',
    ]);
  });

  it('hides advanced sections until they are shown', () => {
    for (const hidden of [undefined, false]) {
      expect(ids(owner(hidden))).not.toEqual(expect.arrayContaining(['tokens']));
      expect(ids(owner(hidden))).not.toContain('forecast');
      expect(ids(owner(hidden))).not.toContain('apiwidget');
    }
    expect(ids(owner(true))).toEqual(expect.arrayContaining(['tokens', 'forecast', 'apiwidget']));
  });

  it('still applies the role rules to shown advanced sections', () => {
    expect(ids({ ...owner(true), owner: false })).not.toContain('tokens');
  });

  it('finds advanced sections and fields only while they are shown', () => {
    const texts: Record<string, string> = {
      'settings.section.tokens': 'API tokens',
      'appearance.custom_css_override_variables_such_as': 'Custom CSS',
    };
    const t = (k: MessageKey) => texts[k] ?? k;
    const search = (advanced: boolean, q: string) => {
      const areas = groupByArea(visibleSections(all, owner(advanced)), (s) => t(s.label));
      const index = buildIndex(areas, {
        ui: t,
        others: [],
        sectionName: (s) => t(s.label),
        sectionLabel: (s) => s.label,
        advanced,
      });
      return searchSettings(index, q).map((h) => h.title);
    };
    expect(search(false, 'api tokens')).toEqual([]);
    expect(search(true, 'api tokens')).toContain('API tokens');
    expect(search(false, 'custom css')).toEqual([]);
    expect(search(true, 'custom css')).toContain('Custom CSS');
  });
});

describe('hint for a hidden advanced section with a problem (#622)', () => {
  const failing = (integration: string, kind: AttentionItem['kind'] = 'connection'): AttentionItem => ({
    id: `${integration}-${kind}`,
    kind,
    integration,
    reason: { message: 'down', messageKey: 'srv_x' },
    since: new Date(0),
    action: kind === 'connection' ? 'fix' : 'choose',
  });
  const sections = visibleSections(all, owner(true));
  const hint = (over: Partial<Parameters<typeof advancedProblem>[0]> = {}) =>
    advancedProblem({ items: [failing('apiwidget')], sections, showAdvanced: false, admin: true, ...over })
      ?.id;

  it('points at the advanced section whose connection fails', () => {
    expect(hint()).toBe('apiwidget');
  });

  it('stays away while advanced settings are shown or for non-admins', () => {
    expect(hint({ showAdvanced: true })).toBeUndefined();
    expect(hint({ admin: false })).toBeUndefined();
  });

  it('ignores problems outside advanced sections and entries still to set up', () => {
    expect(hint({ items: [failing('weather')] })).toBeUndefined();
    expect(hint({ items: [failing('apiwidget', 'setup')] })).toBeUndefined();
    expect(hint({ items: [] })).toBeUndefined();
  });

  it('only offers sections the viewer may open', () => {
    expect(hint({ sections: visibleSections(all, { ...owner(true), modules: new Set() }) })).toBeUndefined();
  });
});
