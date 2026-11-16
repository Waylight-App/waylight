import { describe, expect, it, vi } from 'vitest';
import { AbsencesApi, addDays, sections, validate, type Absence } from './absences';

const absence = (name: string, from: string, to: string): Absence => ({
  id: name,
  accountId: name,
  name,
  from,
  to,
  kind: 'holiday',
  note: '',
  private: false,
  editable: false,
});

describe('absences', () => {
  it('adds days across months', () => {
    expect(addDays('2026-11-30', 1)).toBe('2026-12-01');
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28');
  });

  it('splits into today and later this week', () => {
    const list = [
      absence('Ben', '2026-11-10', '2026-11-16'),
      absence('Anna', '2026-11-16', '2026-11-16'),
      absence('Cleo', '2026-11-18', '2026-11-20'),
      absence('Dora', '2026-11-25', '2026-11-26'),
      absence('Old', '2026-11-01', '2026-11-15'),
    ];
    const s = sections(list, '2026-11-16');
    expect(s.today.map((a) => a.name)).toEqual(['Ben', 'Anna']);
    expect(s.week.map((a) => a.name)).toEqual(['Cleo']);
  });

  it('validates the period and the note', () => {
    expect(validate({ from: '2026-11-16', to: '2026-11-20', note: '' })).toBeNull();
    expect(validate({ from: '2026-11-20', to: '2026-11-16', note: '' })).toBe('dates');
    expect(validate({ from: '2026-11-16', to: '2027-11-17', note: '' })).toBe('too_long');
    expect(validate({ from: '2026-11-16', to: '2026-11-16', note: 'x'.repeat(201) })).toBe('note');
  });

  it('sends accountId only for new absences entered for someone else', async () => {
    const http = { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn(), putBlob: vi.fn() };
    const api = new AbsencesApi(http);
    const input = {
      accountId: '',
      from: '2026-11-16',
      to: '2026-11-17',
      kind: 'sick' as const,
      note: ' flu ',
      private: true,
    };
    await api.save(input);
    expect(http.post).toHaveBeenCalledWith('/api/absences', {
      from: '2026-11-16',
      to: '2026-11-17',
      kind: 'sick',
      note: 'flu',
      private: true,
    });
    await api.save({ ...input, accountId: 'a1' });
    expect(http.post).toHaveBeenLastCalledWith('/api/absences', expect.objectContaining({ accountId: 'a1' }));
    await api.save({ ...input, accountId: 'a1' }, 'x 1');
    expect(http.put).toHaveBeenCalledWith(
      '/api/absences/x%201',
      expect.not.objectContaining({ accountId: 'a1' }),
    );
  });
});
