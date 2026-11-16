import { describe, expect, it } from 'vitest';
import { definitionFor } from '../widgets';
import { MemoryDocumentStore } from '../../shared/local/documentStore';
import { createLocalContainer } from './localContainer';
import { localModules, localWidgetKinds } from './modules';

describe('createLocalContainer', () => {
  it('offers only the modules that work without a server', async () => {
    const c = createLocalContainer(new MemoryDocumentStore());
    const enabled = await c.enabledModules();
    expect([...enabled.modules].sort()).toEqual([...localModules].sort());
    expect(enabled.modules.has('docker')).toBe(false);
    expect(enabled.widgetKinds.has('radar')).toBe(false);
  });

  it('marks itself as local mode, so the UI can say what a server adds (#291)', () => {
    expect(createLocalContainer(new MemoryDocumentStore()).localMode).toBe(true);
  });

  it('signs in as the owner', async () => {
    const c = createLocalContainer(new MemoryDocumentStore());
    expect(await c.auth.currentView()).toBe('app');
    expect(c.auth.admin).toBe(true);
  });
});

describe('local dashboard', () => {
  it('loads without a server', async () => {
    const c = createLocalContainer(new MemoryDocumentStore());
    expect((await c.dashboard.load())[0]?.name).toBe('Home');
  });
});

describe('local widget kinds', () => {
  it('are all registered widgets', () => {
    for (const kind of localWidgetKinds) expect(definitionFor(kind), kind).toBeDefined();
  });
});
