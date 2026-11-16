# Phones, offline use and the browser extension

trisse-home is a web page, but it can behave like an app: installed on your
phone's home screen, readable without a connection and opened on every new
browser tab. This page explains what each of these needs.

## Install it as an app

On Android (Chrome), iOS (Safari, *Share → Add to Home Screen*) and desktop
browsers with app support you can install trisse-home. It then opens in its
own window without the browser's address bar, and on Android it appears in
the share menu for [quick capture](search.md#saving-from-other-pages).

Browsers only allow installation from a **secure context**: an `https://`
address or `localhost`. On a plain `http://192.168.x.x:8090` address the
dashboard works normally, but the browser will not offer to install it and
there is no offline mode. Put trisse-home behind a
[reverse proxy](reverse-proxy.md) with a certificate to get both.

## Offline use

Once installed, or once the page has been opened over HTTPS, trisse-home
keeps the app itself and the last loaded data of most widgets on the device.
When the server cannot be reached, the dashboard still opens and shows that
data, and each widget's status badge tells you how old it is.

To-dos are the exception that also works the other way: you can add, tick off
and edit to-dos while offline. The changes are kept in the browser and sent
to the server when the connection is back. If a to-do was changed elsewhere
in the meantime, trisse-home does not overwrite it silently; it lists the
conflicting changes under *Resolve* and lets you decide.

What is never stored on the device: the login, settings, exports, backups,
logs and every request that changes data, apart from the to-do queue above.
Signing out deletes the stored data.

## Browser extension

The extension makes trisse-home the page you see on every new tab, with the
cursor already in the search field. It also adds *Save page to trisse-home*
to the toolbar and to the context menu, which opens the quick capture dialog
with the current page, the link you right-clicked or the selected text.

It works in Chrome, Edge and Firefox:

1. Download `trisse-home-extension-<version>.zip` from the
   [latest release](https://github.com/itadori26/trisse-home/releases/latest)
   and unzip it.
2. Load it:
   - **Chrome and Edge:** open `chrome://extensions`, switch on
     *Developer mode*, choose *Load unpacked* and select the unzipped folder.
   - **Firefox:** once the add-on is listed on addons.mozilla.org (#290), install
     it from there like any other add-on; it then updates itself. Until then,
     open `about:debugging#/runtime/this-firefox`, choose *Load Temporary
     Add-on* and select `manifest.json`; Firefox removes temporary add-ons on
     restart.
3. Open a new tab and choose *Use without a server* or *Connect my server*.
   For a server, enter the address of your dashboard in the options, for
   example `http://homeserver:8090`.

### Without a server

Without a server the extension runs trisse-home itself, in the browser:
bookmarks, notes, to-dos, countdowns, the focus timer, habits, the clock and
the weather. Everything that needs a machine running in the background stays
out, such as service checks, Docker, parcels and notifications. Your data is
stored in this browser profile only; it is not synced to other devices and
is lost when you remove the extension.

To move to a server later, open the extension's options, choose *Download my
data*, and import the file on the server under *Settings → Transfer*. The
import replaces the server's dashboard and appearance with what you made in
the browser. Then switch the options to *With my trisse-home server*.

The extension asks for three permissions: `storage` for the address and the mode,
`contextMenus` for the menu entry and `activeTab` to read the URL and title
of the tab you save, only at the moment you click. It has no access to other
sites, injects no scripts into pages and sends nothing to anyone but your own
dashboard ([privacy of the extension](extension-privacy.md)). Its only
request of its own checks whether your dashboard is reachable; if it is
not, the new tab says so.

Chrome keeps the cursor in the address bar on a new tab. Click into the page
or press `/` to search there.
