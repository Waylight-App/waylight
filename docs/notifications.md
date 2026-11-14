# Notifications and job monitoring

trisse-home can tell you when something needs attention: a service goes
offline, a disk fills up, a certificate is about to expire, a backup did not
run. This page explains how to set up the channels, how to keep alerts
quiet when they should be, and how your own backup scripts and cron jobs
report to trisse-home.

## Channels

Open Settings → Notifications. The main channel is
[ntfy](https://ntfy.sh): enter a server (`https://ntfy.sh` or your own) and a
topic, and install the ntfy app on your phone. *Save & send test* checks the
setup straight away.

Topics on the public ntfy server can be read by anyone who knows the name.
Use the long random topic trisse-home suggests, or run your own ntfy server
with an access token.

Under *More channels* you can add:

- **Gotify** with a server address and an app token;
- **webhooks**, which receive a JSON message. If you set a secret, each
  request carries an `X-Trisse-Signature: sha256=…` header, an HMAC of the
  body, so the receiver can check that it came from trisse-home;
- **e-mail** through an SMTP server.

All channel secrets are stored encrypted with `SECRET_KEY`.

## What you are notified about

Under *Notify me when …* you choose the events:

- a service goes offline or comes back;
- a container stops or becomes unhealthy;
- a disk is fuller than a percentage you set;
- a TLS certificate expires within a number of days;
- a monitored backup or job fails or is overdue;
- someone signs in from a new device, or sign-ins are locked after too many
  failures;
- and, depending on the modules you use, disk health changes, new image
  versions, due to-dos, a slow internet connection or the daily briefing.

A change of state is only reported after two checks in a row agree, so a
single slow answer does not wake you up. After a restart, the first check
only records the current state.

## Importance, quiet hours and maintenance

Every type of message has an importance: info, warning or critical. Each
channel can be limited to a minimum importance and to selected types, so you
could send everything to ntfy but only critical messages by e-mail.

**Quiet hours** hold back everything except critical messages during the
night and send one summary when they end.

**Maintenance windows** mute selected services and containers for a while,
for example "NAS, 60 minutes" while you update it. The dashboard shows a
banner during the window, and the downtime does not count against the uptime
statistics.

## Backup and job monitoring

trisse-home cannot see your backups, but your backup jobs can tell it how
they went. Under Settings → Backup monitoring, create a check for each job.
trisse-home shows a secret ping URL once. Call it from the job:

```sh
curl -fsS -X POST "$PING_URL/start"     # optional: marks the start and measures the duration
restic backup /data \
  && curl -fsS -X POST "$PING_URL" \
  || curl -fsS -X POST --data "restic failed" "$PING_URL/fail"
```

This works for any job, not only backups: certificate renewals, database
dumps, a nightly `rsync`.

Each check has a schedule, either "at least every N hours" or a cron
expression such as `0 3 * * *` or `@daily`, and a grace period. A check is
*overdue* when the expected run plus the grace period passes without a
successful ping. Failures and overdue checks appear on the dashboard, on the
timeline and as notifications.

The ping URL needs no sign-in, but it is still subject to `ALLOWED_CIDRS`.
trisse-home stores only a hash of the token, so the URL cannot be shown
again. If it leaks, generate a new one.
