# Meter readings

**What it does:** Keeps the readings of your electricity, gas, water and heat meters and shows how much you used per month and per year.

## Using it

Add the *Meter readings* widget and create a meter with **+ New meter**: a name, a kind (electricity, gas, water, heat or other) and the unit the counter shows, for example kWh or m³.

The widget shows one meter at a time; the selector above it switches between meters. To enter a reading, pick the day and type the counter value. The form works on a phone, so you can enter the value while standing at the meter. A reading on a day that already has one replaces it.

A counter only counts up, so a reading cannot be lower than an earlier one or higher than a later one. A reading cannot lie in the future.

## What the numbers mean

The consumption is computed from the readings. Between two readings the counter is assumed to run at an even pace, so a reading in the middle of December and one in the middle of January give the value at the turn of the year, and the year starts and ends exactly on 1 January.

- **Per month:** shown as a bar for every month whose first day and the first day of the next month lie between the first and the last reading. The month you are in has no bar until a reading on or after the first of the next month exists.
- **Per year:** the consumption between the first reading and the last reading of the year, or the whole year when the readings cover it. A year that is not covered yet is marked as such. Use the year selector to look at earlier years.

One reading on the first of every month gives the cleanest chart.

## Compact view

The compact view names the meter that was read longest ago and how many days ago that was.

## Who sees what

New meters are shared with everyone, since meters belong to the household. The person who creates a meter can narrow this to themselves or a group. Everyone who sees a meter can add readings. Only the person who created the meter can change or remove it and remove its readings.

## Configuration

No environment variables. `MODULES_DISABLED=meters` switches the module off.

## Outbound requests

None.

## API

`/api/meters` (list, create), `/api/meters/{id}` (read with the consumption of `?year=`, update, delete), `/api/meters/{id}/readings` (add) and `/api/meters/{id}/readings/{day}` (remove) and `/api/meters/compact` – see `api/modules/meters.yaml`.
