import { describe, expect, it, vi } from 'vitest';
import { PlacesUseCases, type PlacesGateway } from './placesUseCases';

describe('PlacesUseCases', () => {
  it('trims what is saved and keeps the coordinates on rename', async () => {
    const g = {
      list: vi.fn(),
      add: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
    } satisfies PlacesGateway;
    const uc = new PlacesUseCases(g);
    await uc.add({ name: ' Home ', label: ' Oldenburg ', lat: 53.1, lon: 8.2 });
    expect(g.add).toHaveBeenCalledWith({ name: 'Home', label: 'Oldenburg', lat: 53.1, lon: 8.2 });
    await uc.rename(
      { id: 'p1', name: 'Home', label: 'Oldenburg', lat: 53.1, lon: 8.2, usedBy: [] },
      ' Flat ',
    );
    expect(g.update).toHaveBeenCalledWith('p1', { name: 'Flat', label: 'Oldenburg', lat: 53.1, lon: 8.2 });
  });
});
