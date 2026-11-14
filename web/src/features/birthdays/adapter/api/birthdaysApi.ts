import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  BirthdayDTO,
  ImportReportDTO,
  SettingsDTO,
} from '../../../../shared/api-client/generated/birthdays';
import type {
  Birthday,
  BirthdaySettings,
  BirthdaySettingsInput,
  BirthdaySource,
  ImportReport,
} from '../../domain/birthdays';
import type { BirthdaysGateway } from '../../usecase/birthdaysUseCases';

const toBirthday = (d: BirthdayDTO): Birthday => ({
  ...d,
  source: d.source as BirthdaySource,
  year: d.year ?? 0,
  age: d.age ?? 0,
});

/** BirthdaysGateway backed by the REST API. */
export class BirthdaysApi implements BirthdaysGateway {
  constructor(private readonly http: HttpClient) {}

  async upcoming(limit: number): Promise<Birthday[]> {
    return (await this.http.get<BirthdayDTO[]>(`/api/birthdays?limit=${limit}`)).map(toBirthday);
  }
  async add(name: string, month: number, day: number, year: number): Promise<void> {
    await this.http.post('/api/birthdays', { name, month, day, year });
  }
  async remove(id: string): Promise<void> {
    await this.http.delete(`/api/birthdays/${encodeURIComponent(id)}`);
  }
  async importVCard(text: string): Promise<ImportReport> {
    return { ...(await this.http.postText<ImportReportDTO>('/api/birthdays/import', text, 'text/vcard')) };
  }
  async settings(): Promise<BirthdaySettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/birthdays/settings')) };
  }
  async saveSettings(input: BirthdaySettingsInput): Promise<BirthdaySettings> {
    return { ...(await this.http.put<SettingsDTO>('/api/birthdays/settings', input)) };
  }
}
