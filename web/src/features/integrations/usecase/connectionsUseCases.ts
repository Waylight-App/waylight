import { errorMessage } from '../../../shared/api-client/http';
import {
  type Connection,
  type ConnectionDraft,
  type ConnectionService,
  type ConnectionState,
  cleanCredential,
} from '../domain/integrations';

/** Changes to a saved connection; a missing credential keeps the stored one. */
export interface ConnectionChange {
  name: string;
  url: string;
  credential?: Record<string, string>;
}

/** Output port: the connections backend. */
export interface ConnectionsGateway {
  services(): Promise<ConnectionService[]>;
  list(): Promise<Connection[]>;
  add(d: ConnectionDraft): Promise<Connection>;
  update(id: string, change: ConnectionChange): Promise<Connection>;
  replaceCredential(id: string, credential: Record<string, string>): Promise<Connection>;
  check(id: string): Promise<void>;
  remove(id: string): Promise<void>;
}

/** Application logic of the named connections of the instance (#610). */
export class ConnectionsUseCases {
  constructor(private readonly gateway: ConnectionsGateway) {}

  services(): Promise<ConnectionService[]> {
    return this.gateway.services();
  }

  list(): Promise<Connection[]> {
    return this.gateway.list();
  }

  add(d: ConnectionDraft): Promise<Connection> {
    return this.gateway.add({
      service: d.service,
      name: d.name.trim(),
      url: d.url.trim(),
      credential: cleanCredential(d.credential),
    });
  }

  /** Renames or moves a connection; the stored credential stays. */
  edit(c: Connection, name: string, url: string): Promise<Connection> {
    return this.gateway.update(c.id, { name: name.trim(), url: url.trim() });
  }

  replaceCredential(id: string, credential: Record<string, string>): Promise<Connection> {
    return this.gateway.replaceCredential(id, cleanCredential(credential));
  }

  /** Tries the connection; a refusal becomes a state with the reason. */
  async check(id: string): Promise<ConnectionState> {
    try {
      await this.gateway.check(id);
      return { kind: 'ok' };
    } catch (err) {
      return { kind: 'failing', message: errorMessage(err) };
    }
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}
