import type {
  CallFilterKind,
  CallList,
  CallsOptions,
  FritzBoxSettings,
  FritzBoxStatus,
} from '../domain/fritzbox';

/** Output port: FRITZ!Box backend. */
export interface FritzBoxGateway {
  settings(): Promise<FritzBoxSettings>;
  saveSettings(url: string, username: string, password: string): Promise<FritzBoxSettings>;
  status(connection?: string): Promise<FritzBoxStatus>;
  calls(filter: CallFilterKind, limit: number, connection?: string): Promise<CallList>;
}

/** Application logic of the FRITZ!Box feature. */
export class FritzBoxUseCases {
  constructor(private readonly gateway: FritzBoxGateway) {}

  settings(): Promise<FritzBoxSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; user and password are optional. */
  saveSettings(url: string, username: string, password: string): Promise<FritzBoxSettings> {
    return this.gateway.saveSettings(url.trim(), username.trim(), password);
  }

  /** Reads through the named connection; without one, through the module settings. */
  status(connection?: string): Promise<FritzBoxStatus> {
    return this.gateway.status(connection);
  }

  /** The call list of a calls widget; fetched by the server only when asked (#168). */
  calls(options: CallsOptions, connection?: string): Promise<CallList> {
    return this.gateway.calls(options.filter, options.limit, connection);
  }
}
