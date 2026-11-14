import type { MessageKey } from '../../../shared/i18n';
import {
  type Birthday,
  type BirthdayInput,
  type BirthdaySettings,
  type BirthdaySettingsInput,
  type ImportReport,
  toPayload,
} from '../domain/birthdays';

/** Output port: the birthdays API. */
export interface BirthdaysGateway {
  upcoming(limit: number): Promise<Birthday[]>;
  add(name: string, month: number, day: number, year: number): Promise<void>;
  remove(id: string): Promise<void>;
  importVCard(text: string): Promise<ImportReport>;
  settings(): Promise<BirthdaySettings>;
  saveSettings(input: BirthdaySettingsInput): Promise<BirthdaySettings>;
}

/** Thrown when the form is invalid before anything is sent; carries a message key. */
export class InvalidBirthday extends Error {
  constructor(readonly key: MessageKey) {
    super(key);
  }
}

/** Upcoming birthdays, manual entries, vCard import and the address book (#287). */
export class BirthdaysUseCases {
  constructor(private readonly gateway: BirthdaysGateway) {}

  upcoming(limit: number): Promise<Birthday[]> {
    return this.gateway.upcoming(limit);
  }

  async add(input: BirthdayInput): Promise<void> {
    const p = toPayload(input);
    if (typeof p === 'string') throw new InvalidBirthday(p);
    await this.gateway.add(p.name, p.month, p.day, p.year);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  async importVCard(file: { text(): Promise<string> }): Promise<ImportReport> {
    return this.gateway.importVCard(await file.text());
  }

  settings(): Promise<BirthdaySettings> {
    return this.gateway.settings();
  }

  saveSettings(input: BirthdaySettingsInput): Promise<BirthdaySettings> {
    return this.gateway.saveSettings({
      ...input,
      cardDavUrl: input.cardDavUrl.trim(),
      username: input.username.trim(),
    });
  }
}
