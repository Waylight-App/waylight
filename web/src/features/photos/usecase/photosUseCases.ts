import type { Photo, PhotoOverview } from '../domain/photos';

/** Output port: the photo frame API. */
export interface PhotosGateway {
  list(album: string): Promise<Photo[]>;
  overview(): Promise<PhotoOverview>;
  upload(album: string, file: Blob, name: string): Promise<Photo>;
  reorder(album: string, ids: string[]): Promise<void>;
  remove(id: string): Promise<void>;
  imageUrl(id: string, thumb: boolean): string;
}

/** Result of uploading several files: what worked and the first problem. */
export interface UploadResult {
  added: Photo[];
  failed: { name: string; error: unknown }[];
}

/** Albums of uploaded pictures for the photo frame widget (#286). */
export class PhotosUseCases {
  constructor(private readonly gateway: PhotosGateway) {}

  list(album: string): Promise<Photo[]> {
    return this.gateway.list(album);
  }
  overview(): Promise<PhotoOverview> {
    return this.gateway.overview();
  }

  /**
   * Uploads the files one after another: the server scales each picture,
   * and one request at a time keeps its memory use flat. A failing file
   * does not stop the others.
   */
  async upload(
    album: string,
    files: readonly File[],
    onProgress?: (done: number) => void,
  ): Promise<UploadResult> {
    const result: UploadResult = { added: [], failed: [] };
    for (const [i, f] of files.entries()) {
      try {
        result.added.push(await this.gateway.upload(album, f, f.name));
      } catch (error) {
        result.failed.push({ name: f.name, error });
      }
      onProgress?.(i + 1);
    }
    return result;
  }

  reorder(album: string, photos: readonly Photo[]): Promise<void> {
    return this.gateway.reorder(
      album,
      photos.map((p) => p.id),
    );
  }
  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
  imageUrl(id: string, thumb = false): string {
    return this.gateway.imageUrl(id, thumb);
  }
}
