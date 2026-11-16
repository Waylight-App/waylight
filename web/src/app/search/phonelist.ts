import { filterEntries } from '../../features/phonelist/domain/phonelist';
import { t } from '../../shared/i18n';
import { defineSearch, terms } from '../searchKit';

/**
 * The phone list in the search field (#300): "phone anna", "telefon weber"
 * or "nummer empfang" answers with the number and the e-mail address;
 * Enter copies the number.
 */
export const search = defineSearch({
  module: 'phonelist',
  intents: [
    {
      id: 'phonelist.person',
      source: 'phonelist',
      prefixes: ['phone', 'telefon', 'tel', 'nummer', 'number', 'durchwahl'],
      answer: 'text',
      widgetKind: 'phonelist',
      slots: [{ name: 'person', type: 'entity', entityKind: 'person', required: true }],
      terms: terms(1, 'phone', 'telefon', 'nummer', 'number', 'anrufen', 'call', 'durchwahl'),
    },
  ],
  source: (c) => ({
    id: 'phonelist',
    async entities() {
      // Names, plus role and team, so "phone reception" finds the person.
      return (await c.phonelist.list()).flatMap((e) =>
        [e.name, e.role, e.team].filter(Boolean).map((name) => ({ kind: 'person', id: e.id, name })),
      );
    },
    async resolve(m) {
      const list = await c.phonelist.list();
      const target = m.slots['person']?.type === 'entity' ? m.slots['person'].entity : null;
      const e = list.find((x) => x.id === target?.id) ?? filterEntries(list, target?.name ?? '')[0];
      const base = { intent: m.intent.id, widgetKind: 'phonelist', at: null };
      if (!e) return { ...base, kind: 'text', title: target?.name ?? '', text: t('phonelist.none_match') };
      const how = [e.phone, e.email].filter(Boolean).join(' · ');
      return {
        ...base,
        kind: 'text',
        title: [e.name, e.role].filter(Boolean).join(' · '),
        text: how || t('phonelist.no_contact'),
        copy: e.phone || e.email || undefined,
      };
    },
  }),
});
