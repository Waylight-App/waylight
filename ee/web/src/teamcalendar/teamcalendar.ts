import type { PluginHttp } from '../../../../web/src/app/plugins/api';

/** An own calendar subscription and who else sees its events (ee/teamcalendar, #299). */
export interface Subscription {
  id: string;
  name: string;
  /** "private", "everyone" or "group:<id>". */
  visibility: string;
}

export interface Group {
  id: string;
  name: string;
}

export class TeamCalendarApi {
  constructor(private readonly http: PluginHttp) {}

  subscriptions(): Promise<Subscription[]> {
    return this.http.get<Subscription[]>('/api/teamcalendar/subscriptions');
  }

  /** The groups the signed-in account may share with: its own, or all for admins. */
  async groups(): Promise<Group[]> {
    const list = await this.http.get<{ id: string; name: string }[]>('/api/auth/my-groups');
    return list.map((g) => ({ id: g.id, name: g.name }));
  }

  share(id: string, visibility: string): Promise<{ visibility: string }> {
    return this.http.put(`/api/teamcalendar/subscriptions/${encodeURIComponent(id)}`, { visibility });
  }
}
