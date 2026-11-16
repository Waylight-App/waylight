# Licence keys

The features in `ee/` (team tabs, single sign-on with groups, branding and the
others listed in [ADR 0020](adr/0020-ee-stays-proprietary.md)) work only with a
licence key. Everything else in trisse-home needs no key. This page explains
how a key is checked and how to enter one.

Builds published so far do not contain `ee/`. A server built with
`go build -tags ee ./cmd/server` does.

## What a key contains

A key is a line of text starting with `THL1.`. It names the licensee, the
last day it is valid, the number of accounts it covers and the features it
unlocks. trisse-home checks its signature on the server itself; no request
leaves the instance to validate it.

Every account counts: the owner, members and guests that are not disabled.
Kiosk displays do not count.

## Enter a key

As an admin, either

- open *Settings → Licence*, paste the key and save, or
- set the environment variable `LICENCE_KEY` to the key and restart. A key
  set this way wins over one entered in the settings and cannot be changed
  there.

## When the key does not cover everything

| Situation | What happens |
| --- | --- |
| No key, or a key that is not valid | The licensed features are hidden for every account; their API answers 403. |
| The key has expired | Data of the licensed features stays readable; changes are refused until a valid key is entered. |
| More accounts than the key covers | The same as an expired key, until accounts are removed or a key with more seats is entered. |
| The key names a feature this version does not know | The rest of the key works; the unknown feature is listed in the settings. |

Removing the key hides the licensed features again. Their data is kept.
