# Photo frame

**What it does:** Shows your own pictures one after another, like a digital photo frame. The pictures are uploaded to Waylight itself, so no photo server is needed (#286). If you run Immich, the [Immich widget](immich.md) shows pictures from there instead.

## Configuration

Add the **Photo frame** widget and open its settings:

- **Album:** each widget shows one album. Type a new name to start one; several widgets can share an album.
- **Upload** one or more pictures (JPEG, PNG or GIF, up to 25 MB each). Move the thumbnails with the arrow buttons to change the order, or delete them.
- **Interval** between two pictures (5 seconds to one hour) and **Fit**: *show whole picture* leaves bars where the shape does not match, *fill widget* crops the edges.

The widget pauses while the browser tab is hidden and fades without motion when the system asks for reduced motion.

## Storage and backups

Every upload is turned upright (phone photos store their rotation separately), scaled down to at most 1920 pixels on the longer side and saved as JPEG. That also removes the EXIF data, including the location where a photo was taken. A picture then takes about 0.2–0.6 MB.

All albums together may hold **300 MB or 2,000 pictures**; the settings show how much is used. The pictures live in the database, so **backups contain them**. Data exports (the JSON file under *Settings → Data*) do not, because they would become very large; keep the original photos elsewhere or rely on backups.

## Outbound requests

None. The pictures only ever go to your own browser.

## API

`GET /api/photos?album=…`, `GET /api/photos/albums`, `POST /api/photos?album=…`, `PUT /api/photos/order`, `DELETE /api/photos/{id}`, `GET /api/photos/{id}/image` – see `api/modules/photos.yaml`.
