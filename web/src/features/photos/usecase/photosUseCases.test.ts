import { describe, expect, it, vi } from 'vitest';
import type { Photo } from '../domain/photos';
import { type PhotosGateway, PhotosUseCases } from './photosUseCases';

const photo = (id: string): Photo => ({ id, album: 'A', width: 1, height: 1, size: 1 });

describe('PhotosUseCases', () => {
  it('uploads one file after another and keeps going after a failure', async () => {
    const upload = vi.fn(async (_album: string, _file: Blob, name: string) => {
      if (name === 'bad.svg') throw new Error('unsupported');
      return photo(name);
    });
    const uc = new PhotosUseCases({ upload } as unknown as PhotosGateway);
    const progress: number[] = [];
    const files = ['a.jpg', 'bad.svg', 'c.jpg'].map((n) => new File(['x'], n));
    const r = await uc.upload('A', files, (n) => progress.push(n));
    expect(r.added.map((p) => p.id)).toEqual(['a.jpg', 'c.jpg']);
    expect(r.failed.map((f) => f.name)).toEqual(['bad.svg']);
    expect(progress).toEqual([1, 2, 3]);
  });

  it('sends the order as ids', async () => {
    const reorder = vi.fn(async () => {});
    await new PhotosUseCases({ reorder } as unknown as PhotosGateway).reorder('A', [photo('b'), photo('a')]);
    expect(reorder).toHaveBeenCalledWith('A', ['b', 'a']);
  });
});
