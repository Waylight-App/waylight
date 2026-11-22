/**
 * Bar heights in percent of the chart for values that may be negative: the
 * scale includes zero, so negative values hang below the base line. Returns
 * the base line position (percent from the top) too.
 */
export function barHeights(values: readonly number[]): { heights: number[]; zero: number } {
  const max = Math.max(0, ...values);
  const min = Math.min(0, ...values);
  const span = max - min || 1;
  return {
    heights: values.map((v) => (Math.abs(v) / span) * 100),
    zero: (max / span) * 100,
  };
}
