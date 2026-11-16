import type { PluginHttp } from "../../../../web/src/app/plugins/api";

/** A notice as GET /api/announcements returns it (ee/announcements, #297). */
export interface Announcement {
  id: string;
  title: string;
  /** Markdown. */
  body: string;
  pinned: boolean;
  authorName: string;
  expiresAt: string | null;
  createdAt: string;
  updatedAt: string;
  read: boolean;
  /** The viewer may change or delete it. */
  editable: boolean;
}

/** What the editor sends. */
export interface AnnouncementInput {
  title: string;
  body: string;
  pinned: boolean;
  /** "YYYY-MM-DD" or ''; the notice ends at the end of that day. */
  expiresOn: string;
  notify: boolean;
}

/** End of a local day as an ISO time, or null for ''. */
export function endOfDay(day: string): string | null {
  if (!day) return null;
  const [y, m, d] = day.split("-").map(Number);
  return new Date(y ?? 0, (m ?? 1) - 1, d ?? 1, 23, 59, 59).toISOString();
}

/** The local day of an ISO time ("YYYY-MM-DD"), for the date input. */
export function dayOf(iso: string | null): string {
  if (!iso) return "";
  const t = new Date(iso);
  const p = (n: number) => String(n).padStart(2, "0");
  return `${t.getFullYear()}-${p(t.getMonth() + 1)}-${p(t.getDate())}`;
}

/** Unread notices first within pinned and others; the server already sorts pinned first, newest first. */
export const unreadCount = (list: readonly Announcement[]) =>
  list.filter((a) => !a.read).length;

/** The announcements API through the core's client. */
export class AnnouncementsApi {
  constructor(private readonly http: PluginHttp) {}

  list(archive = false): Promise<Announcement[]> {
    return this.http.get<Announcement[]>(
      archive ? "/api/announcements?archive=1" : "/api/announcements",
    );
  }

  save(input: AnnouncementInput, id?: string): Promise<Announcement> {
    const body = {
      title: input.title.trim(),
      body: input.body,
      pinned: input.pinned,
      expiresAt: endOfDay(input.expiresOn),
      notify: input.notify,
    };
    return id
      ? this.http.put<Announcement>(
          `/api/announcements/${encodeURIComponent(id)}`,
          body,
        )
      : this.http.post<Announcement>("/api/announcements", body);
  }

  remove(id: string): Promise<void> {
    return this.http.delete(`/api/announcements/${encodeURIComponent(id)}`);
  }

  markRead(id: string): Promise<void> {
    return this.http.post(`/api/announcements/${encodeURIComponent(id)}/read`);
  }
}
