# Store listing texts

Texts for the Chrome Web Store and Microsoft Edge Add-ons dashboards, which
take them as form fields. The addons.mozilla.org listing is filled from
`amo-metadata.json` by the release workflow. The privacy statement is
`docs/extension-privacy.md`.

Limits: Chrome Web Store summary 132 characters, name 75 characters.

## English

**Name:** Waylight

**Summary:** Opens your self-hosted Waylight start page on every new tab and saves pages to it.

**Category:** Productivity

**Description**

Waylight is a self-hosted start page in the spirit of iGoogle: widgets on tabs for the weather, your calendar, to-dos and the state of your home server.

This extension shows your own Waylight on every new tab and adds "Save page to Waylight" to the toolbar and the context menu.

You can use it in two ways. With a Waylight server, enter its address in the extension's options; the extension talks only to that address. Without a server, the extension runs Waylight in the browser and keeps what you create in this browser profile.

It reads the URL and title of a tab only when you click "Save page", injects nothing into websites and sends nothing to the developers or third parties.

**Single purpose:** Show the user's Waylight dashboard as the new tab page and save the current page to it.

**Permission justifications**

- `storage`: keeps the address of the user's Waylight server and the chosen mode.
- `contextMenus`: adds "Save page to Waylight" to the context menu.
- `activeTab`: reads the URL and title of the current tab, only when the user clicks "Save page".
- Host access to the server: the extension requests the user's own server address at run time, after the user enters it in the options. It has no access to other sites.

**Data use disclosure:** collects web history (URL and title of a page the user saves), sent only to the user's own Waylight server. No analytics, no sale or transfer to third parties.

**Support:** https://github.com/Waylight-App/waylight/issues

**Homepage:** https://getwaylight.app

## Deutsch

**Name:** Waylight

**Kurzbeschreibung:** Öffnet deine selbst gehostete Waylight-Startseite in jedem neuen Tab und speichert Seiten dorthin.

**Kategorie:** Produktivität

**Beschreibung**

Waylight ist eine selbst gehostete Startseite im Stil von iGoogle: Widgets auf Tabs für Wetter, Kalender, To-dos und den Zustand deines Heimservers.

Diese Erweiterung zeigt dein eigenes Waylight in jedem neuen Tab und ergänzt „Seite in Waylight speichern“ in der Symbolleiste und im Kontextmenü.

Du kannst sie auf zwei Arten nutzen. Mit einem Waylight-Server trägst du dessen Adresse in den Einstellungen der Erweiterung ein; sie spricht nur mit dieser Adresse. Ohne Server führt die Erweiterung Waylight im Browser aus und speichert, was du anlegst, in diesem Browserprofil.

Sie liest URL und Titel eines Tabs nur, wenn du auf „Seite speichern“ klickst, verändert keine Websites und sendet nichts an die Entwickler oder Dritte.

**Support:** https://github.com/Waylight-App/waylight/issues

**Startseite:** https://getwaylight.app

## Screenshots

Both stores need at least one screenshot (1280×800 or 640×400). They are
taken by hand from the new tab page and the options page and uploaded in the
dashboards; they are not part of the repository.
