/** A stored picture (#286); the image is fetched by its id. */
export interface Photo {
  id: string;
  album: string;
  width: number;
  height: number;
  size: number;
}

export interface PhotoAlbum {
  name: string;
  count: number;
  bytes: number;
}

/** Albums and the storage used against its limits. */
export interface PhotoOverview {
  albums: PhotoAlbum[];
  count: number;
  usedBytes: number;
  quotaBytes: number;
  maxPhotos: number;
  maxUploadBytes: number;
}

/** How a picture fills the widget: whole with bars, or cropped to fill. */
export type PhotoFit = 'contain' | 'cover';

export interface PhotoFrameConfig {
  album: string;
  /** Seconds between two pictures. */
  interval: number;
  fit: PhotoFit;
}

export const DEFAULT_ALBUM = 'Photos';
export const MIN_INTERVAL = 5;
export const MAX_INTERVAL = 3600;
export const DEFAULT_INTERVAL = 30;
export const MAX_ALBUM_LENGTH = 40;

/** Reads the widget configuration with safe defaults. */
export function parseConfig(raw: Record<string, unknown>): PhotoFrameConfig {
  const album =
    typeof raw['album'] === 'string' && raw['album'].trim()
      ? raw['album'].trim().slice(0, MAX_ALBUM_LENGTH)
      : DEFAULT_ALBUM;
  const i = raw['interval'];
  const interval =
    typeof i === 'number' && Number.isInteger(i) && i >= MIN_INTERVAL && i <= MAX_INTERVAL
      ? i
      : DEFAULT_INTERVAL;
  const fit: PhotoFit = raw['fit'] === 'cover' ? 'cover' : 'contain';
  return { album, interval, fit };
}

/** Index of the next picture, wrapping around; 0 for an empty album. */
export const nextIndex = (current: number, count: number): number => (count > 0 ? (current + 1) % count : 0);

/** Moves the picture at index by delta places; returns null at the edges. */
export function move<T>(list: readonly T[], index: number, delta: number): T[] | null {
  const to = index + delta;
  if (index < 0 || index >= list.length || to < 0 || to >= list.length) return null;
  const out = [...list];
  const [item] = out.splice(index, 1);
  if (item !== undefined) out.splice(to, 0, item);
  return out;
}

/** Share of the storage in use, 0–100. */
export const usedPercent = (o: Pick<PhotoOverview, 'usedBytes' | 'quotaBytes'>): number =>
  o.quotaBytes > 0 ? Math.min(100, Math.round((o.usedBytes / o.quotaBytes) * 100)) : 0;
