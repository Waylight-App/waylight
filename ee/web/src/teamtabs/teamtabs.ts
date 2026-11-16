import type { PluginHttp } from "../../../../web/src/app/plugins/api";

/** An own tab and who else sees it (ee/teamtabs, #295). */
export interface TeamTab {
  id: string;
  name: string;
  /** "private", "everyone" or "group:<id>". */
  visibility: string;
}

export interface Group {
  id: string;
  name: string;
}

/** The team tab API and the groups to share with. */
export class TeamTabsApi {
  constructor(private readonly http: PluginHttp) {}

  tabs(): Promise<TeamTab[]> {
    return this.http.get<TeamTab[]>("/api/team-tabs");
  }

  async groups(): Promise<Group[]> {
    const list =
      await this.http.get<{ id: string; name: string }[]>("/api/auth/groups");
    return list.map((g) => ({ id: g.id, name: g.name }));
  }

  /** Group ID → default team tab ID. */
  defaults(): Promise<Record<string, string>> {
    return this.http.get<Record<string, string>>("/api/team-tabs/defaults");
  }

  /** Sets the default team tab of a group; "" removes it. */
  setDefault(groupId: string, tabId: string): Promise<void> {
    return this.http.put(
      `/api/team-tabs/defaults/${encodeURIComponent(groupId)}`,
      { tabId },
    );
  }

  share(id: string, visibility: string): Promise<{ visibility: string }> {
    return this.http.put(`/api/team-tabs/${encodeURIComponent(id)}`, {
      visibility,
    });
  }
}

/** Tabs a group sees: shared with it or with everyone. */
export const tabsFor = (tabs: readonly TeamTab[], groupId: string) =>
  tabs.filter(
    (t) => t.visibility === "everyone" || t.visibility === `group:${groupId}`,
  );
