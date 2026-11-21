import { describe, expect, it } from 'vitest';
import { definitionFor, widgetDefinitions } from './widgets';

describe('widget registry (#116)', () => {
  it('loads every view and settings component lazily and caches the chunk', async () => {
    for (const def of widgetDefinitions) {
      for (const r of [def.view, def.settings]) {
        if (!r) continue;
        const first = await r.load();
        expect(first, def.type.kind).toBeTypeOf('function');
        expect(await r.load()).toBe(first);
      }
    }
  });

  it('has unique kinds', () => {
    const kinds = widgetDefinitions.map((d) => d.type.kind);
    expect(new Set(kinds).size).toBe(kinds.length);
  });

  it('declares compact views for the widgets with a compact state (#595)', () => {
    const modules = Object.fromEntries(
      widgetDefinitions.flatMap((d) => {
        const module = d.compact?.({});
        return module ? [[d.type.kind, module]] : [];
      }),
    );
    expect(modules).toEqual({ todos: 'todos', focus: 'focus', habits: 'habits' });
    // A connected to-do list has no compact state on the server.
    expect(definitionFor('todos')?.compact?.({ source: 'caldav-1' })).toBeUndefined();
  });
});
