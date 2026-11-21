# Places

**What it does:** you enter a place once, by address search or with your
browser's location, and choose it in every widget that needs a location.

## Using places

Open **Settings → Connections & services → Places**. There, search for an
address or town, or press *Use my location*, give the place a name such as
"Home" and save it. Each place shows which widgets use it; you can rename or
remove it there. Removing a place that widgets use asks first and names
them. Those widgets keep the location they had, they just no longer follow
the place.

In the settings of the weather and rain radar widgets and in Settings →
Fuel prices, *Saved place* lists your places. Choosing one fills in the
location; the coordinates stay out of the way under *Advanced*. A location
found by the search can be kept with *Save as a place*.

Places belong to the account that saved them. Like other personal content
they are visible to the other accounts of the instance by default, but only
their owner can change them. One account keeps at most 50 places.

Departures still choose a stop by name, and the waste calendar reads an
ICS calendar, so neither takes a place yet.

## Outbound requests

None of its own. The address search uses the weather module's geocoder
(Open-Meteo); without the weather module only the browser location is
offered.

## API

`GET /api/places` lists the places with the widgets on your own tabs that
use them (`usedBy`); `POST /api/places`, `PUT /api/places/{id}` and
`DELETE /api/places/{id}` change them. A widget names its place with
`placeId` in its settings, next to the coordinates it copied. See
`api/modules/places.yaml`.
