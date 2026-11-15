import { describe, expect, it } from 'vitest';
import { definitionFor } from '../widgets';
import { createLocalContainer } from './localContainer';
import { localModules, localWidgetKinds } from './modules';

describe('createLocalContainer', () => {
  it('offers only the modules that work without a server', async () => {
    const c = createLocalContainer();
    const enabled = await c.enabledModules();
    expect([...enabled.modules].sort()).toEqual([...localModules].sort());
    expect(enabled.modules.has('docker')).toBe(false);
    expect(enabled.widgetKinds.has('radar')).toBe(false);
  });

  it('signs in as the owner', async () => {
    const c = createLocalContainer();
    expect(await c.auth.currentView()).toBe('app');
    expect(c.auth.admin).toBe(true);
  });
});

describe('local widget kinds', () => {
  it('are all registered widgets', () => {
    for (const kind of localWidgetKinds) expect(definitionFor(kind), kind).toBeDefined();
  });
});
