# Design guide

Every screen of Waylight uses the same small set of sizes, so the web app, the
mobile app and new designs look like one product. The values live as CSS
variables in `web/src/shared/theme/tokens.css`. Components use the variables,
never a number of their own.

## Font sizes

| Token | Size | Use |
|---|---|---|
| `--fs-xs` | 13 px | Hints, labels, meta text |
| `--fs-sm` | 15 px | Everything else: text, fields, buttons, tiles |
| `--fs-md` | 21 px | Card titles, numbers, dialog titles |
| `--fs-lg` | 32 px | Page titles, including the welcome page and the setup steps |

Weights are 400, 500 and 600. Figtree is the face for the interface;
Fraunces is only for titles from 21 px. The sizes are in `rem`, so the text
size option in the accessibility settings scales them.

## Corner radii

| Token | Size | Use |
|---|---|---|
| `--radius-md` | 12 px | Tiles, cards, buttons, dialogs |
| `--radius-sm` | 8 px | Fields, blocks inside cards, small things |

Chips, toggles and icon badges are pills or circles (`999px` or `50%`). No
other radius is used. A design sets `--radius` and `--radius-small`; these are
the names it saves, and `--radius-md` and `--radius-sm` read them.

## Spacing

Spacing follows a 4 px grid: 4, 8, 12, 16, 24, 32 and 48 px, as
`--space-1`, `--space-2`, `--space-3`, `--space-4`, `--space-6`, `--space-8`
and `--space-12`.

## Checks

`npm run lint` in `web/` runs stylelint, which rejects a `font-size` or
`border-radius` that is not one of the tokens.
