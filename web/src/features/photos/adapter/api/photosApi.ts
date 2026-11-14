import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, PhotoDTO } from '../../../../shared/api-client/generated/photos';
import type { Photo, PhotoOverview } from '../../domain/photos';
import type { PhotosGateway } from '../../usecase/photosUseCases';

const q = (album: string) => `album=${encodeURIComponent(album)}`;

/** PhotosGateway backed by the REST API. */
export class PhotosApi implements PhotosGateway {
  constructor(private readonly http: HttpClient) {}

  async list(album: string): Promise<Photo[]> {
    return (await this.http.get<PhotoDTO[]>(`/api/photos?${q(album)}`)).map((p) => ({ ...p }));
  }
  async overview(): Promise<PhotoOverview> {
    const o = await this.http.get<OverviewDTO>('/api/photos/albums');
    return { ...o, albums: o.albums.map((a) => ({ ...a })) };
  }
  async upload(album: string, file: Blob, name: string): Promise<Photo> {
    const form = new FormData();
    form.append('file', file, name);
    return { ...(await this.http.postForm<PhotoDTO>(`/api/photos?${q(album)}`, form)) };
  }
  async reorder(album: string, ids: string[]): Promise<void> {
    await this.http.put('/api/photos/order', { album, ids });
  }
  async remove(id: string): Promise<void> {
    await this.http.delete(`/api/photos/${encodeURIComponent(id)}`);
  }
  imageUrl(id: string, thumb: boolean): string {
    return `/api/photos/${encodeURIComponent(id)}/image${thumb ? '?thumb=1' : ''}`;
  }
}
