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
Gruvbox Dark, Solarized Light and Dark, a high-contrast theme and a classic portal
look. *Apply* switches to a theme at once.

Applying a theme replaces your light or dark choice, accent colour and custom
CSS. Your own settings are kept, and **Restore my own style** brings them
back.

Every theme can be downloaded as a small JSON file, and **Import theme…**
adds one somebody shared with you. Themes contain colours and corner radii
only. They cannot load fonts, images or scripts, so a theme file from a
stranger cannot track you or hide parts of the page. The
[theme format](https://github.com/itadori26/trisse-home/blob/main/themes/README.md)
explains how to write one and how to contribute it.

## Background image

A *background image URL* sets a picture behind all tabs. Each tab can have
its own background instead, set in edit mode (see
[tabs and widgets](dashboard.md#tabs)). The Astronomy Picture of the Day from
the [space module](modules/space.md) can also serve as a background.

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
