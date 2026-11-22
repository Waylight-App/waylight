import { describe, expect, it } from 'vitest';
import { SEARCH_EXAMPLES } from '../features/onboarding/domain/wizard';
import { interpret } from '../features/search/domain/smart/intent';
import { INTENTS } from '../features/search/domain/smart/intents';
import { THRESHOLD } from '../features/search/domain/smart/intent';
import { contributions } from './searchSources';

const intents = [...INTENTS, ...contributions.flatMap((c) => c.intents)];

describe('wizard search examples', () => {
  // The lists are plain data; this keeps them honest when intents change.
  it('are all understood by the answer engine', () => {
    for (const sets of Object.values(SEARCH_EXAMPLES))
      for (const examples of Object.values(sets))
        for (const { query } of examples) {
          const best = interpret(query, intents, [], new Date()).filter((m) => m.score >= THRESHOLD);
          expect(best.length, query).toBeGreaterThan(0);
        }
  });
});
