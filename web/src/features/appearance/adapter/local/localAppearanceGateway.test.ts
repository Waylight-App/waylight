import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { AppearanceNeedsServerError, LocalAppearanceGateway } from './localAppearanceGateway';

describe('LocalAppearanceGateway', () => {
  it('starts with the everyday start page', async () => {
    expect((await new LocalAppearanceGateway(new MemoryDocumentStore()).load()).profile).toBe('everyday');
  });

  it('stores preferences and the background', async () => {
    const store = new MemoryDocumentStore();
    const gateway = new LocalAppearanceGateway(store);
    await gateway.save({ ...(await gateway.load()), theme: 'dark', greetingName: 'Sam' });
    await gateway.setBackground('https://example.org/sky.jpg');
    const loaded = await new LocalAppearanceGateway(store).load();
    expect(loaded).toMatchObject({
      theme: 'dark',
      greetingName: 'Sam',
      backgroundImage: 'https://example.org/sky.jpg',
    });
  });

  it('has no uploads or theme gallery without a server', async () => {
    const gateway = new LocalAppearanceGateway(new MemoryDocumentStore());
    expect((await gateway.gallery()).themes).toEqual([]);
    expect((await gateway.backgrounds()).items).toEqual([]);
    await expect(gateway.applyTheme()).rejects.toBeInstanceOf(AppearanceNeedsServerError);
  });
});
