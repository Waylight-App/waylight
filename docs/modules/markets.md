# Markets

**What it does:** A small watchlist: currencies, shares, indices and crypto with the latest value, the change against the previous close and a 30-day sparkline (#284).

## Configuration

Add the **Markets** widget and list the symbols in its settings, one per line (at most 12):

- `fx:USD`, `fx:CHF`, `fx:GBP` … – the euro in that currency, from the **ECB reference rates** (published on working days around 16:00 CET).
- `stooq:aapl.us`, `stooq:sap.de`, `stooq:^dax`, `stooq:btcusd` … – shares, indices and crypto in the notation of [Stooq](https://stooq.com) (search a symbol there; the part in the address after `s=` is the code).

No account or API key is needed.

## What the numbers mean

The values are **daily closes** (the ECB rate of the day), not live prices. They are for a quick look, not for trading.

## Outbound requests

Only while a markets widget is shown and at most once an hour per symbol: the ECB's 90-day reference rate file (`www.ecb.europa.eu`) and one CSV per Stooq symbol (`stooq.com`). When a source does not answer, the last known value stays visible and is marked as outdated. `MARKETS_ECB_URL` and `MARKETS_STOOQ_URL` point the module elsewhere (for tests).

## API

`GET /api/markets/watchlist?symbols=fx:USD,stooq:aapl.us` – see `api/modules/markets.yaml`.
