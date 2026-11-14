import { describe, expect, it, vi } from 'vitest';
import { type BirthdaysGateway, BirthdaysUseCases, InvalidBirthday } from './birthdaysUseCases';

function gateway(): BirthdaysGateway {
  return {
    upcoming: vi.fn(async () => []),
    add: vi.fn(async () => {}),
    remove: vi.fn(async () => {}),
    importVCard: vi.fn(async () => ({ added: 1, duplicates: 0 })),
    settings: vi.fn(),
    saveSettings: vi.fn(async (s) => ({ ...s, hasPassword: true })),
  };
}

describe('BirthdaysUseCases', () => {
  it('validates before sending', async () => {
    const g = gateway();
    const uc = new BirthdaysUseCases(g);
    await expect(uc.add({ name: ' ', date: '2000-01-01', yearKnown: true })).rejects.toBeInstanceOf(
      InvalidBirthday,
    );
    expect(g.add).not.toHaveBeenCalled();
    await uc.add({ name: 'Anna', date: '1990-11-20', yearKnown: false });
    expect(g.add).toHaveBeenCalledWith('Anna', 11, 20, 0);
  });

  it('reads the file and trims the address book fields', async () => {
    const g = gateway();
    const uc = new BirthdaysUseCases(g);
    await uc.importVCard({ text: async () => 'BEGIN:VCARD' });
    expect(g.importVCard).toHaveBeenCalledWith('BEGIN:VCARD');
    await uc.saveSettings({
      remindEnabled: false,
      remindAt: '08:00',
      cardDavUrl: ' https://x/ ',
      username: ' u ',
      password: '',
    });
    expect(g.saveSettings).toHaveBeenCalledWith(
      expect.objectContaining({ cardDavUrl: 'https://x/', username: 'u' }),
    );
  });
});
