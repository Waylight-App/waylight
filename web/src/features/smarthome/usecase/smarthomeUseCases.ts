import type { Entity, SmartHomeSettings } from '../domain/smarthome';

/** Output port: smart home backend. */
export interface SmartHomeGateway {
  settings(): Promise<SmartHomeSettings>;
  /** signIn is a finished sign-in of the setup assistant (#654) that replaces the token. */
  saveSettings(url: string, token: string, signIn?: string): Promise<SmartHomeSettings>;
  entities(connection?: string): Promise<Entity[]>;
  states(ids: string[], connection?: string): Promise<Entity[]>;
  act(id: string, connection?: string): Promise<void>;
}

/** Application logic of the smart home feature (#49). */
export class SmarthomeUseCases {
  constructor(private readonly gateway: SmartHomeGateway) {}

  settings(): Promise<SmartHomeSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty token keeps the stored one. */
  saveSettings(url: string, token: string): Promise<SmartHomeSettings> {
    return this.gateway.saveSettings(url.trim(), token.trim());
  }

  /** Saves the connection of a finished sign-in of the setup assistant (#654). */
  saveSignIn(signIn: string): Promise<SmartHomeSettings> {
    return this.gateway.saveSettings('', '', signIn);
  }

  /** Every entity, for the widget's entity picker. */
  /** Reads through the named connection; without one, through the module settings. */
  entities(connection?: string): Promise<Entity[]> {
    return this.gateway.entities(connection);
  }

  /** The widget's entities in their configured order; none means no request. */
  states(ids: string[], connection?: string): Promise<Entity[]> {
    return ids.length === 0 ? Promise.resolve([]) : this.gateway.states(ids, connection);
  }

  /** Toggles or presses an entity; the UI asks for confirmation first. */
  act(id: string, connection?: string): Promise<void> {
    return this.gateway.act(id, connection);
  }
}
