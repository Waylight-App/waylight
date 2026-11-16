import type { PluginSearchAction, Texts } from '../../../../web/src/app/plugins/api';
import { PresenceApi, settable, type Status } from './presence';
import { texts } from './texts';

/**
 * Words that offer each status in the start page search (#453), English and
 * German. The core matches them from two letters on, so "hom" already
 * offers "home".
 */
const keywords: Record<Exclude<Status, ''>, string[]> = {
  office: ['office', 'büro', 'buero'],
  home: ['home', 'homeoffice', 'home office', 'remote'],
  away: ['away', 'weg', 'abwesend'],
};

/** "Set my status: Office", in every language of the status name. */
function withStatus(template: Texts, status: Texts): Texts {
  const out: Texts = { en: template.en.replace('{status}', status.en) };
  for (const [lang, t] of Object.entries(template)) {
    out[lang] = t.replace('{status}', status[lang] ?? status.en);
  }
  return out;
}

/** One action per status a person can choose. */
export function presenceSearchActions(): PluginSearchAction[] {
  return settable.map((status) => {
    const name = texts[status] ?? { en: status };
    return {
      id: `presence-${status}`,
      feature: 'presence',
      keywords: keywords[status],
      title: withStatus(texts.search_set ?? { en: 'Set my status: {status}' }, name),
      detail: texts.title ?? { en: 'Presence' },
      done: withStatus(texts.search_done ?? { en: 'Status set: {status}' }, name),
      run: (context) => new PresenceApi(context.http).set(status),
    };
  });
}
