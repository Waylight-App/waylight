import type { AttentionItem, ServiceConnection } from '../domain/attention';

/** Output port: the attention list of the server. */
export interface AttentionGateway {
  /** fresh asks the services again instead of the statuses the server keeps. */
  list(fresh: boolean): Promise<AttentionItem[]>;
  /** Every integration that reports a connection (#591). */
  connections(fresh: boolean): Promise<ServiceConnection[]>;
}

/** Application logic of the attention list (#592). */
export class AttentionUseCases {
  constructor(private readonly gateway: AttentionGateway) {}

  /** The current list, failures first, then what is not set up. */
  list(): Promise<AttentionItem[]> {
    return this.gateway.list(false);
  }

  /** The list after asking every service again ("Check again"). */
  check(): Promise<AttentionItem[]> {
    return this.gateway.list(true);
  }

  /** The connections of every integration; fresh asks the services again. */
  connections(fresh = false): Promise<ServiceConnection[]> {
    return this.gateway.connections(fresh);
  }
}
