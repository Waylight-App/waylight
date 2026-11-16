import type { PluginHttp } from "../../../../web/src/app/plugins/api";

/** One service in GET /api/connections (ee/connections, #326). */
export interface Connection {
  service: string;
  name: string;
  /** False until an admin registers the instance at the service. */
  configured: boolean;
  /** '', 'connected' or 'expired'. */
  status: string;
  since?: string;
}

/** The registration of the instance at a service (admins). */
export interface Client {
  service: string;
  name: string;
  clientId: string;
  tenant: string;
  hasSecret: boolean;
}

export interface Clients {
  /** The address to enter at the service as redirect URL. */
  redirectUrl: string;
  clients: Client[];
}

/** The outcome the server appends after the round trip to the service. */
export function outcome(search: string): {
  connected?: string;
  failed?: string;
} {
  const q = new URLSearchParams(search);
  return {
    connected: q.get("connected") ?? undefined,
    failed: q.get("connection_error") ?? undefined,
  };
}

export class ConnectionsApi {
  constructor(private readonly http: PluginHttp) {}

  list(): Promise<Connection[]> {
    return this.http.get<Connection[]>("/api/connections");
  }

  /** Starts the sign-in at the service; returns the address to open. */
  async start(service: string): Promise<string> {
    return (
      await this.http.post<{ url: string }>(
        `/api/connections/${encodeURIComponent(service)}`,
      )
    ).url;
  }

  disconnect(service: string): Promise<void> {
    return this.http.delete(`/api/connections/${encodeURIComponent(service)}`);
  }

  clients(): Promise<Clients> {
    return this.http.get<Clients>("/api/connections/clients");
  }

  saveClient(
    service: string,
    input: { clientId: string; clientSecret: string; tenant: string },
  ): Promise<unknown> {
    return this.http.put(
      `/api/connections/clients/${encodeURIComponent(service)}`,
      {
        clientId: input.clientId.trim(),
        clientSecret: input.clientSecret,
        tenant: input.tenant.trim(),
      },
    );
  }
}
