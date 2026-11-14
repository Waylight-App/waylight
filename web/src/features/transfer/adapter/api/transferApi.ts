import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  ConfigResultDTO,
  ConfigStatusDTO,
  ImportResultDTO,
} from '../../../../shared/api-client/generated/transfer';
import type { ConfigResult, ImportResult } from '../../domain/transfer';
import type { TransferGateway } from '../../usecase/transferUseCases';

/** TransferGateway backed by the REST API. */
export class TransferApi implements TransferGateway {
  constructor(private readonly http: HttpClient) {}

  exportUrl(): string {
    return '/api/transfer/export';
  }

  async import(file: unknown): Promise<ImportResult> {
    const d = await this.http.post<ImportResultDTO>('/api/transfer/import', file);
    return { imported: d.imported ?? [], skipped: d.skipped ?? [], failed: d.failed ?? {} };
  }

  configUrl(): string {
    return '/api/transfer/config';
  }

  async managedSections(): Promise<string[]> {
    return (await this.http.get<ConfigStatusDTO>('/api/transfer/config/status')).managed ?? [];
  }

  async applyConfig(yaml: string): Promise<ConfigResult> {
    const d = await this.http.postText<ConfigResultDTO>('/api/transfer/config', yaml, 'application/yaml');
    return { applied: d.applied ?? [], failed: d.failed ?? {} };
  }
}
