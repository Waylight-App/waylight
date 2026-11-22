import { connectionIdOf, withConnection } from '../../../shared/connection/connection';

/** Connection settings as shown in the UI (the API key is never returned). */
export interface ImmichSettings {
  configured: boolean;
  url: string;
  hasApiKey: boolean;
}

/** "On this day" some years ago; thumbnails are proxied by the backend. */
export interface Memory {
  yearsAgo: number;
  assetIds: string[];
}

/** What the widget shows. */
export interface ImmichOverview {
  version: string;
  photos: number;
  videos: number;
  diskUsed: number;
  diskTotal: number;
  diskPercent: number;
  memories: Memory[];
}

/** URL of the proxied thumbnail of a memory photo. */
export function thumbnailUrl(assetId: string, connection?: string): string {
  return withConnection(`/api/immich/thumbnails/${encodeURIComponent(assetId)}`, connection);
}

/** Widget configuration. */
export interface ImmichWidgetConfig {
  showMemories: boolean;
  /** Named connection to read through (#709); empty reads the module settings. */
  connectionId: string;
}

/** Reads the widget configuration; memories are shown by default. */
export function parseWidgetConfig(raw: Record<string, unknown>): ImmichWidgetConfig {
  return { showMemories: raw['showMemories'] !== false, connectionId: connectionIdOf(raw) };
}
