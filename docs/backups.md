# Backups and restore

Everything Waylight knows lives in one SQLite database. This page explains
what is backed up automatically, how to keep a copy away from the server and
how to restore it. The examples use Docker Compose; the
[Raspberry Pi](raspberry-pi.md#backups) and
[native install](native-install.md) guides cover the same steps for their
paths.

## Automatic backups

Waylight makes a consistent copy of the database once a day and keeps the
last seven. Both can be changed with `BACKUP_INTERVAL` and
`BACKUP_RETENTION` ([configuration](configuration.md#storage-and-backups)).
In Docker the backups are stored in the data volume under `/data/backups`.
Settings → Backups lists them and lets you download each one.

Backups contain the passwords and tokens of your integrations only in
encrypted form. **Keep your `SECRET_KEY` separately**, for example in your
password manager. A backup restored without the matching key still works,
but you will have to enter every integration password again.

## Is my backup all right?

The *Data & backup* tile on the settings start page and the top of
Settings → Backups show a traffic light:

- **Green**: the newest backup is as recent as the schedule promises, for
  example "Last backup today 03:00, all good".
- **Amber**: there is no backup yet, the newest one is older than expected
  (one interval plus half an interval, at least two hours of slack), or
  the backup worked but the upload to the off-site target failed.
- **Red**: the latest attempt failed. Settings → Backups shows the reason;
  the light turns green again with the next backup that works.

The *Waylight backup* widget puts the same light on a dashboard, and its
compact view fits a small tile. Only the owner can read it. To watch backups
made by other tools, such as restic or borg on your NAS, use the
[backup monitor](notifications.md#backup-and-job-monitoring) instead.

## Copies away from the server

A backup on the same disk does not help when that disk dies. Under Settings
→ Backups you can add an **off-site target**: any WebDAV folder, such as a
Nextcloud folder with an app password. Each new backup is then uploaded there
too, and the copies are pruned with their own retention. A failed upload
raises a [notification](notifications.md).

To copy the backups by hand instead:

```sh
sudo docker run --rm -v waylight-data:/data -v "$PWD":/backup alpine \
  cp -r /data/backups /backup/
```

## Restore in the web interface

Only the owner can restore a backup, and only from their own browser
session: admins, members, guests and API tokens cannot. Settings → Backups →
*Restore a backup* walks you through three steps.

1. **Choose a backup.** Every local and off-site backup is listed with its
   date and size. You can also upload a `.db` file you downloaded earlier.
2. **Check it.** For a local backup Waylight shows right away whether the
   file is intact and fits the running version, and how many tabs, widgets
   and modules with data it holds. An older backup is fine: the missing
   database updates run on the next start. Off-site backups and uploads are
   checked the same way when you confirm, before anything changes.
3. **Confirm.** Tick that you understand the current data will be replaced,
   for every account, then enter your password and your two-factor code if
   it is on.

Waylight then saves the current state as a new backup, the **safety copy**,
and tells you its date. If you change your mind after the restore, restore
that safety copy the same way. If the safety copy cannot be written, for
example because the disk is full, nothing is restored.

Finally Waylight restarts; Docker and systemd start it again with the
restored data. The database file from before is also kept next to the new
one as `waylight.db.before-restore` until the next restore. Changes made
between the safety copy and the restart, a matter of seconds, are only in
that file.

## Restore without the web interface

If Waylight does not start any more:

1. Stop it: `sudo docker compose stop waylight`.
2. In the data volume, replace `waylight.db` with the backup file and
   delete `waylight.db-wal` and `waylight.db-shm` if they exist.
3. Start it again: `sudo docker compose start waylight`.

## Export and import

Settings → Import & export downloads your data as one JSON file: tabs and
widgets, notes, to-dos, bookmarks, feeds and the other content of each
module. Passwords, tokens and other secrets are never part of the export.

The export is not a replacement for backups. It is meant for moving data
between installations, for keeping a readable copy and for starting over.
Importing it into a fresh installation recreates your dashboard; you then
enter the integration passwords again.

The same page imports bookmarks exported from a browser as an HTML file.

For settings you want to keep in Git or copy to another instance, the same
page also writes and reads `waylight.yaml`: see
[Settings as a YAML file](config-as-code.md).
