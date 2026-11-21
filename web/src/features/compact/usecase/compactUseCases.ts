import type { CompactState } from '../domain/compact';

/** Output port: the compact states of the modules. */
export interface CompactGateway {
  state(module: string): Promise<CompactState>;
}

/** Application logic of compact widget views (#595). */
export class CompactUseCases {
  constructor(private readonly gateway: CompactGateway) {}

  /** The short state of a module's widget. */
  state(module: string): Promise<CompactState> {
    return this.gateway.state(module);
  }
}
