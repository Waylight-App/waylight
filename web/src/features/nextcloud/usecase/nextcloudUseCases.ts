import type { NextcloudSettings, NextcloudStatus } from '../domain/nextcloud';

/** Output port: Nextcloud backend. */
export interface NextcloudGateway {
  settings(): Promise<NextcloudSettings>;
  /** signIn is a finished sign-in of the setup assistant (#654) that replaces the credential. */
  saveSettings(
    url: string,
    username: string,
    appPassword: string,
    signIn?: string,
  ): Promise<NextcloudSettings>;
  status(connection?: string): Promise<NextcloudStatus>;
}

/** Application logic of the Nextcloud feature. */
export class NextcloudUseCases {
  constructor(private readonly gateway: NextcloudGateway) {}

  settings(): Promise<NextcloudSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty app password keeps the stored one. */
  saveSettings(url: string, username: string, appPassword: string): Promise<NextcloudSettings> {
    return this.gateway.saveSettings(url.trim(), username.trim(), appPassword.trim());
  }

  /** Saves the connection of a finished sign-in of the setup assistant (#654). */
  saveSignIn(signIn: string): Promise<NextcloudSettings> {
    return this.gateway.saveSettings('', '', '', signIn);
  }

  /** Reads through the named connection; without one, through the module settings. */
  status(connection?: string): Promise<NextcloudStatus> {
    return this.gateway.status(connection);
  }
}
