#!/bin/sh
# preinst for the waylight .deb.
#
# upgrade: stop the service and copy the database before the new version is
# unpacked, so that a failed upgrade can be rolled back with exactly the data
# of the old version (sudo apt install waylight=<old>, then copy the file
# back). The three newest copies are kept in backups/pre-upgrade. postinst
# starts the service again.
set -e

STATE_DIR=/var/lib/waylight
DB="$STATE_DIR/waylight.db"
DEST="$STATE_DIR/backups/pre-upgrade"
KEEP=3

if [ "$1" = upgrade ] && [ -f "$DB" ]; then
	if [ -d /run/systemd/system ]; then
		# A stopped service means a consistent copy, WAL included.
		deb-systemd-invoke stop waylight.service >/dev/null || true
	fi
	mkdir -p "$DEST"
	name="waylight-$2-$(date -u +%Y%m%dT%H%M%SZ).db"
	cp -p "$DB" "$DEST/$name"
	for suffix in -wal -shm; do
		if [ -f "$DB$suffix" ]; then
			cp -p "$DB$suffix" "$DEST/$name$suffix"
		fi
	done
	chown -R waylight:waylight "$STATE_DIR/backups" 2>/dev/null || true
	chmod 0750 "$DEST"
	ls -1t "$DEST"/waylight-*.db | tail -n +"$((KEEP + 1))" | while read -r old; do
		rm -f "$old" "$old-wal" "$old-shm"
	done
	echo "waylight: database copied to $DEST/$name"
fi

exit 0
