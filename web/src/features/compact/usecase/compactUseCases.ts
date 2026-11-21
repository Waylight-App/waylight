import type { CompactAction, CompactChoice, CompactState } from '../domain/compact';

/** Output port: the compact states of the modules and their actions. */
export interface CompactGateway {
  /** The state of a compact source (see compactSource). */
  state(source: string): Promise<CompactState>;
  /** Sends the request an action describes (#598). */
  send(method: string, path: string, body?: Record<string, unknown>): Promise<void>;
}

/** Methods an action may use; the server only describes state changes. */
const ACTION_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

/** Application logic of compact widget views (#595, #598). */
export class CompactUseCases {
  constructor(private readonly gateway: CompactGateway) {}

  /** The short state of a widget, from its compact source (see compactSource). */
  state(source: string): Promise<CompactState> {
    return this.gateway.state(source);
  }

  /**
   * Runs a compact action, with the body of the chosen chip when the action
   * has choices. Only same-origin API paths are accepted, so a state can
   * never make the browser send a request elsewhere.
   */
  async act(action: CompactAction, choice?: CompactChoice): Promise<void> {
    const method = action.method.toUpperCase();
    if (!ACTION_METHODS.has(method) || !action.path.startsWith('/api/')) {
      throw new Error(`unsupported compact action ${action.method} ${action.path}`);
    }
    await this.gateway.send(method, action.path, choice?.body ?? action.body);
  }
}
