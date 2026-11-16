import type { PluginHttp } from "../../../../web/src/app/plugins/api";

/** Where someone works today (#346); "" is not set. */
export type Status = "office" | "home" | "away" | "";

/** Statuses a person can choose, in the order of the buttons. */
export const settable: readonly Exclude<Status, "">[] = [
  "office",
  "home",
  "away",
];

/** One person on the board. */
export interface Person {
  accountId: string;
  name: string;
  status: Status;
  /** "absence" when a planned absence (#299) decides the status. */
  source: "self" | "absence" | "none";
  returnOn?: string;
}

/** A group's presence today, as GET /api/presence returns it. */
export interface Board {
  day: string;
  people: Person[];
  counts: { office: number; home: number; away: number; none: number };
  /** null for guests and kiosk displays. */
  mine: Person | null;
}

/** The widget's settings: one group, or everyone without. */
export interface PresenceConfig {
  group: string;
}

/** Interval of the live refresh while the page is visible. */
export const REFRESH_MS = 30_000;

/** Reads the widget config; unknown values fall back to everyone. */
export function parseConfig(
  config: Record<string, unknown> | undefined,
): PresenceConfig {
  const group = config?.group;
  return { group: typeof group === "string" ? group : "" };
}

/**
 * The status to send when a button is pressed: pressing the active status
 * again clears it, unless an absence decides it (then nothing changes).
 */
export function nextStatus(
  mine: Person | null,
  pressed: Status,
): Status | null {
  if (!mine || mine.source === "absence") return null;
  return mine.status === pressed ? "" : pressed;
}

/** Calls fn now and every interval while the document is visible. */
export function poll(
  fn: () => void,
  interval = REFRESH_MS,
  doc: Pick<
    Document,
    "visibilityState" | "addEventListener" | "removeEventListener"
  > = document,
): () => void {
  const timer = setInterval(() => {
    if (doc.visibilityState === "visible") fn();
  }, interval);
  const onVisible = () => {
    if (doc.visibilityState === "visible") fn();
  };
  doc.addEventListener("visibilitychange", onVisible);
  fn();
  return () => {
    clearInterval(timer);
    doc.removeEventListener("visibilitychange", onVisible);
  };
}

/** The presence API. */
export class PresenceApi {
  constructor(private readonly http: PluginHttp) {}

  board(group: string): Promise<Board> {
    const q = group ? `?group=${encodeURIComponent(group)}` : "";
    return this.http.get<Board>(`/api/presence${q}`);
  }

  set(status: Status, returnOn = ""): Promise<void> {
    return this.http.put<void>("/api/presence/me", { status, returnOn });
  }

  groups(): Promise<{ id: string; name: string }[]> {
    return this.http.get("/api/auth/my-groups");
  }

  async enabled(): Promise<boolean> {
    return (await this.http.get<{ enabled: boolean }>("/api/presence/settings"))
      .enabled;
  }

  async setEnabled(enabled: boolean): Promise<boolean> {
    return (
      await this.http.put<{ enabled: boolean }>("/api/presence/settings", {
        enabled,
      })
    ).enabled;
  }
}
