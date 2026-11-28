# Design guide

Every screen of Waylight uses the same small set of sizes, so the web app, the
mobile app and new designs look like one product. The values live as CSS
variables in `web/src/shared/theme/tokens.css`. Components use the variables,
never a number of their own.

The redesign mockups are the reference. Where this guide and the mockups
differ, the mockups win and this guide is changed to match. Values that the
mockups draw below the contrast rules are listed under Colour with the nearest
value that passes.

## Font sizes

| Token     | Size  | Line | Use                                                     |
| --------- | ----- | ---- | ------------------------------------------------------- |
| `--fs-xs` | 13 px | 16   | Hints, labels, meta text, axis labels                   |
| `--fs-sm` | 15 px | 24   | Everything else: text, fields, buttons, tiles           |
| `--fs-md` | 21 px | 28   | Card titles, numbers, dialog titles                     |
| `--fs-lg` | 32 px | 40   | Page titles, the clock and the temperature, setup steps |

Weights are 400, 500 and 600. Figtree is the face for the interface.
Fraunces is for titles from 21 px, large figures and the wordmark: weight 500,
the wordmark 600, and the page title with a letter spacing of -0.01 em. Times,
keycaps and axis labels use the system monospace face at `--fs-xs`. The sizes
are in `rem`, so the text size option in the accessibility settings scales
them. Numbers that line up use tabular figures. The line heights are 16, 24, 28 and 40 px (`--lh-xs` to `--lh-lg`). There is no fifth size.

## Corner radii

| Token         | Size  | Use                                              |
| ------------- | ----- | ------------------------------------------------ |
| `--radius-md` | 12 px | Cards, buttons, the search field, dialogs        |
| `--radius-sm` | 8 px  | Fields, icon buttons, tiles, blocks inside cards |

Chips, tabs, toggles and badges are pills or circles (`999px` or `50%`). The
checkbox uses 6 px (`--radius-box`) and the keycap and the small squares of the habit
week 4 px (`--radius-xs`). No other radius is used. A design sets `--radius` and `--radius-small`; these are the
names it saves, and `--radius-md` and `--radius-sm` read them.

## Spacing

Spacing follows a 4 px grid: 4, 8, 12, 16, 24, 32 and 48 px, as
`--space-1`, `--space-2`, `--space-3`, `--space-4`, `--space-6`, `--space-8`
and `--space-12`. The dashboard grid has 12 columns with a 16 px gap and 24 px
of padding around it. A card has 24 px of padding (16 px on a phone) and 16 px
between its parts.

## Colour

Amber is the only accent. It marks the primary button, checked boxes, the
avatar, due dates and the one series a chart is about. Everything else is
neutral: surfaces, borders and text come from one tinted ramp. There are two
looks, night (navy) and paper (warm light).

| Token                     | Night     | Paper     |
| ------------------------- | --------- | --------- |
| `--color-bg`              | `#131829` | `#f6f2e9` |
| `--color-surface`         | `#1b2238` | `#fffdf8` |
| `--color-surface-2`       | `#232b45` | `#f1ebdc` |
| `--color-border`          | `#2e3650` | `#e4dccb` |
| `--color-text`            | `#f2efe8` | `#1b2238` |
| `--color-text-muted`      | `#c9ced8` | `#4a5271` |
| `--color-text-subtle`     | `#a7aebb` | `#5f6988` |
| `--color-accent`          | `#f2b441` | `#f2b441` |
| `--color-accent-contrast` | `#1a1206` | `#1a1206` |
| `--color-accent-text`     | `#f2b441` | `#935800` |
| `--color-ok`              | `#8fd6a8` | `#1f7a45` |
| `--color-warning`         | `#f2b441` | `#8a5700` |
| `--color-danger`          | `#f2a199` | `#b3261e` |

Text has three roles. Primary text is `--color-text`. Muted text carries
descriptions and the second line of a row. Subtle text carries meta text, hints
and values on the right of a row. The hairline in the mockups is the text
colour at 9 % on night and 10 % on paper, and 16 % and 20 % for a stronger
line such as the border of a quiet button; `--color-border` holds the same
colours as solid values. Hover on the accent is `--color-accent-hover`,
mixed from the accent (`#e9a41f` on paper as in the mockups, about `#f9d380`
on night where the mockups draw `#ffd27a`), so a theme with another accent
keeps a matching hover. The stronger line is
`--color-border-strong`, the text colour at 16 % (night) or 20 % (paper).

