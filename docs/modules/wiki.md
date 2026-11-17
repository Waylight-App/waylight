# Wiki search

**What it does:** Finds pages of your wikis from the search field of the start page (#321). Supported are **BookStack**, **Outline** (#322) and **Wiki.js** (#323). Results show the title, a short excerpt, the book, collection or language they belong to, and open the page in the wiki.

The search is **federated**: Waylight asks the wikis while you type and stores no page content, so there is no index to keep up to date and nothing grows in the database.

## Configuration

Under **Settings → Wiki search** connect up to five wikis: a name, the wiki software, the address and an API token.

- **BookStack:** in your profile under *API tokens* create a token and enter it as `token ID:token secret`. The account needs the permission to use the API.
- **Outline:** *Settings → API & apps → New API key*.
- **Wiki.js:** *Administration → API access*, enable the API and create a key; read access is enough.

*Test* sends a sample search to the wiki.

**Who sees what:** everyone who uses this start page sees what the API token may read. Use a token of an account that can only read what all of you may see. Searching with each person's own permissions is planned as a team feature (#327 and following).

## Searching

Type at least two characters into the search field. Wiki pages appear as their own group below bookmarks, services, notes and to-dos. Type `w ` (or `wiki `) first to search only the wikis, for example `w vpn`.

The wikis are asked after a short pause, all at the same time. A wiki that is slow or does not answer is named below the results and never holds up the other results; each wiki gets at most four seconds. The same search within a minute is answered from a short cache.

## Outbound requests

Only to the wikis you connect, and only while you search (plus one test request when you press *Test*). For Outline the names of the collections are fetched once every ten minutes.

## Limits

- Five results per wiki, five wikis.
- Links in results are only followed when they point to the wiki's own address.
- Gateway tests use responses shaped like the documented APIs of BookStack, Outline and Wiki.js 2, not recordings of real instances.

## API

`GET /api/wiki/search?q=…`, `GET`/`POST /api/wiki/wikis`, `PUT`/`DELETE /api/wiki/wikis/{id}`, `POST /api/wiki/wikis/{id}/test` – see `api/modules/wiki.yaml`.

## Team wikis with each person's account

In a server built with the `ee` features and a licence key that includes
`workplace-search` ([licence keys](../licence-keys.md)), the search also asks
Confluence (Atlassian Cloud), SharePoint and OneDrive (Microsoft 365),
Notion, and messages in Slack and Microsoft Teams, each with the account the searching person connected under
*Settings → Connected accounts* ([connected accounts](../connected-accounts.md)).
The services answer with what that person may see, so two people searching
the same words can get different results.

- Results of these services are never cached and never shared between
  accounts; the wikis an admin connected above are cached as before.
- A service the person has not connected is skipped without a message; an
  expired connection is skipped too, and the connected accounts page offers
  to reconnect.
- Confluence searches every site the Atlassian account can reach (pages and
  blog posts). Microsoft Graph searches files and list items. Notion's search
  matches page titles only, not their text.
- Slack messages come from the Real-time Search API with the person's user
  token: public and private channels, group and direct messages they are in.
  Slack allows about ten searches per person and minute; Waylight asks at
  most eight times a minute per person and otherwise lists Slack as not
  reachable for that search instead of running into Slack's limit.
- Teams messages come from the Microsoft Graph search: chats and channels the
  person is in, with a link into Teams. Graph cannot sort messages, so they
  come in Microsoft's order.
- No message or page is stored; results exist only in the answer to the
  search.
- Background jobs and API tokens never search these services.
