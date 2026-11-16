import { describe, expect, it, vi } from 'vitest';
import type { PluginContext } from '../../../../web/src/app/plugins/api';
import { presenceSearchActions } from './searchActions';

describe('presence search actions (#453)', () => {
  const actions = presenceSearchActions();

  it('offers office, home and away behind the presence feature', () => {
    expect(actions.map((a) => a.id)).toEqual(['presence-office', 'presence-home', 'presence-away']);
    expect(actions.every((a) => a.feature === 'presence')).toBe(true);
  });

  it('knows the English and German words of the issue', () => {
    const words = actions.flatMap((a) => a.keywords);
    for (const w of ['office', 'home', 'away', 'büro', 'homeoffice', 'weg']) expect(words).toContain(w);
  });

  it('names the status in the title and the confirmation, per language', () => {
    const [office] = actions;
    expect(office?.title).toMatchObject({ en: 'Set my status: Office', de: 'Meinen Status setzen: Büro' });
    expect(office?.done).toMatchObject({ en: 'Status set: Office', de: 'Status gesetzt: Büro' });
  });

  it('sets the own status through PUT /api/presence/me', async () => {
    const put = vi.fn(() => Promise.resolve());
    const context = { http: { put } } as unknown as PluginContext;
    await actions[1]?.run(context);
    expect(put).toHaveBeenCalledWith('/api/presence/me', { status: 'home', returnOn: '' });
  });
});