Text and the accent used as text reach 4.5:1 on the surface they sit on. Two
mockup values fall short and use the next darker value in the app: the accent
as text on paper is `#935800` where the mockup draws `#a86400`, and a warning
as text on paper is `#8a5700` where the mockup draws `#a86400`.

The lit card has its own colours, set together by a theme.

| Token                | Night     | Paper     |
| -------------------- | --------- | --------- |
| `--color-lit`        | `#fff3d6` | `#131829` |
| `--color-lit-text`   | `#1b2238` | `#f2efe8` |
| `--color-lit-muted`  | `#525c7c` | `#c9ced8` |
| `--color-lit-accent` | `#855000` | `#f2b441` |

The mockup draws the muted text of the lit card as `#6b5f45` and its accent as
`#a86400` on night; both are darker in the table so that they reach 4.5:1. Its
inner lines are `#ead9ae` on night and `#2e3650` on paper, and its chips
`#fce2a6` and `#2b2414`. `--color-lit-warning` and `--color-lit-danger` colour
the dot of a weather warning there.

A status is always a dot plus a word, never colour alone. `StatusDot`
(`web/src/shared/ui/StatusDot.svelte`) draws it with four levels: ok, warning,
failed and stopped. The dot is 8 px; stopped is a hollow ring, so it differs
from a filled dot by shape as well as by colour.

## Surfaces

Depth comes from the step between background, surface and surface 2 and from a
1 px hairline. A card has no shadow. There is no glow, no coloured shadow and no
gradient from one hue to another; the only fade is the area under a chart line,
which fades the line's own colour to transparent. The page background is a
single colour.

Each screen has at most one lit card, such as the clock and weather card on the
Home tab (the Tonight widget). It inverts the scheme: warm paper on the night
theme, navy on the paper theme. It shows the time, the date, the current
weather, the active weather warning and a four-day forecast. Nothing else on
the screen competes with it.

## Icons

Icons are line icons from one set, 16 px with a 1.5 stroke and round ends.
They take the colour of the text beside them and never have a colour of their
own. Built-in labels, sample data and the demo contain no emoji. Emoji that a
person types into their own notes, habits or bookmarks stay as they are. The
logo is the lamp mark at 24 px next to the wordmark "Waylight" in Fraunces at
`--fs-sm`, weight 600.

## App frame

The top bar, the header and the widget grid sit in one frame: a 1 px hairline
in the border colour with `--radius-md` corners, at most 1440 px wide and
centred, with 24 px of page background around it. The frame has no fill of its
own, so the page and tab backgrounds show through, and menus may open past its
edge. It uses the theme's border colour, so every theme draws it in its own
tone and high contrast with its stronger line. Below 640 px there is no frame
and the page uses the full width. Kiosk links and wall displays show no frame
either.

## Top bar and header

The top bar is 64 px high with 24 px of padding at the sides, in three parts.
On the left the logo. In the middle one line at `--fs-xs` in the subtle colour
with the date, the temperature and the number of to-dos due today, separated by
small dots. On the right the status pill, the launcher, the edit button and
the avatar. The status pill is 32 px high with a hairline and a dot and opens
only the service status; the launcher (four squares) opens services and
bookmarks; the launcher and the edit button are 32 px icon buttons; the avatar is a 32 px amber circle with the initial in
`--color-accent-contrast`.

Under it the header has the page title on the left in Fraunces at `--fs-lg`,
and the tabs on the right of the same row. Below them the search field is 48 px
high and at most 520 px wide, on the surface colour with a hairline and
`--radius-md`. It has the line icon on the left and the keycap (`Ctrl K`,
13 px monospace, hairline, 4 px radius) on the right. With one tab no tab row
is shown.

Tabs are pills, 32 px high with 16 px of padding and weight 500. The selected
tab is filled with surface 2 and has a 1 px inner ring in the stronger
hairline; the others are muted text.

## Widgets

A card has the surface colour, a 1 px hairline, `--radius-md` and no shadow.

Every widget starts with `WidgetHeader`
(`web/src/shared/ui/WidgetHeader.svelte`), 24 px high: a 16 px line icon, the
title in sentence case at `--fs-sm` and weight 600 in the primary colour, then
optional meta text at `--fs-xs` in the subtle colour on the right, or one 32 px
icon button there. The full card and the compact view use the same header.
A widget hands its meta text to the card with `useHeaderMeta` (in the warning
or danger colour when it needs a look) and its add action with
`useHeaderAction`; outside a card it shows the action in its own body. The
help button sits right after the title and shows on hover.
Titles are never set in capitals with wide letter spacing.

