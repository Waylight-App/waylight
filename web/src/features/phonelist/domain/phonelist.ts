/** One person of the phone list (#300). */
export interface PhoneEntry {
  id: string;
  name: string;
  role: string;
  phone: string;
  /** Number for a tel: link (digits and a leading +); '' without a phone. */
  dial: string;
  email: string;
  /** Department or group shown next to the name. */
  team: string;
  /** From the address book: changed there, not here. */
  readOnly: boolean;
}

/** The address book the list reads people from (#300). */
export interface PhonelistSettings {
  cardDavUrl: string;
  username: string;
  hasPassword: boolean;
  lastSync: Date | null;
  lastError: string;
}

/** Input of the address book settings; an empty password keeps the stored one. */
export interface PhonelistSettingsInput {
  cardDavUrl: string;
  username: string;
  password: string;
}

/** Editable fields of an entry. */
export interface PhoneEntryInput {
  name: string;
  role: string;
  phone: string;
  email: string;
  team: string;
}

/** Entries whose name, role or team contain every word of the query. */
export function filterEntries(entries: readonly PhoneEntry[], query: string): PhoneEntry[] {
  const words = query.toLowerCase().split(/\s+/).filter(Boolean);
  if (words.length === 0) return [...entries];
  return entries.filter((e) => {
    const text = `${e.name} ${e.role} ${e.team}`.toLowerCase();
    return words.every((w) => text.includes(w));
  });
}

/** An input with all fields empty, for the add form. */
export const emptyInput = (): PhoneEntryInput => ({ name: '', role: '', phone: '', email: '', team: '' });

/** A team heading with its entries; team '' collects people without a team. */
export interface TeamGroup {
  team: string;
  entries: PhoneEntry[];
}

/**
 * Groups entries by team (#417), teams sorted by name and people without
 * a team last; the order inside a team is kept.
 */
export function groupByTeam(entries: readonly PhoneEntry[]): TeamGroup[] {
  const byTeam = new Map<string, PhoneEntry[]>();
  for (const e of entries) {
    const team = e.team.trim();
    const list = byTeam.get(team) ?? [];
    list.push(e);
    byTeam.set(team, list);
  }
  return [...byTeam.entries()]
    .map(([team, list]) => ({ team, entries: list }))
    .sort((a, b) => {
      if (a.team === '' || b.team === '') return a.team === '' ? 1 : -1;
      return a.team.localeCompare(b.team);
    });
}

/** Widget configuration of the phone list. */
export interface PhonelistConfig {
  groupByTeam: boolean;
}

/** Reads the widget configuration. */
export function parsePhonelistConfig(raw: Record<string, unknown>): PhonelistConfig {
  return { groupByTeam: raw['groupByTeam'] === true };
}

const csvField = (value: string): string =>
  /[",\r\n]/.test(value) ? `"${value.replace(/"/g, '""')}"` : value;

/**
 * The list as CSV for spreadsheets (#418): name, role, team, phone, e-mail,
 * CRLF line ends and a UTF-8 byte order mark so umlauts survive.
 */
export function toCsv(entries: readonly PhoneEntry[], header: readonly string[]): string {
  const rows = [header, ...entries.map((e) => [e.name, e.role, e.team, e.phone, e.email])];
  return '﻿' + rows.map((r) => r.map(csvField).join(',')).join('\r\n') + '\r\n';
}
