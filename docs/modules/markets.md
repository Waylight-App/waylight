# Markets

**What it does:** A small watchlist: currencies, shares, indices and crypto with the latest value, the change against the previous close and a 30-day sparkline (#284).

## Configuration

Add the **Markets** widget and list the symbols in its settings, one per line (at most 12):

- `fx:USD`, `fx:CHF`, `fx:GBP` … – the euro in that currency, from the **ECB reference rates** (published on working days around 16:00 CET).
- `stooq:aapl.us`, `stooq:sap.de`, `stooq:^dax`, `stooq:btcusd` … – shares, indices and crypto in the notation of [Stooq](https://stooq.com) (search a symbol there; the part in the address after `s=` is the code).

No account or API key is needed.

## What the numbers mean

The values are **daily closes** (the ECB rate of the day), not live prices. They are for a quick look, not for trading.

## Currency conversion in the search

The search converts currencies with the same ECB reference rates: type
`50 usd in eur`, `€20 to $` or `100 Franken in Euro`. ISO codes work for
every currency the ECB publishes, and common symbols and names (`€`, `$`,
`£`, `¥`, euro, dollar, pound, Pfund, franc, Franken, …) for the usual ones.
The answer shows the date of the rate.

The rates are fetched when the search is first opened and kept on the
server for a day. When the ECB cannot be reached, the search uses the last
rates the server has and says so with their date; right after a restart
without internet it says that no rates are available. Installations without
internet switch this off with `CURRENCY_RATES=false`; the markets widget
keeps working with whatever it can reach.

## Outbound requests

Only while a markets widget is shown and at most once an hour per symbol: the ECB's 90-day reference rate file (`www.ecb.europa.eu`) and one CSV per Stooq symbol (`stooq.com`). When a source does not answer, the last known value stays visible and is marked as outdated. `MARKETS_ECB_URL` and `MARKETS_STOOQ_URL` point the module elsewhere (for tests).

## API

- `GET /api/markets/rates`: the newest reference rates, `{base: "EUR", date, rates: {USD: 1.08, …}, stale}`; `stale` is true when they are the last known ones because the ECB could not be reached. 404 with `CURRENCY_RATES=false`.


`GET /api/markets/watchlist?symbols=fx:USD,stooq:aapl.us` – see `api/modules/markets.yaml`.
