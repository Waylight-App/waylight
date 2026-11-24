/** A point in chart units (x and y both from 0 to the chart's size). */
export type Point = readonly [number, number];

const round = (n: number) => Math.round(n * 100) / 100;

/**
 * A smooth path through the points: a Catmull-Rom curve as cubic Béziers,
 * with each control point held between its neighbours' heights, so the
 * line never swings past a measurement (a speed below zero, say).
 */
export function smoothPath(points: readonly Point[]): string {
  const [first] = points;
  if (!first) return '';
  let d = `M${round(first[0])},${round(first[1])}`;
  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[Math.max(i - 1, 0)] as Point;
    const p1 = points[i] as Point;
    const p2 = points[i + 1] as Point;
    const p3 = points[Math.min(i + 2, points.length - 1)] as Point;
    const lo = Math.min(p1[1], p2[1]);
    const hi = Math.max(p1[1], p2[1]);
    const c1y = Math.min(Math.max(p1[1] + (p2[1] - p0[1]) / 6, lo), hi);
    const c2y = Math.min(Math.max(p2[1] - (p3[1] - p1[1]) / 6, lo), hi);
    const c1x = p1[0] + (p2[0] - p0[0]) / 6;
    const c2x = p2[0] - (p3[0] - p1[0]) / 6;
    d += `C${round(c1x)},${round(c1y)} ${round(c2x)},${round(c2y)} ${round(p2[0])},${round(p2[1])}`;
  }
  return d;
}

/** The same path closed down to `baseline`, for the light fill under a line. */
export function areaPath(points: readonly Point[], baseline: number): string {
  const first = points[0];
  const last = points[points.length - 1];
  if (!first || !last) return '';
  return `${smoothPath(points)}L${round(last[0])},${baseline}L${round(first[0])},${baseline}Z`;
}

/** A "nice" top for a value axis, so the labels read 0, 50, 100 rather than 0, 47.3. */
export function niceMax(value: number): number {
  if (!(value > 0)) return 1;
  const exp = Math.pow(10, Math.floor(Math.log10(value)));
  const f = value / exp;
  const step = f <= 1 ? 1 : f <= 2 ? 2 : f <= 5 ? 5 : 10;
  return step * exp;
}
