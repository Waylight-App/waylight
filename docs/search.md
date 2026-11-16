# Search and quick capture

The search field at the top does more than send you to a search engine. It
finds your own bookmarks, notes and to-dos, answers questions about your
server, calculates, and saves new to-dos and notes without opening a
widget. This page lists what it understands.

## Opening the search

Click the search field, or press `Ctrl+K` or `/` anywhere on the dashboard.
While you type, matching entries from your bookmarks, service tiles, notes
and to-dos appear below the field. `Enter` opens the selected entry. If
nothing matches, `Enter` searches the web.

The search learns which results you pick and ranks them higher next time.
This history stays in your browser. You can switch it off or clear it under
Settings → Search.

## Web search

DuckDuckGo is the default engine, so all of its
[!bangs](https://duckduckgo.com/bangs) work: `!w Berlin` searches Wikipedia,
`!gh trisse-home` GitHub. Under Settings → Search you can pick a different
engine, add your own shortcuts with a keyword and a URL containing `{query}`, and
have results open in a new tab. The search is also registered as an
OpenSearch provider, so most browsers offer to add it as a search engine.

## Calculator and units

Simple arithmetic and unit conversions are answered while you type, without
a request to the server. Numbers follow the format of your language, and
`Enter` copies the result.

| Input | Result |
| --- | --- |
| `2+3*4` | 14 |
| `19 % of 240` | 45.6 |
| `240 + 19 %` | 285.6 |
| `5 km in mi` | about 3.11 mi |
| `30 °C in F` | 86 °F |
| `2 GiB in MB` | about 2147.48 MB |

## Time and dates

Questions about the time and the calendar are answered in the browser too,
in German or English, and `Enter` copies the answer. Times and dates follow
the format of your language.

| Input | Answer |
| --- | --- |
| `uhrzeit`, `wie spät`, `time` | The current time |
| `datum`, `heute`, `date` | Today's date with the weekday |
| `kw`, `week` | The ISO calendar week |
| `zeit in tokio`, `time in new york` | The time there and how far it is from yours |
| `tage bis weihnachten`, `days until 2027-03-01` | The number of days and the weekday of that day |
| `in 90 tagen`, `vor 2 wochen`, `90 days ago` | The resulting date |

For "time in …", the search knows large cities and many countries by their
German and English names; other places go to the normal search. Named days are Christmas Eve, Christmas, New Year's Eve, New Year's
Day and Easter. A date without a year, such as `1.3.`, means the next such
day. With slashes, German reads day first (`1/3` is 1 March) and English
month first (`3/1`).

## Answers about your home

The quick search answers short questions in English or German, and it
forgives typos. You can ask in full ("how full is the disk?", "wetter
morgen") or use a prefix with a slash for a direct answer:

| Prefix | Answer |
| --- | --- |
| `/cpu`, `/ram`, `/load`, `/disk`, `/uptime` | Current values of the host |
| `/docker <name>` | Status of a container, an update hint, and a restart button if [Docker actions](host-integrations.md#docker-actions) allow it |
| `/service <name>` | Whether a service tile is online and how fast it answers |
| `/weather <place>` or `/wetter <place>` | Weather for a place |
| `/calendar`, `/termine` | Your next appointments |
| `/waste`, `/muell` | The next waste collection |
| `/departures <stop>`, `/bus` | Next departures |
| `/parcel`, `/paket` | Parcels on their way |
| `/backup` | State of your monitored backup jobs |
| `/smart` | Disk health |
| `/speed` | Latest internet speed test |
| `/timeline` | Recent events |

An answer only appears for modules that are enabled and set up. A restart
from the search asks for confirmation first, exactly like the button in the
Docker widget.

## Quick capture

A short prefix turns the search field into an inbox:

| Type | Saves |
| --- | --- |
| `t call the bank` | a to-do |
| `t dentist tomorrow` | a to-do due tomorrow |
| `n idea for the garden` | a note |
| `b https://example.org` | a bookmark |

The long forms `todo:`, `note:` and `bm:` work as well. You can change the
short prefixes under Settings → Search.

## Saving from other pages

You do not have to open trisse-home to save something:

- **Bookmarklet.** Settings → Search offers a bookmarklet. Drag it to your
  browser's bookmarks bar; clicking it on any page opens a small dialog that
  saves the page as a bookmark, to-do or note.
- **Share menu.** When trisse-home is [installed as an app](apps-and-offline.md)
  on Android, it appears in the share menu. Sharing a link or text opens the
  same dialog.
- **Browser extension.** The [extension](apps-and-offline.md#browser-extension)
  adds *Save page to trisse-home* to the toolbar and the context menu.

Nothing is saved until you confirm the dialog.
