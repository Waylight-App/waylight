import type {
  BackupFile,
  BackupPreview,
  BackupStatus,
  RemoteSettings,
  RemoteSettingsInput,
  RestoreRequest,
  RestoreResult,
} from '../domain/backup';

/** Output port: backup backend. */
export interface BackupGateway {
  list(): Promise<BackupFile[]>;
  create(): Promise<BackupFile>;
  remoteSettings(): Promise<RemoteSettings>;
  saveRemoteSettings(input: RemoteSettingsInput): Promise<RemoteSettings>;
  testRemote(input: RemoteSettingsInput): Promise<void>;
  disableRemote(): Promise<void>;
  listRemote(): Promise<BackupFile[]>;
  restore(req: RestoreRequest): Promise<RestoreResult>;
  preview(name: string): Promise<BackupPreview>;
  status(): Promise<BackupStatus>;
}

const newestFirst = (files: BackupFile[]) =>
  [...files].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());

/** Application logic of the backup feature. */
export class BackupUseCases {
  constructor(private readonly gateway: BackupGateway) {}

  /** Lists backups, newest first. */
  async list(): Promise<BackupFile[]> {
    return newestFirst(await this.gateway.list());
  }

  /** The traffic light of the backups (#677). */
  status(): Promise<BackupStatus> {
    return this.gateway.status();
  }

  backupNow(): Promise<BackupFile> {
    return this.gateway.create();
  }

  /** Off-site target (#46). */
  remoteSettings(): Promise<RemoteSettings> {
    return this.gateway.remoteSettings();
  }

  saveRemoteSettings(input: RemoteSettingsInput): Promise<RemoteSettings> {
    return this.gateway.saveRemoteSettings({
      ...input,
      url: input.url.trim(),
      username: input.username.trim(),
    });
  }

  testRemote(input: RemoteSettingsInput): Promise<void> {
    return this.gateway.testRemote({ ...input, url: input.url.trim(), username: input.username.trim() });
  }

  disableRemote(): Promise<void> {
    return this.gateway.disableRemote();
  }

  async listRemote(): Promise<BackupFile[]> {
    return newestFirst(await this.gateway.listRemote());
  }

  /** Checks a local backup before it is restored (#677). */
  preview(name: string): Promise<BackupPreview> {
    return this.gateway.preview(name);
  }

  /**
   * Restores a backup; the server saves the current state as a safety copy
   * first and restarts afterwards.
   */
  restore(req: RestoreRequest): Promise<RestoreResult> {
    return this.gateway.restore(req);
  }
}
