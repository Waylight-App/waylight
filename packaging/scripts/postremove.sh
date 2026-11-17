#!/bin/sh
# postrm for the waylight .deb.
#
# remove: the unit is masked (dh_installsystemd behaviour), configuration and
#         data stay.
# purge:  the configuration in /etc/waylight is deleted. The data in
#         /var/lib/waylight is never deleted automatically; if it still
#         holds a database, the generated SECRET_KEY is moved next to it, so
#         a later reinstall can still decrypt stored credentials.
set -e

UNITS="waylight.service waylight-setup-info.timer"
STATE_DIR=/var/lib/waylight

case "$1" in
remove)
	if [ -x /usr/bin/deb-systemd-helper ]; then
		for unit in $UNITS; do
			deb-systemd-helper mask "$unit" >/dev/null || true
		done
	fi
	;;
purge)
	if [ -x /usr/bin/deb-systemd-helper ]; then
		for unit in $UNITS; do
			deb-systemd-helper purge "$unit" >/dev/null || true
			deb-systemd-helper unmask "$unit" >/dev/null || true
		done
	fi
	if [ -f /etc/waylight/secret.env ] && [ -f "$STATE_DIR/waylight.db" ]; then
		(umask 077 && cp /etc/waylight/secret.env "$STATE_DIR/secret.env.purged")
	fi
	rm -rf /etc/waylight
	rm -f /etc/issue.d/waylight.issue /boot/firmware/waylight-setup.txt
	if [ -d "$STATE_DIR" ]; then
		echo "waylight: data kept in $STATE_DIR (delete it yourself if you no longer need it)"
	fi
	;;
esac

if [ -d /run/systemd/system ]; then
	systemctl --system daemon-reload >/dev/null || true
fi

exit 0
