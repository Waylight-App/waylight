import { describe, expect, it, vi } from 'vitest';
import { type HelpGateway, HelpUseCases } from './helpUseCases';

const gateway = (over: Partial<HelpGateway> = {}): HelpGateway => ({
  widgets: vi
    .fn()
    .mockResolvedValue([
      {
        kind: 'calendar',
        shows: 'Appointments.',
        setUp: 'Connect a calendar.',
        docsUrl: 'https://d/calendar/',
      },
    ]),
  docs: vi.fn().mockResolvedValue({
    baseUrl: 'https://d/',
    pages: [{ path: 'modules/calendar/', title: 'Calendar', summary: '', sections: [] }],
  }),
  ...over,
});
const describeKind = (kind: string) =>
  kind === 'calendar' ? { title: 'Calendar', keywords: [] } : undefined;

describe('HelpUseCases (#678)', () => {
  it('loads the widget help once', async () => {
    const g = gateway();
    const uc = new HelpUseCases(g);
    expect((await uc.widget('calendar'))?.shows).toBe('Appointments.');
    expect(await uc.widget('weather')).toBeUndefined();
    expect(g.widgets).toHaveBeenCalledTimes(1);
  });

  it('retries after a failed load', async () => {
    const widgets = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValue([]);
    const uc = new HelpUseCases(gateway({ widgets }));
    await expect(uc.widgets()).rejects.toThrow('offline');
    expect(await uc.widgets()).toEqual([]);
  });

  it('answers with the widget first, then the documentation', async () => {
    const uc = new HelpUseCases(gateway());
    const topics = await uc.answer('How do I add a calendar?', describeKind);
    expect(topics.map((t) => t.id)).toEqual(['widget:calendar', 'docs:modules/calendar/']);
  });

  it('still answers from the widgets without the documentation index', async () => {
    const uc = new HelpUseCases(gateway({ docs: vi.fn().mockRejectedValue(new Error('x')) }));
    expect((await uc.answer('calendar?', describeKind)).map((t) => t.id)).toEqual(['widget:calendar']);
  });
});
