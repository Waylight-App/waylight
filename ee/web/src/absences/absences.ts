import type { PluginHttp } from '../../../../web/src/app/plugins/api';

/** Kinds of absence; "away" is someone else's private absence (#299). */
export type AbsenceKind = 'holiday' | 'sick' | 'remote' | 'other' | 'away';

/** Kinds a person can choose when entering an absence. */
export const enterableKinds = ['holiday', 'sick', 'remote', 'other'] as const;

/** An absence as GET /api/absences returns it (ee/absences). */
export interface Absence {
  id: string;
  accountId: string;
  name: string;
  /** First and last day, both included, "YYYY-MM-DD". */
  from: string;
  to: string;
  kind: AbsenceKind;
  note: string;
  private: boolean;
  /** The viewer may change or delete it. */
  editable: boolean;
}

/** What the form sends; accountId only when an admin enters for someone else. */
export interface AbsenceInput {
  accountId: string;
  from: string;
  to: string;
  kind: (typeof enterableKinds)[number];
  note: string;
  private: boolean;
}

/** A day as "YYYY-MM-DD" in local time. */
export function dayKey(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

/** The day `n` days after `day` (local calendar). */
export function addDays(day: string, n: number): string {
  const [y, m, d] = day.split('-').map(Number);
  return dayKey(new Date(y ?? 0, (m ?? 1) - 1, (d ?? 1) + n));
}

/** Whether an absence covers a day (string compare works for ISO days). */
export const covers = (a: Pick<Absence, 'from' | 'to'>, day: string) => a.from <= day && day <= a.to;

/** The widget's two sections: away today, and away later this week. */
export interface AbsenceSections {
  today: Absence[];
  week: Absence[];
}

/**
 * Splits absences into "today" and "later in the next days" (`days` days
 * including today), sorted by start and name; ones that ended are dropped.
 */
export function sections(list: readonly Absence[], today: string, days = 7): AbsenceSections {
  const last = addDays(today, days - 1);
  const byStart = (a: Absence, b: Absence) => a.from.localeCompare(b.from) || a.name.localeCompare(b.name);
  const relevant = list.filter((a) => a.to >= today && a.from <= last);
  return {
    today: relevant.filter((a) => covers(a, today)).sort(byStart),
    week: relevant.filter((a) => !covers(a, today)).sort(byStart),
  };
}

/** Problem with the form, or null: the period must be valid and at most a year. */
export function validate(
  input: Pick<AbsenceInput, 'from' | 'to' | 'note'>,
): 'dates' | 'too_long' | 'note' | null {
  if (
    !/^\d{4}-\d{2}-\d{2}$/.test(input.from) ||
    !/^\d{4}-\d{2}-\d{2}$/.test(input.to) ||
    input.to < input.from
  )
    return 'dates';
  if (input.to > addDays(input.from, 365)) return 'too_long';
  if ([...input.note].length > 200) return 'note';
  return null;
}

/** The absences API through the core's client. */
export class AbsencesApi {
  constructor(private readonly http: PluginHttp) {}

  list(from: string, to: string): Promise<Absence[]> {
    const q = new URLSearchParams({ from, to });
    return this.http.get<Absence[]>(`/api/absences?${q}`);
  }

  save(input: AbsenceInput, id?: string): Promise<Absence> {
    const body = {
      from: input.from,
      to: input.to,
      kind: input.kind,
      note: input.note.trim(),
      private: input.private,
    };
    if (id) return this.http.put<Absence>(`/api/absences/${encodeURIComponent(id)}`, body);
    return this.http.post<Absence>(
      '/api/absences',
      input.accountId ? { ...body, accountId: input.accountId } : body,
    );
  }

  remove(id: string): Promise<void> {
    return this.http.delete(`/api/absences/${encodeURIComponent(id)}`);
  }

  /** Members an admin can enter absences for; empty for everyone else. */
  async members(): Promise<{ id: string; name: string }[]> {
    const r = await this.http.get<{ members: { id: string; name: string; disabled: boolean }[] }>(
      '/api/auth/members',
    );
    return r.members.filter((m) => !m.disabled).map((m) => ({ id: m.id, name: m.name }));
  }
}
