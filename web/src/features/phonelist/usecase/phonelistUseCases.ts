import type {
  PhoneEntry,
  PhoneEntryInput,
  PhonelistSettings,
  PhonelistSettingsInput,
} from '../domain/phonelist';

/** Output port: phone list backend. */
export interface PhonelistGateway {
  list(): Promise<PhoneEntry[]>;
  add(input: PhoneEntryInput): Promise<PhoneEntry>;
  update(id: string, input: PhoneEntryInput): Promise<PhoneEntry>;
  remove(id: string): Promise<void>;
  settings(): Promise<PhonelistSettings>;
  saveSettings(input: PhonelistSettingsInput): Promise<PhonelistSettings>;
}

/** Application logic of the phone list (#300); the server validates the entries. */
export class PhonelistUseCases {
  constructor(private readonly gateway: PhonelistGateway) {}

  list(): Promise<PhoneEntry[]> {
    return this.gateway.list();
  }

  /** Adds or changes an entry (admins only, checked by the server). */
  save(input: PhoneEntryInput, id?: string): Promise<PhoneEntry> {
    const trimmed: PhoneEntryInput = {
      name: input.name.trim(),
      role: input.role.trim(),
      phone: input.phone.trim(),
      email: input.email.trim(),
      team: input.team.trim(),
    };
    return id ? this.gateway.update(id, trimmed) : this.gateway.add(trimmed);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  /** The address book (admins). */
  settings(): Promise<PhonelistSettings> {
    return this.gateway.settings();
  }

  saveSettings(input: PhonelistSettingsInput): Promise<PhonelistSettings> {
    return this.gateway.saveSettings({
      ...input,
      cardDavUrl: input.cardDavUrl.trim(),
      username: input.username.trim(),
    });
  }
}
