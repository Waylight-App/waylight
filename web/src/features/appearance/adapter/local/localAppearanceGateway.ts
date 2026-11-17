import type { DocumentStore } from '../../../../shared/local/documentStore';
import { defaultAppearance, type Appearance } from '../../domain/appearance';
import type { BackgroundLibrary, UploadedBackground } from '../../domain/backgrounds';
import type { ThemeGallery, ThemeInfo } from '../../domain/themes';
import type { AppearanceGateway } from '../../usecase/appearanceUseCases';

/** Raised for appearance features that need a server in the extension's local mode. */
export class AppearanceNeedsServerError extends Error {
  override name = 'AppearanceNeedsServerError';
  constructor() {
    super('This needs a waylight server.');
  }
}

const needsServer = (): Promise<never> => Promise.reject(new AppearanceNeedsServerError());

/**
 * The local mode starts with the everyday start page (#281): no home network
 * widgets, since those need a server anyway.
 */
const localDefault: Appearance = { ...defaultAppearance, profile: 'everyday' };

/**
 * Appearance gateway of the browser extension's local mode (ADR 0019).
 * Theme mode, accent colour, background by URL, own CSS, language, greeting
 * and launcher are stored locally. Uploaded pictures and the theme gallery
 * live on the server (scaling, built-in theme files), so the gallery is empty
 * and uploads reject with AppearanceNeedsServerError.
 */
export class LocalAppearanceGateway implements AppearanceGateway {
  constructor(private readonly store: DocumentStore) {}

  async load(): Promise<Appearance> {
    // Merged with the defaults, so fields added in a later version get a value.
    return { ...localDefault, ...(await this.store.get<Partial<Appearance>>('appearance')) };
  }

  async save(a: Appearance): Promise<Appearance> {
    await this.store.put('appearance', a);
    return a;
  }

  async setBackground(value: string): Promise<Appearance> {
    return this.save({ ...(await this.load()), backgroundImage: value });
  }

  backgrounds(): Promise<BackgroundLibrary> {
    return Promise.resolve({ items: [], max: 0, maxUploadBytes: 0 });
  }

  uploadBackground(): Promise<UploadedBackground> {
    return needsServer();
  }

  deleteBackground(): Promise<void> {
    return needsServer();
  }

  gallery(): Promise<ThemeGallery> {
    return Promise.resolve({ themes: [], applied: '', canRestore: false });
  }

  applyTheme(): Promise<Appearance> {
    return needsServer();
  }

  restoreStyle(): Promise<Appearance> {
    return needsServer();
  }

  importTheme(): Promise<ThemeInfo> {
    return needsServer();
  }

  deleteTheme(): Promise<void> {
    return needsServer();
  }

  // No server to download a theme file from; the gallery is empty anyway.
  exportUrl(): string {
    return '';
  }
}
