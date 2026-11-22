import type { MediaOverview, MediaProvider, MediaSettings } from '../domain/media';

/** Output port: media backend. */
export interface MediaGateway {
  settings(): Promise<MediaSettings>;
  saveSettings(provider: MediaProvider, url: string, token: string): Promise<MediaSettings>;
  overview(connection?: string): Promise<MediaOverview>;
}

/** Application logic of the media server feature (#61). */
export class MediaUseCases {
  constructor(private readonly gateway: MediaGateway) {}

  settings(): Promise<MediaSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty token keeps the stored one. */
  saveSettings(provider: MediaProvider, url: string, token: string): Promise<MediaSettings> {
    return this.gateway.saveSettings(provider, url.trim(), token.trim());
  }

  /** Reads through the named connection; without one, through the module settings. */
  overview(connection?: string): Promise<MediaOverview> {
    return this.gateway.overview(connection);
  }
}
