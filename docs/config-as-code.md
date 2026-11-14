# Settings as a YAML file

trisse-home can write its settings into one readable file, `trisse-home.yaml`, and read them back. Keep that file in Git, review changes before you apply them, or set up a second instance the same way as the first. This page explains what the file contains, how to export and apply it, and where its limits are.

## What the file contains

The file has five sections. Each one belongs to a module and is written in the order below:

| Section | Contents |
|---|---|
| `appearance` | Colour scheme, accent colour, background image, language, greeting, launcher, profile and custom CSS |
| `search` | Web search provider, whether results open in a new tab, the quick capture prefixes and your `!` shortcuts |
| `services` | The service tiles you created yourself, in display order |
| `bookmarks` | Bookmarks without a folder, then each folder with its bookmarks |
| `dashboard` | Tabs with their background, default-tab rules and widgets, including each widget's settings and its place in the phone layout |

The file never contains passwords, API tokens or other secrets, and nothing that changes on its own such as notes, to-dos or feed items. Those belong in [backups](backups.md) and the JSON export.

A short example:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/itadori26/trisse-home/main/docs/schema/trisse-home.v1.schema.json
version: 1
appearance:
  theme: dark
  accentColor: "#3b6ef5"
services:
  - name: Jellyfin
    url: https://jf.${HOME_DOMAIN}
    icon: 🎬
bookmarks:
  - title: Router
    url: http://192.168.178.1
  - folder: Work
    bookmarks:
      - { title: Wiki, url: "https://wiki.${HOME_DOMAIN}", tags: [docs] }
dashboard:
  tabs:
    - name: Home
      rules: [{ days: [1, 2, 3, 4, 5], from: "07:00", to: "09:00" }]
      widgets:
        - kind: clock
          at: { x: 0, y: 0, w: 4, h: 2 }
        - kind: bookmarks
          at: { x: 4, y: 0, w: 4, h: 4 }
          phone: { order: 1, rows: 4 }
```

Things are referred to by name, not by internal IDs, and lists keep their order. Exporting, applying and exporting again gives exactly the same file, so a diff in Git only shows what you really changed.

## Export the file

Open **Settings → Import & export** and choose **Download trisse-home.yaml**. Or use the API with an [API token](api-tokens.md) that has *read* access to **transfer**:

```sh
curl -H "Authorization: Bearer $TRISSE_TOKEN" https://home.example.org/api/transfer/config > trisse-home.yaml
```

## Apply a file

On the same settings page, choose **Apply trisse-home.yaml** and pick the file, or send it to the API with a token that has *read & write* access to **transfer**:

```sh
curl -X POST -H "Authorization: Bearer $TRISSE_TOKEN" -H "Content-Type: application/yaml" \
  --data-binary @trisse-home.yaml https://home.example.org/api/transfer/config
```

trisse-home first reads the whole file. A YAML error, an unknown section or field (often a typo), a missing environment variable or a newer file version rejects the file, and nothing changes.

After that check, each section in the file replaces the settings of its module completely: a tile that is not in `services` is removed, a tab that is not in `dashboard` is removed with its widgets. A section that is missing from the file is not touched, so a file with only `bookmarks:` changes only the bookmarks. If one section fails its own checks, for example because a widget belongs to a disabled module, that section stays as it was and the answer names the reason; the other sections are still applied.

Some things are kept on purpose. A service tile or a tab with the same name as before keeps its internal ID, so uptime history and guest access to a tab survive. A bookmark with the same URL keeps its fetched icon and link check. Service tiles that came from Docker labels are not part of the file and are left alone.

## Secrets and private host names

Any text value may contain `${NAME}`. When the file is applied, trisse-home replaces it with the environment variable `NAME` of the server process, for example from your `.env` file. That keeps private domains and similar values out of the repository. An unset variable is an error rather than an empty string. Write `$${NAME}` if you need the literal text `${NAME}`.

Variables only work inside text values. A number such as a widget's coordinates cannot come from the environment.

## Editor support

The first line of an exported file points to a [JSON Schema](https://raw.githubusercontent.com/itadori26/trisse-home/main/docs/schema/trisse-home.v1.schema.json). Editors with the YAML language server, such as VS Code with the Red Hat YAML extension, use it to complete keys and mark mistakes while you type. Widget settings (`config`) are free-form there, because each widget defines its own; the server checks them when the file is applied.

## Limits

- Widget settings are copied as they are stored. A widget that points to something by ID, such as a single service, still points to that ID on another instance.
- The file has no secrets, so integration passwords and tokens have to be entered again on a new instance.
- Every widget gets a new internal ID when its tab is applied.
- A mode in which the file is applied at every start and the UI cannot change those settings is planned as a follow-up to [#88](https://github.com/itadori26/trisse-home/issues/88).
