import {
  type ConfigResult,
  type ForeignFormat,
  type ForeignImport,
  type ImportResult,
  parseExportFile,
} from '../domain/transfer';

/** Output port: transfer backend. */
export interface TransferGateway {
  exportUrl(): string;
  import(file: unknown): Promise<ImportResult>;
  configUrl(): string;
  applyConfig(yaml: string): Promise<ConfigResult>;
  managedSections(): Promise<string[]>;
  importForeign(format: ForeignFormat, files: readonly File[], apply: boolean): Promise<ForeignImport>;
}

/** Application logic of the transfer feature. */
export class TransferUseCases {
  constructor(private readonly gateway: TransferGateway) {}

  /** URL that downloads the export file. */
  exportUrl(): string {
    return this.gateway.exportUrl();
  }

  /** Validates the file locally and imports it. */
  async importFile(text: string): Promise<ImportResult> {
    const parsed = parseExportFile(text);
    if (typeof parsed === 'string') throw new Error(parsed);
    return this.gateway.import(parsed);
  }

  /** URL that downloads waylight.yaml (#88). */
  configUrl(): string {
    return this.gateway.configUrl();
  }

  /** Shows what importing another dashboard would add, without changing anything (#77). */
  previewForeign(format: ForeignFormat, files: readonly File[]): Promise<ForeignImport> {
    return this.gateway.importForeign(format, files, false);
  }

  /** Adds the other dashboard's tiles, bookmarks and widgets. */
  applyForeign(format: ForeignFormat, files: readonly File[]): Promise<ForeignImport> {
    return this.gateway.importForeign(format, files, true);
  }

  /** Sections owned by CONFIG_FILE; they are read-only in the UI (#88). */
  managedSections(): Promise<string[]> {
    return this.gateway.managedSections();
  }

  /** Applies a waylight.yaml; the server checks the whole file first. */
  applyConfigFile(yaml: string): Promise<ConfigResult> {
    return this.gateway.applyConfig(yaml);
  }
}
