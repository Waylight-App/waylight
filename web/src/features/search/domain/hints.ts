/**
 * Rotating placeholder of the search pill (#189): questions and ideas that
 * show what the search can do. Since #479 the examples come from the same
 * data as the help ("?"), so a hint never names a module that is off. Pure
 * rules; the timer lives in the UI.
 */

/** How long one hint is shown. */
export const HINT_INTERVAL_MS = 4000;

/** Most example hints in one rotation, so the pointer to "?" comes round. */
export const MAX_EXAMPLE_HINTS = 8;

/**
 * The hint to show out of `count`: rotation pauses while the field has
 * focus (the current hint stays) and is off for reduced motion (always the
 * first hint).
 */
export function nextHint(
  current: number,
  count: number,
  opts: { focused: boolean; reducedMotion: boolean },
): number {
  if (opts.reducedMotion || count === 0) return 0;
  if (opts.focused) return current;
  return (current + 1) % count;
}
