import { detectTool, type ToolEnv } from '../features/search/domain/tools/detect';
import { runTool, type ToolResult } from '../features/search/domain/tools/run';
import type { Answer } from '../features/search/domain/smart/answer';
import type { SearchSource } from '../features/search/usecase/smartSearch';
import { intlLocale, t } from '../shared/i18n';

/** The browser's randomness and clock; getRandomValues works over plain HTTP too (#310). */
const browserEnv: ToolEnv = {
  random: (n) => crypto.getRandomValues(new Uint8Array(n)),
  now: () => new Date(),
};

/**
 * Small tools in the search field (#477): password, UUID, colour, Unix time
 * and subnet. Detection is synchronous and tiny; the tool itself is imported
 * when a query first needs it (lazy). Once an answer for a query exists it
 * is also returned by direct, so Enter copies exactly what is on screen.
 * Nothing here reaches the search history: copying records no choice, and
 * the generated values are never used as a target.
 */
export function toolsSource(env: ToolEnv = browserEnv): SearchSource {
  // Only the latest query is kept, so an old password does not linger.
  let last: { query: string; answer: Answer } | null = null;

  async function make(query: string): Promise<Answer | null> {
    const name = detectTool(query);
    const result = name ? await runTool(name, query, env) : null;
    if (!result) return null;
    const answer: Answer = { ...toAnswer(result), at: null };
    if (result.kind === 'password' || result.kind === 'uuid')
      answer.regenerate = async () => (await make(query)) ?? answer;
    last = { query, answer };
    return answer;
  }

  return {
    id: 'tools',
    direct: (query) => (last?.query === query ? last.answer : null),
    lazy: (query) => (last?.query === query || !detectTool(query) ? null : make(query)),
  };
}

// Omit per member of the union; a plain Omit would keep only the common fields.
type Body = Answer extends infer A ? (A extends Answer ? Omit<A, 'at'> : never) : never;

function toAnswer(r: ToolResult): Body {
  switch (r.kind) {
    case 'password':
      return {
        kind: 'metric',
        intent: 'tools.password',
        title: t('tools.password', { length: r.length }),
        value: r.value,
        detail: `${t(r.digitsOnly ? 'tools.password_digits' : r.symbols ? 'tools.password_all' : 'tools.password_no_symbols')}, ${t('tools.password_note')}`,
        copy: r.value,
      };
    case 'uuid':
      return {
        kind: 'metric',
        intent: 'tools.uuid',
        title: t('tools.uuid', { version: r.version }),
        value: r.value,
        copy: r.value,
      };
    case 'colour':
      return {
        kind: 'metric',
        intent: 'tools.colour',
        title: t('tools.colour'),
        value: r.hex,
        detail: `${r.rgb} · ${r.hsl} · ${r.oklch}`,
        swatch: r.css,
        copy: r.hex,
      };
    case 'timestamp': {
      const local = new Intl.DateTimeFormat(intlLocale(), { dateStyle: 'full', timeStyle: 'medium' }).format(
        r.date,
      );
      const utc = r.date.toISOString().replace('.000Z', 'Z');
      if (r.direction === 'from-unix')
        return {
          kind: 'metric',
          intent: 'tools.timestamp',
          title: t('tools.unix_to_date', { seconds: r.seconds }),
          value: local,
          detail: `${utc} (UTC)`,
          copy: utc,
        };
      return {
        kind: 'metric',
        intent: 'tools.timestamp',
        title: t('tools.date_to_unix'),
        value: String(r.seconds),
        detail: `${local} (${t('tools.local')}) · ${utc}`,
        copy: String(r.seconds),
      };
    }
    case 'subnet': {
      const count = BigInt(r.hosts).toLocaleString(intlLocale());
      const items = [
        ...(r.mask ? [{ label: t('tools.mask'), value: `${r.mask} (/${r.prefix})` }] : []),
        ...(r.broadcast ? [{ label: t('tools.broadcast'), value: r.broadcast }] : []),
        { label: t('tools.first_host'), value: r.firstHost },
        { label: t('tools.last_host'), value: r.lastHost },
        { label: t(r.family === 4 ? 'tools.hosts' : 'tools.addresses'), value: count },
      ];
      return {
        kind: 'list',
        intent: 'tools.subnet',
        title: t('tools.network', { network: r.network }),
        items,
        empty: '',
        copy: r.network,
      };
    }
  }
}
