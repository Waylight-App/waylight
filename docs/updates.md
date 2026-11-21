# Updates and rollback

New releases are listed with their changes on the
[releases page](https://github.com/Waylight-App/waylight/releases). Settings →
About shows the version you run and, if you turn on the update check, whether
a newer one exists. The check asks the GitHub API at most once a day.

## Docker Compose

With the default `WAYLIGHT_VERSION=latest`, an update is:

```sh
cd /opt/waylight
git pull
sudo docker compose pull
sudo docker compose up -d
```

`git pull` keeps `compose.yaml` and the helper scripts current; your `.env`
is not tracked and stays as it is.

**Image name.** Releases are published as `ghcr.io/waylight-app/waylight`.
The previous name `ghcr.io/itadori26/waylight` stays as an alias: every
release is pushed there too, with the same tags and the same digest, so an
installation that still pulls the old name keeps updating. After `git pull`,
`compose.yaml` uses the new name.

**Pin a version** by setting `WAYLIGHT_VERSION=1.2.3` in `.env` and
running `sudo docker compose up -d`. You then update only when you change
the version.

**Roll back** the same way: set the previous version and run
`sudo docker compose up -d`. Database migrations only ever add tables and
columns, so an older version runs on data written by a newer one. If you want
the exact data from before the update, [restore](backups.md) the backup from
that morning.

**Build from source**, for development or an unreleased commit:

```sh
sudo docker compose -f compose.yaml -f compose.build.yaml up -d --build
```

## Debian package and Raspberry Pi

The Raspberry Pi image updates itself every night, including Raspberry Pi
OS. The Debian package updates with `sudo apt upgrade` once the Waylight
APT repository is set up, and before every update the database is copied to
`/var/lib/waylight/backups/pre-upgrade/`. Pinning (`apt-mark hold`) and
rollback (`apt install --allow-downgrades`) are described in the
[native install guide](native-install.md#update-and-rollback) and the
[Raspberry Pi guide](raspberry-pi.md#updates).

## Checking a release

The container image carries an SBOM and a build provenance record, which you
can inspect with:

```sh
docker buildx imagetools inspect ghcr.io/waylight-app/waylight:1.2.3 \
  --format '{{ json .Provenance }}'
```

The Debian packages come with a SHA-256 checksum file (see the
[native install guide](native-install.md#install)), and the Raspberry Pi
image with a signed attestation you can check with `gh attestation verify`
(see the [Raspberry Pi guide](raspberry-pi.md#1-flash-the-sd-card)).
