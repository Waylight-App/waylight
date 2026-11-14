import { type ConfigResult, type ImportResult, parseExportFile } from '../domain/transfer';

/** Output port: transfer backend. */
export interface TransferGateway {
  exportUrl(): string;
  import(file: unknown): Promise<ImportResult>;
  configUrl(): string;
  applyConfig(yaml: string): Promise<ConfigResult>;
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

  /** URL that downloads trisse-home.yaml (#88). */
  configUrl(): string {
    return this.gateway.configUrl();
  }

  /** Applies a trisse-home.yaml; the server checks the whole file first. */
  applyConfigFile(yaml: string): Promise<ConfigResult> {
    return this.gateway.applyConfig(yaml);
  }
}