Widgets list their items with `ListRow` (`web/src/shared/ui/ListRow.svelte`).
A row is at least 40 px high with 8 px of padding above and below and 12 px
between its parts, separated from the next by a 1 px hairline; the first row
has no line above it. It has the main text, an optional second line at
`--fs-xs` in the subtle colour, and a value on the right at `--fs-xs` in the
subtle colour. A value that needs attention, such as a due date, uses the
accent text colour and weight 500. Text that does not fit wraps or ends in an
ellipsis; it is never cut off.

Controls follow one set of measurements. A button inside a widget header or an
alert bar is 32 px high (`.btn-sm`). `Meter` draws a bar in the accent from
60 % or when it is marked `hot`, and in the warning and danger colours from
75 % and 90 %. A second, lighter segment after the fill shows a share that is
still in progress, such as the credits of running modules in Study progress.

| Control        | Size and look                                                             |
| -------------- | ------------------------------------------------------------------------- |
| Primary button | 40 px high, 16 px padding, `--radius-md`, accent fill, weight 600         |
| Quiet button   | The same size, no fill, stronger hairline, primary text, weight 500       |
| Icon button    | 32 px square, `--radius-sm`, subtle colour, surface 2 on hover            |
| Text field     | 40 px high, `--radius-sm`, page background, hairline, 12 px padding       |
| Checkbox       | 20 px, 6 px radius, 1.5 px stronger hairline; checked: accent fill        |
| Tile           | 32 px square, `--radius-sm`, surface 2, initials at `--fs-xs`, weight 600 |
| Chip           | 32 px high pill, 12 px padding, `--fs-xs`, weight 500                     |
| Progress bar   | 4 px high pill on surface 2, muted fill; accent for the one that matters  |

A widget has four sizes: small, medium, large and strip. They show different
content, not the same content scaled. Medium and large show the full widget.
Small and strip show a compact view of one or two lines, such as
"3 open · 1 overdue", and a click opens the full widget in a panel. A new
widget therefore needs both a full view and a compact view. The content scales
with the widget's own box.

## Charts

A screen has one main chart, clearly the largest data element, 176 px high on a
desktop and 96 px on a phone. It uses a 2 px line with round ends, a light area
fill under it, four 1 px hairlines as the grid and axis labels at `--fs-xs` in
monospace. The series the chart is about uses the accent; any other series is a
1.5 px line in the subtle colour. A legend is a 16 px by 2 px swatch with a
label at `--fs-xs`. Other widgets show data as lists and thin bars, or as small
rounded bars with one accent bar for the day that matters, not as a second
chart of the same weight. `LineChart` (`web/src/shared/ui/LineChart.svelte`)
draws the main chart. On the Server tab it is the internet speed chart, with a
24 hour and a 7 day range. The mockups set the axis labels of the chart at
11 px; the app keeps 13 px.

## Navigation

Above 640 px the tabs sit in the header as described above. At 640 px and
below they become a bottom navigation fixed to the bottom edge: an icon at
20 px and a label at `--fs-xs` per tab in a row of equal cells, 48 px high with
a hairline above. The selected tab uses the accent text colour; the others use
the subtle colour. The page leaves room for it, so no widget ends behind it.
The last entry is **Widgets** with the four-tile icon; it opens the widget
picker for the current tab and stays at the right edge when many tabs scroll.
Guests, kiosk displays and team tabs do not get it. With a single tab the bar
still shows on a phone, so the entry can be reached.

On a phone the top bar shows the logo, a short status pill and the avatar, and
no date line. The title stays at `--fs-lg`, the search field is 44 px high,
cards use 16 px of padding in one column, and the side gutter is 16 px.

## Copy

Text is in sentence case. Labels are not capitalised with wide letter
spacing. Short strings such as titles, labels and buttons have no full stop
and no exclamation mark. A change says what it compares to, for example
"in 4 days". English and German texts change together.

## Motion

Transitions run on colour, background, opacity and transform, in 150 ms. Nothing
loops without a reason: no pulsing, floating or shimmer unless it shows live
data. Motion is off when the system asks for reduced motion.

## Checks

`npm run lint` in `web/` runs stylelint, which rejects a `font-size` or
`border-radius` that is not one of the tokens and a `padding`, `margin` or
`gap` with a fixed size. Negative offsets and `calc()` are left alone.
