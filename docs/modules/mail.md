# Mail

**What it does:** Shows how many unread messages are waiting and lists the newest ones (sender, subject, time) of one or more IMAP inboxes. A click opens your webmail (#283).

## Configuration

1. Create an **app password** for mail clients at your provider (GMX, web.de, Gmail, Outlook, Posteo, mailbox.org …).
2. In **Settings → Mail** add the account: IMAP server (e.g. `imap.gmx.net`), user name, app password, optionally a folder (default `INBOX`) and the address of your webmail. The login is tested before saving.
3. Add the **Mail** widget and choose an account in its settings, or "all accounts" for one combined list.

## Privacy

- **Read-only:** the folder is opened with `EXAMINE` and headers are fetched with `BODY.PEEK`, so nothing is marked as read, moved or deleted.
- **Headers only:** sender, subject and date are read; message bodies are never downloaded or stored.
- The app password is stored encrypted with `SECRET_KEY` and never sent back to the browser.
- **Guests** who can see a tab with a mail widget see the senders and subjects in it, just as they see a pinned note. Put the widget on a tab of your own if that matters.

## Outbound requests

A TLS connection to each IMAP server (usually port 993) while a mail widget is shown, at most every two minutes per account. Only IMAP over TLS is supported.

## API

`GET /api/mail/inbox?account=<id>`, accounts under `/api/mail/settings/accounts` (not available to guests) – see `api/modules/mail.yaml`.
