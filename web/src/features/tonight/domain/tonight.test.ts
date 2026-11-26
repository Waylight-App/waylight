import { describe, expect, it } from 'vitest';
import {
  sentenceCase,
  degrees,
  formatDate,
  formatTime,
  parseTonightConfig,
  type TonightWarning,
  warningLevel,
  worstWarning,
} from './tonight';

const warning = (id: string, severity: TonightWarning['severity']): TonightWarning => ({
  id,
  title: id,
  validity: '',
  severity,
});

describe('parseTonightConfig', () => {
  it('uses the device zone and a 24-hour clock by default', () => {
    expect(parseTonightConfig({})).toEqual({ timeZone: '', hour12: false });
  });

  it('keeps a valid zone and drops an unknown one', () => {
    expect(parseTonightConfig({ timeZone: 'Asia/Tokyo', hour12: true })).toEqual({
      timeZone: 'Asia/Tokyo',
      hour12: true,
    });
    expect(parseTonightConfig({ timeZone: 'Mars/Base' }).timeZone).toBe('');
  });
});

describe('worstWarning', () => {
  it('is undefined without warnings', () => {
    expect(worstWarning([])).toBeUndefined();
  });

  it('picks the most severe, the first on a tie', () => {
    const list = [warning('a', 'minor'), warning('b', 'severe'), warning('c', 'severe')];
    expect(worstWarning(list)?.id).toBe('b');
  });
});

describe('warningLevel', () => {
  it('asks for attention below severe and signals a problem from severe', () => {
    expect(warningLevel({ severity: 'moderate' })).toBe('attention');
    expect(warningLevel({ severity: 'severe' })).toBe('problem');
    expect(warningLevel({ severity: 'extreme' })).toBe('problem');
  });
});

describe('formatting', () => {
  const at = new Date('2026-11-23T18:05:00Z');
  const berlin = { timeZone: 'Europe/Berlin', hour12: false };

  it('formats the time in the configured zone', () => {
    expect(formatTime(at, berlin, 'en-GB')).toBe('19:05');
  });

  it('formats the date with the weekday', () => {
    expect(formatDate(at, berlin, 'en-GB')).toContain('Monday');
  });

  it('rounds degrees', () => {
    expect(degrees(7.6)).toBe('8°');
  });
});

describe('sentenceCase', () => {
  it('lowers a title sent in capitals', () => {
    expect(sentenceCase('STARKWIND')).toBe('Starkwind');
    expect(sentenceCase('FROST AND ICE')).toBe('Frost and ice');
  });
  it('keeps a title that already has lower-case letters', () => {
    expect(sentenceCase('Wind warning')).toBe('Wind warning');
  });
});
