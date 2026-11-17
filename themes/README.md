# Themes

The theme gallery (Settings → Appearance, #111) ships the themes in
[`internal/modules/appearance/adapter/gateway/themes/`](../internal/modules/appearance/adapter/gateway/themes/).
They are embedded in the binary, so adding a theme is one JSON file.

## Format

```json
{
  "id": "nord",
  "name": "Nord",
  "author": "Arctic Ice Studio (MIT)",
  "scheme": "dark",
  "tokens": {
    "--color-bg": "#2e3440",
    "--color-surface": "#3b4252",
    "--color-text": "#eceff4",
    "--color-accent": "#88c0d0"
  }
}
```

- **id:** 2–40 characters `a-z`, `0-9`, `-`. It is also the file name (`nord.json`).
- **scheme:** `light` or `dark`. Applying the theme switches to that scheme, so native controls and scrollbars match. Left out in a theme with variants (below).
- **tokens:** only the design tokens of [`web/src/shared/theme/tokens.css`](../web/src/shared/theme/tokens.css), each with a fixed syntax:
  - colors as `#rgb` or `#rrggbb`: `--color-bg`, `--color-surface`, `--color-surface-2`, `--color-border`, `--color-text`, `--color-text-muted`, `--color-accent`, `--color-accent-contrast`, `--color-accent-text`, `--color-danger`, `--color-success`, `--color-warning`;
  - radii as `0px`–`32px`: `--radius`, `--radius-small`;
  - fonts by keyword, never by file or URL: `--font-sans`, `--font-display` take `system`, `figtree`, `fraunces` or `serif`. Figtree and Fraunces ship with the app;
  - shadows: `--shadow` (cards) and `--shadow-pop` (menus and panels) take `none` or up to three layers of `x y blur [spread] #rrggbb[aa]`, for example `0 1px 2px #1b22381a`;
  - `--glow`: `none` or `0 0 <blur>px #rrggbb[aa]`, a soft light around the widget area;
  - `--page-gradient`: `none` or `linear-gradient(180deg, #rrggbb, #rrggbb)`, shown behind the page when no background picture is chosen.

  `--color-bg`, `--color-surface`, `--color-text` and `--color-accent` are required. Set all colors for a polished result.

### Light and dark in one theme

A theme can carry a variant for each scheme. It then follows the account's light, dark or system choice instead of switching it. `tokens` holds what both variants share, `variants.light` and `variants.dark` hold the rest, and each variant together with the shared tokens must set the required ones:

```json
{
  "id": "harbour",
  "name": "Harbour",
  "tokens": { "--font-display": "fraunces", "--radius": "14px" },
  "variants": {
    "light": { "--color-bg": "#f4f1ea", "--color-surface": "#ffffff", "--color-text": "#1d2433", "--color-accent": "#2f5f8a" },
    "dark": { "--color-bg": "#121722", "--color-surface": "#1b2230", "--color-text": "#eef0f4", "--color-accent": "#8db8e0" }
  }
}
```

Free CSS, layout rules, font files, `url(…)` and unknown fields are refused, for built-in and imported themes alike. Updates to the app therefore cannot break a theme, and a shared theme file cannot load anything or hide parts of the page.

## Contributing a theme

1. Add `<id>.json` to the folder above.
   - Contrast must reach WCAG AA (4.5:1). This covers text and muted text on `--color-bg`/`--color-surface`/`--color-surface-2`, accent and danger on `--color-surface`, the accent text colour (or the accent, if not set) on `--color-surface` and `--color-surface-2`, and `--color-accent-contrast` on `--color-accent`, for every variant. A test enforces it.
   - Credit the original palette and its license in `author`.
2. Run `go test ./internal/modules/appearance/...`. The tests load every file and fail on an invalid one.
3. Open a pull request with a screenshot of the dashboard in the new theme.

## Sharing your own

In the gallery, **⬇** downloads any theme as a file and **Import theme…** adds one (at most 16 KB, 50 imported themes).

Applying a theme replaces your custom CSS and accent color, and a single-scheme theme also your light/dark choice. Your own values are kept, and **Restore my own style** brings them back.
