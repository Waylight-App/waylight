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
