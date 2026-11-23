# Design guide

Every screen of Waylight uses the same small set of sizes, so the web app, the
mobile app and new designs look like one product. The values live as CSS
variables in `web/src/shared/theme/tokens.css`. Components use the variables,
never a number of their own.

## Font sizes

| Token     | Size  | Use                                                         |
| --------- | ----- | ----------------------------------------------------------- |
| `--fs-xs` | 13 px | Hints, labels, meta text                                    |
| `--fs-sm` | 15 px | Everything else: text, fields, buttons, tiles               |
| `--fs-md` | 21 px | Card titles, numbers, dialog titles                         |
| `--fs-lg` | 32 px | Page titles, including the welcome page and the setup steps |

Weights are 400, 500 and 600. Figtree is the face for the interface;
Fraunces is only for titles from 21 px. The sizes are in `rem`, so the text
size option in the accessibility settings scales them.

## Corner radii

| Token         | Size  | Use                                       |
| ------------- | ----- | ----------------------------------------- |
| `--radius-md` | 12 px | Tiles, cards, buttons, dialogs            |
| `--radius-sm` | 8 px  | Fields, blocks inside cards, small things |

Chips, toggles and icon badges are pills or circles (`999px` or `50%`). No
other radius is used. A design sets `--radius` and `--radius-small`; these are
the names it saves, and `--radius-md` and `--radius-sm` read them.

## Spacing

Spacing follows a 4 px grid: 4, 8, 12, 16, 24, 32 and 48 px, as
`--space-1`, `--space-2`, `--space-3`, `--space-4`, `--space-6`, `--space-8`
and `--space-12`.

## Colour

Amber is the only accent. It marks the primary button, the selected tab,
checked boxes, due dates and the one series a chart is about. Everything
else is neutral: surfaces, borders and text come from one tinted ramp.

| Token                     | Night     | Paper     |
| ------------------------- | --------- | --------- |
| `--color-bg`              | `#131829` | `#f6f2e9` |
| `--color-surface`         | `#1b2238` | `#fffdf8` |
| `--color-surface-2`       | `#232b45` | `#f1ebdc` |
| `--color-border`          | `#2e3650` | `#e4dccb` |
| `--color-text`            | `#f2efe8` | `#1b2238` |
| `--color-text-muted`      | `#a7aebb` | `#5f6988` |
| `--color-accent`          | `#f2b441` | `#f2b441` |
| `--color-accent-contrast` | `#1a1206` | `#1a1206` |
| `--color-accent-text`     | `#f2b441` | `#935800` |
| `--color-lit`             | `#fff3d6` | `#1b2238` |
| `--color-lit-text`        | `#1b2238` | `#f2efe8` |
| `--color-lit-muted`       | `#525c7c` | `#a7aebb` |
| `--color-lit-accent`      | `#855000` | `#f2b441` |

Text and the accent used as text reach 4.5:1 on the surface they sit on.
The `--color-lit-*` tokens belong to the lit card (see Surfaces) and keep that
ratio on it; `--color-lit-warning` and `--color-lit-danger` colour the dot of a
weather warning there.
Status colours carry meaning only: green for healthy, amber for a warning,
red for a failure. A status is always a dot plus a word, never colour alone.

## Surfaces

Depth comes from the step between background, surface and surface 2, and
from a 1 px border. There is no glow, no coloured shadow and no gradient
from one hue to another. The page background is a single colour.

Each screen has at most one lit card, such as the clock and weather card on
the Home tab (the Tonight widget). It inverts the scheme: warm paper
(`#fff3d6` with ink `#1b2238`) on the night theme, navy on the paper theme. Nothing else on the
screen competes with it.

## Icons

Icons are line icons from one set, 16 px with a 1.5 stroke. They take the
colour of the text beside them and never have a colour of their own. Built-in
labels, sample data and the demo contain no emoji. Emoji that a person types
into their own notes, habits or bookmarks stay as they are.

## Widgets

A widget header has a line icon, the title in sentence case at `--fs-sm` and
weight 600, and at most one icon button on the right. Titles are never set in
capitals with wide letter spacing.

A list row is at least 40 px high, separated from the next by a 1 px line. It
has the main text, an optional second line at `--fs-xs` in the muted colour,
and a value on the right. A value that needs attention, such as a due date,
uses the accent text colour. Text that does not fit wraps or ends in an
ellipsis; it is never cut off.

Numbers that line up use tabular figures. The clock and other large numbers
use Fraunces at `--fs-lg`; there is no fifth size.

## Charts

A screen has one main chart, clearly the largest data element. It uses a
smooth line with a light area fill, a muted grid and axis labels at
`--fs-xs`. The series the chart is about uses the accent; any other series
uses a neutral step. Other widgets show data as lists and thin 4 px bars, not
as a second chart of the same weight.

## Copy

Text is in sentence case. Labels are not capitalised with wide letter
spacing. Short strings such as titles, labels and buttons have no full stop
and no exclamation mark. A change says what it compares to, for example
"in 4 days". English and German texts change together.

## Motion

Transitions run on colour, background, opacity and transform, in 150 to
200 ms. Nothing loops without a reason: no pulsing, floating or shimmer
unless it shows live data. Motion is off when the system asks for reduced
motion.

## Checks

`npm run lint` in `web/` runs stylelint, which rejects a `font-size` or
`border-radius` that is not one of the tokens and a `padding`, `margin` or
`gap` with a fixed size. Negative offsets and `calc()` are left alone.
