# Appearance

Settings → Appearance controls how the dashboard looks and which language it
speaks. Settings are stored on the server, so every browser you sign in
with looks the same.

## Language

trisse-home speaks English and German. By default it follows the language of
your browser; you can fix one under *Language*. Dates, times and numbers
follow the chosen language. Messages from the server, such as validation
errors, are translated too.

## Light, dark and accent colour

*Theme* switches between light, dark and *System*, which follows your
operating system and changes with it in the evening. The *Accent color*
tints buttons, links and highlights. Both built-in colour schemes meet WCAG
AA contrast; a very light or very dark accent colour can lower it, so check
that text on buttons stays readable.

## Theme gallery

The gallery offers ready-made themes: Nord, Catppuccin Latte and Mocha,
Gruvbox Dark, Solarized Light and Dark, Dracula, Tokyo Night, One Dark, a
high-contrast theme and a classic portal look. *Apply* switches to a theme at once.

Applying a theme replaces your accent colour and custom CSS. A theme made
for light or for dark also switches to that scheme; a theme marked *Light and
dark* brings both and follows your choice, including *System*. Your own
settings are kept, and **Restore my own style** brings them back.

Every theme can be downloaded as a small JSON file, and **Import theme…**
adds one somebody shared with you. Themes set colours, corner radii, shadows
and one of the fonts that come with the app, each in a fixed format. They
cannot load fonts, images or scripts, so a theme file from a stranger cannot
track you or hide parts of the page. The
[theme format](https://github.com/itadori26/trisse-home/blob/main/themes/README.md)
explains how to write one and how to contribute it.

## Background image

The background picker opens from **Settings → Appearance → Background** and,
in edit mode, from the **Background** button in the action bar or the tab's
menu. It offers three kinds of backgrounds and shows a live preview before
you save:

- **Built in:** a few gradients and plain colours. They are drawn with CSS,
  so they cost nothing and work offline.
- **Your pictures:** upload a JPEG, PNG or GIF of up to 25 MB. trisse-home
  turns it upright, scales it to at most 2560 pixels and stores it as JPEG
  without EXIF data, so the location where a photo was taken is not kept.
  Up to 30 pictures are stored in the database, which means backups contain
  them; the JSON data export does not. Deleting a picture that is in use
  brings back the default background.
- **Image from the web:** any http(s) address of a picture, as before. The
  Astronomy Picture of the Day from the [space module](modules/space.md) works
  here too.

Opened from the dashboard, the picker asks whether the background is for
**all tabs** or **only this tab**. A tab's own background always wins over the
one for all tabs; the picker says so and offers to remove the tab's own
background when you change the one for all tabs. **Default background**
(or, for a tab, **Use the background of all tabs**) resets the choice.

## Greeting and top bar

Below the top bar you can show a greeting that changes with the time of day
("Good morning, Alex"), today's date and the weather from your first weather
widget. Enter your name under *Your name in the greeting*, or leave it empty
for a greeting without a name.

The launcher, the grid icon in the top bar, opens a quick list of your
service tiles, your bookmarks or both; choose which under *Appearance*.

## Custom CSS

For changes the settings do not cover, *Custom CSS* applies your own styles
on top of the theme. The dashboard's design tokens are CSS variables such as
`--color-accent` and `--radius`; overriding them survives updates better than styling
individual elements, whose class names can change between releases.

```css
:root {
  --radius: 4px;
}
```

Custom CSS is not checked for contrast. If you rely on a screen reader or
high contrast, prefer the built-in themes.

### Design tokens

These are the variables themes and custom CSS can set. Components use
nothing else, so changing a token changes every place that uses it.

| Token | What it is for |
|---|---|
| `--color-bg` | Page background behind the widgets |
| `--color-surface`, `--color-surface-2` | Widget cards; inputs and nested areas inside them |
| `--color-border` | Lines and outlines |
| `--color-text`, `--color-text-muted` | Body text; secondary text |
| `--color-accent` | Fill of buttons, the active tab and toggles |
| `--color-accent-contrast` | Text on the accent fill |
| `--color-accent-text` | The accent as text or icon colour on surfaces: links, highlighted values, the current day. Follows `--color-accent` unless set. Set it when the accent is light, such as yellow or amber, which reads well as a fill but not as text on a light card |
| `--color-danger`, `--color-success`, `--color-warning` | Status colours, also used as text |
| `--radius`, `--radius-small` | Corner radius of cards; of buttons and inputs |
| `--shadow`, `--shadow-pop` | Shadow of cards; of menus and panels that float above the page |
| `--glow` | A soft light around the widget area, `none` by default |
| `--page-gradient` | A gradient behind the page when no background picture is chosen, `none` by default |
| `--font-sans` | Body text |
| `--font-display` | Headings and large numerals. Follows `--font-sans` unless set |
| `--font-mono` | Code and keyboard hints |

Two typefaces come with the app and can be used in these font tokens without
any request to another server: `'Figtree Variable'`, a sans serif, and
`'Fraunces Variable'`, a serif for display text. The browser downloads a font
file only when text on the page uses it.

```css
:root {
  --font-display: 'Fraunces Variable', Georgia, serif;
}
```

Every built-in theme keeps text colours, including `--color-accent-text`, at
a contrast of at least 4.5:1 to the surfaces they appear on (WCAG AA); a test
checks this.
