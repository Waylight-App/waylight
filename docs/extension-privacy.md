# Privacy of the browser extension

This page describes what the Waylight browser extension stores and sends. It is the privacy statement of the listing on addons.mozilla.org.

## What it stores

The address of your Waylight server, which you enter in the extension's options, and whether you use the extension with a server or without one. Both are kept in the browser's extension storage and are removed when you uninstall the extension.

Without a server, what you create (bookmarks, notes, to-dos, your dashboard and its appearance) is stored in the extension's own database in your browser profile. It never leaves the browser unless you download it yourself with *Download my data*, and it is removed when you uninstall the extension.

## What it sends, and to whom

Only to **your own Waylight server**, at the address you entered:

- When a new tab opens, one request checks whether the server is reachable, so the tab can say so if it is not.
- When you click **Save page to Waylight** (toolbar or context menu), the **URL and title of that tab** go to your server, which adds them to your read-later list or bookmarks.

Without a server, the weather widget asks **Open-Meteo** (open-meteo.com) for the forecast of the place you chose, with its coordinates rounded to about one kilometre, and sends the place names you search for when you set it up. Open-Meteo is the service a Waylight server uses as well; no other data goes there.

Apart from that, nothing is sent to the developers of Waylight or to any third party. There is no analytics, no telemetry and no advertising.

## Permissions

| Permission | Why |
|---|---|
| `storage` | Keeps the address of your server and the chosen mode. |
| `contextMenus` | Adds *Save page to Waylight* to the menu. |
| `activeTab` | Reads the URL and title of the current tab, only at the moment you click *Save page*. |

The extension has no access to other websites and injects no scripts into pages.

Firefox shows these data types when you install the add-on: *browsing activity* (the URL and title of a page you save). It is declared because the data leaves the browser, even though it only goes to your own server.

## Questions

Open an issue at [github.com/Waylight-App/waylight](https://github.com/Waylight-App/waylight/issues).
