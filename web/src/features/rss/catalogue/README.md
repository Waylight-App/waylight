# Feed catalogue

`feeds.json` is the list of news sources offered under **Settings → Feeds → Browse catalogue** (#289). It ships with the app; opening the catalogue makes no outside request. Subscribing to an entry works like entering its address by hand: the server fetches the feed once and refuses addresses that do not answer with a feed.

## Format

A JSON array of entries:

```json
{
  "topic": "tech",
  "language": "de",
  "title": "heise online",
  "url": "https://www.heise.de/rss/heise-atom.xml"
}
```

- `topic`: one of `news`, `tech`, `selfhosting`, `science`, `sports`, `fun` (the UI groups by it; a new topic needs a label in `web/src/features/rss/i18n`).
- `language`: ISO 639-1 code of the feed's language (`de`, `en`, …).
- `title`: the source's own name.
- `url`: the address of the RSS, Atom or JSON feed, `https://` only.

A test checks the format and that no address appears twice.

## Adding a source

Open a pull request that adds one line. Please only add sources that are free to read, publish a full or partial feed officially, and are of general interest.
