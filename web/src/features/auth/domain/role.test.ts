import { describe, expect, it } from 'vitest';
import { en } from '../../../shared/i18n/locales/en';
import { de } from '../../../shared/i18n/locales/de';
import { roleTexts } from './role';

describe('roleTexts', () => {
  it('has a label and a hint in both languages for every role', () => {
    for (const { label, hint } of Object.values(roleTexts)) {
      for (const key of [label, hint]) {
        expect(en[key]).toBeTruthy();
        expect(de[key]).toBeTruthy();
      }
    }
  });
});
