# Electricity prices

**What it does:** Shows the day-ahead electricity price per hour for today and, once it is published (usually around 13:00), for tomorrow, as a bar chart. The three cheapest hours of each day are highlighted, so you can plan the washing machine, the dishwasher or charging the car (#288).

## Configuration

Add the **Electricity prices** widget and choose the bidding zone in its settings: **Germany/Luxembourg (DE-LU)** or **Austria (AT)**. No account or API key is needed.

## What the numbers mean

The prices are the **market prices** of the EPEX day-ahead auction in cent per kWh, without taxes, grid fees or your supplier's margin. With a dynamic tariff your price follows them; the absolute amount on your bill is higher. Since October 2025 the market uses 15-minute intervals; the widget shows the average of each hour. Prices can be negative.

## Outbound requests

The public market data API of [aWATTar](https://www.awattar.de) (`api.awattar.de` or `api.awattar.at`), only while a widget shows prices and at most every 15 minutes per zone. When it does not answer, the last known prices stay visible and are marked as outdated. `ELECTRICITY_API_URL` points the module elsewhere (for tests).

## API

`GET /api/electricity/prices?zone=DE-LU` – see `api/modules/electricity.yaml`.
