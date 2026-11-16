import { describe, expect, it } from 'vitest';
import { type Entity, interpret, rank } from '../../features/search/domain/smart/intent';
import { allIntents } from '../searchSources';

/**
 * The help (#479) is generated from the intents, so every intent needs a
 * description and examples in German and English – and the examples have
 * to work, or the help would teach queries that answer something else.
 */
const INTENTS = allIntents();
const entities: Entity[] = [
  { kind: 'container', id: 'c1', name: 'jellyfin' },
  { kind: 'service', id: 's1', name: 'NAS' },
  { kind: 'job', id: 'j1', name: 'restic-offsite' },
  { kind: 'person', id: 'p1', name: 'Anna' },
];
const today = new Date(2026, 10, 12, 10, 0);
const top = (query: string) => rank(interpret(query, INTENTS, entities, today))[0];

describe('help of the intents (#479)', () => {
  it.each(INTENTS.map((i) => [i.id, i] as const))('%s has a description and examples', (_, intent) => {
    const help = intent.help;
    expect(help, 'help missing').toBeDefined();
    expect(help?.de.trim()).not.toBe('');
    expect(help?.en.trim()).not.toBe('');
    expect(help?.examples.de.length).toBeGreaterThan(0);
    expect(help?.examples.en.length).toBeGreaterThan(0);
  });

  const examples = INTENTS.flatMap((i) =>
    [...(i.help?.examples.de ?? []), ...(i.help?.examples.en ?? [])].map((q) => [q, i.id] as const),
  );
  it.each(examples)('the example "%s" answers %s', (query, id) => {
    expect(top(query)?.intent.id).toBe(id);
  });
});
