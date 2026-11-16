import { describe, expect, it } from 'vitest';
import { text } from './texts';

describe('text', () => {
  const cat = { hello: { en: 'Hello {name}', de: 'Hallo {name}' } };
  it('picks the language, falls back to English and fills placeholders', () => {
    expect(text(cat, 'hello', 'de', { name: 'Anna' })).toBe('Hallo Anna');
    expect(text(cat, 'hello', 'fr', { name: 'Anna' })).toBe('Hello Anna');
    expect(text(cat, 'missing', 'en')).toBe('missing');
  });
});
