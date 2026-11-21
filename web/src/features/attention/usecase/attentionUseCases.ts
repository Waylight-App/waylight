import type { AttentionItem } from '../domain/attention';

/** Output port: the server's attention list (#599). */
export interface AttentionGateway {
  list(): Promise<AttentionItem[]>;
}

/** Application logic of the attention line (#593). */
export class AttentionUseCases {
  constructor(private readonly gateway: AttentionGateway) {}

  /** The current attention items of the account. */
  list(): Promise<AttentionItem[]> {
    return this.gateway.list();
  }
}
