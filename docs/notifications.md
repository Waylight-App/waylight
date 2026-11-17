# Notifications and job monitoring

Waylight can tell you when something needs attention: a service goes
offline, a disk fills up, a certificate is about to expire, a backup did not
run. This page explains how to set up the channels, how to keep alerts
quiet when they should be, and how your own backup scripts and cron jobs
report to Waylight.

## Channels

Open Settings → Notifications. The main channel is
[ntfy](https://ntfy.sh): enter a server (`https://ntfy.sh` or your own) and a
topic, and install the ntfy app on your phone. *Save & send test* checks the
setup straight away.

Topics on the public ntfy server can be read by anyone who knows the name.
Use the long random topic Waylight suggests, or run your own ntfy server
with an access token.

Under *More channels* you can add:

- **Gotify** with a server address and an app token;
- **webhooks**, which receive a JSON message. If you set a secret, each
  request carries an `X-Waylight-Signature: sha256=…` header, an HMAC of the
  body, so the receiver can check that it came from Waylight;
- **e-mail** through an SMTP server;
- **Slack**, through an incoming webhook;
- **Microsoft Teams**, through a Workflows webhook.

All channel secrets are stored encrypted with `SECRET_KEY`. For Slack and
Teams the webhook address itself is the secret: it is never shown again
after saving, and it is removed from error messages.

### Slack

1. In Slack, open the channel's settings → *Integrations* → *Add an app*
   and add **Incoming Webhooks**, or create your own Slack app with incoming
   webhooks at [api.slack.com/apps](https://api.slack.com/apps).
2. Copy the webhook URL (`https://hooks.slack.com/services/…`).
3. In Waylight choose **+ Slack**, paste the URL and save. *Test* posts
   a sample message.

Messages show the title, the text, the importance and the alert type. With
`PUBLIC_URL` set, they link back to the start page. Mentions such as
`<!channel>` in a service name are escaped, so an alert never pings a whole
channel.

### Microsoft Teams

The old Office 365 connectors ("Incoming Webhook" in a channel's
connectors) were switched off by Microsoft. Waylight uses the
replacement, a Workflows webhook:

1. In Teams, open the channel's **⋯** menu → **Workflows**.
2. Choose the template **Post to a channel when a webhook request is
   received**, give it a name, and pick the team and channel.
3. Copy the address the workflow shows at the end.
4. In Waylight choose **+ Microsoft Teams**, paste the address and save.
   *Test* posts a sample card.

Messages arrive as an Adaptive Card with the title in the colour of the
importance, the text and, with `PUBLIC_URL` set, an *Open Waylight*
button.

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

Waylight cannot see your backups, but your backup jobs can tell it how
they went. Under Settings → Backup monitoring, create a check for each job.
Waylight shows a secret ping URL once. Call it from the job:

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
Waylight stores only a hash of the token, so the URL cannot be shown
again. If it leaks, generate a new one.
