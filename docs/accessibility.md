# Accessibility

Waylight aims for **WCAG 2.2 level AA** (#113). Everything can be used with a keyboard and a screen reader.

## Options for each account

**Settings → Appearance → Accessibility** has three options. They belong to
the account that sets them: a household member who switches on large text
does not change anybody else's dashboard, and a new member does not inherit
the options of the admin.

- **Text size:** *Normal*, *Large* (115 %) or *Larger* (130 %). Text and the
  spacing around it grow together, so buttons and fields get bigger too. On
  the free grid widgets keep their cells; a small tile whose full view no
  longer fits switches to its compact view.
- **High contrast:** black text on white, or white on black in the dark
  scheme, with strong accent and status colours. Text reaches WCAG AAA
  (7:1) on every surface. While it is on, it replaces the gallery theme, the
  accent colour and the background picture; switching it off brings them
  back.
- **Simple view:** one column with the short state of each widget, larger
  buttons and fields (at least 44 × 44 px) and no edit mode. Widgets with
  nothing to show are left out. A widget that has no short state, such as
  the clock or notes, shows its full view. The tabs and the search stay. To
  change the dashboard, switch the simple view off again.

The options are part of `GET /api/appearance` (`textSize`, `highContrast`,
`simpleView`), so the mobile app can follow them.

## What is checked automatically

- **axe-core in the Playwright suite** (`web/e2e/a11y.ts`) runs the WCAG 2.0–2.2 A/AA rules on:
  - login, dashboard and edit mode;
  - the widget picker and the shortcuts overview;
  - **every settings section**;
  - the new-tab, tab rules and widget settings dialogs, the tab menu and the background picker;
  - the dashboard in dark mode and on a phone.

  A new violation fails CI.

- **Contrast unit tests:**
  - The light and dark design tokens must reach 4.5:1 for text: normal and muted text on all surfaces, accent and status colours on surfaces, and text on accent (`web/src/shared/theme/tokens.test.ts`).
  - The same check runs for every built-in theme of the gallery (`TestBuiltinThemesMeetWCAGAA`).
  - The high-contrast colours must reach 7:1 (WCAG AAA) for the same pairs, light and dark.
- Components may only use colour tokens that exist. An undefined token would fall back to a colour with poor contrast, and a test catches it.

## Keyboard

| Keys                                | Action                                                               |
| ----------------------------------- | -------------------------------------------------------------------- |
| `Ctrl+K` or `/`                     | Open the quick search                                                |
| `1` … `9`                           | Switch to tab 1 to 9                                                 |
| `e`                                 | Start or finish edit mode                                            |
| `?`                                 | Show all shortcuts                                                   |
| `Ctrl+Z`, `Ctrl+Shift+Z` / `Ctrl+Y` | Undo and redo in edit mode                                           |
| Arrow keys / `Shift` + arrow keys   | Move / resize the focused widget in edit mode                        |
| Arrow keys, `Home`, `End`           | Navigate the widget picker                                           |
| `Esc`                               | Close dialogs, menus and the launcher; focus returns to where it was |

Shortcuts are ignored while you type in a field or a dialog is open.

The `?` overview in the app is the reference; this table can lag behind it. Dialogs are native `<dialog>` elements: focus stays inside and `Esc` closes them. Every focusable element shows a visible focus ring.

## Screen readers

- Widgets have headings, buttons have names that include their visible text (WCAG 2.5.3), and decorative icons are hidden.
- One shared **polite live region** announces a widget whose updates start failing ("Weather: update failed").
- In edit mode, moving or resizing a widget with the arrow keys announces its new place through the same region ("Weather moved: column 5, row 2, 4 × 3 cells"), or that it cannot move any further. Dragging with the mouse stays silent.
  - It says this once, when the status changes, not on every failed poll.
  - The same message is not repeated within five minutes, so polling widgets cannot flood it.
- Form errors use `role="alert"`.

## Testing with a screen reader

Automated checks cannot judge reading order or whether an announcement is useful, so the manual audit ([#354](https://github.com/Waylight-App/waylight/issues/354)) needs a person with a screen reader. The pairs to cover are NVDA with Firefox on Windows, and VoiceOver with Safari on macOS or iOS.

1. Start a local install (`docker compose up`, see [getting started](getting-started.md)) or open the public demo.
2. Turn the screen reader on and use only the keyboard (or the screen reader's touch gestures on iOS).
3. Go through one flow from the checklist in #354, for example quick search: `Ctrl+K`, type a word, listen to the results, pick one with `Enter`.
4. For each problem, open an issue with the *Accessibility finding* form. Say which screen reader, browser and version you used, what you heard and what you expected. Short notes are enough.

Useful things to listen for: a button or field read without a name, a dialog that does not take the focus or does not give it back on `Esc`, a status change that is not announced, and an announcement that repeats.

When the audit is done, the result (readers, browsers, date) goes under *Known limitations* below.

## Motion

With _reduce motion_ switched on in the operating system:

- animations and transitions are shortened to nothing;
- smooth scrolling is off;
- the rotating placeholder of the search field stays still.

## Known limitations

- **Not yet audited manually with NVDA, JAWS or VoiceOver.** Automated checks find roughly a third of WCAG problems. The manual audit is tracked in #354; findings become issues of their own.
- Charts (sparklines, heatmaps, radar) show their key values as text. The radar image itself has no text alternative beyond its title.
- Custom CSS and imported themes can lower the contrast. Only the built-in themes are checked.

## Reporting a problem

Please open an issue with the label `accessibility`. Include the browser, the assistive technology and the steps to reproduce.
