import type { FeatureLevel, PluginHttp } from '../../../../web/src/app/plugins/api';

/** The licence as GET /api/licence returns it to admins (ee/licence). */
export interface Licence {
  status: 'missing' | 'invalid' | 'valid' | 'expired' | 'over-seats';
  source: '' | 'environment' | 'settings';
  problem?: string;
  licensee?: string;
  issued?: string;
  until?: string;
  seats: number;
  accounts: number;
  features: string[];
  unknown: string[];
}

/** Access to licence and feature levels through the core's API client. */
export class LicenceApi {
  constructor(private readonly http: PluginHttp) {}

  async levels(): Promise<Record<string, FeatureLevel>> {
    return (await this.http.get<{ features: Record<string, FeatureLevel> }>('/api/licence/features'))
      .features;
  }

  get(): Promise<Licence> {
    return this.http.get<Licence>('/api/licence');
  }

  install(key: string): Promise<Licence> {
    return this.http.put<Licence>('/api/licence', { key: key.trim() });
  }

  remove(): Promise<void> {
    return this.http.delete('/api/licence');
  }
}

/** Key of the status sentence for a licence. */
export function statusKey(l: Licence): string {
  return `status_${l.status.replace('-', '_')}`;
}
