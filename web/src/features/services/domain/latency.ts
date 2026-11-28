import type { LatencyPoint } from './service';

/**
 * The response time chart of a service tile (#862). Outages are drawn apart
 * from slow answers: the line only joins answered checks, and checks without
 * an answer become bars, so a timeout never reads as a slow response.
 */

/** Key values of a response time history; latencies in ms, null without an answered check. */
export interface LatencySummary {
  latest: number | null;
  average: number | null;
  max: number | null;
  /** Number of outages: runs of consecutive checks without an answer. */
  outages: number;
  /** The latest check got no answer. */
  down: boolean;
}

/** Sums a history up for the caption and the text for screen readers. */
export function latencySummary(points: readonly LatencyPoint[]): LatencySummary {
  const answered = points.filter((p) => p.online).map((p) => Math.round(p.latencyMs));
  let outages = 0;
  points.forEach((p, i) => {
    if (!p.online && (i === 0 || points[i - 1]?.online)) outages++;
  });
  return {
    latest: answered.at(-1) ?? null,
    average: answered.length ? Math.round(answered.reduce((a, b) => a + b, 0) / answered.length) : null,
    max: answered.length ? Math.max(...answered) : null,
    outages,
    down: points.at(-1)?.online === false,
  };
}

/** The shapes of the chart in a width × height box. */
export interface LatencyShapes {
  /** Polyline point lists, one per stretch of answered checks. */
  lines: string[];
  /** Bars of the outages: x and width. */
  outages: { x: number; width: number }[];
}

const round = (n: number) => Math.round(n * 10) / 10;

/** Builds the shapes; the scale is set by the slowest answered check only. */
export function latencyShapes(points: readonly LatencyPoint[], width: number, height: number): LatencyShapes {
  const step = points.length > 1 ? width / (points.length - 1) : width;
  const top = Math.max(1, ...points.filter((p) => p.online).map((p) => p.latencyMs));
  const lines: string[] = [];
  const outages: LatencyShapes['outages'] = [];
  let line: string[] = [];
  points.forEach((p, i) => {
    const x = points.length > 1 ? i * step : width / 2;
    if (p.online) {
      line.push(`${round(x)},${round(height - (Math.max(p.latencyMs, 0) / top) * height)}`);
      return;
    }
    if (line.length) lines.push(line.join(' '));
    line = [];
    // A bar covers the half steps on both sides of its check, within the box.
    const from = Math.max(0, x - step / 2);
    outages.push({ x: round(from), width: round(Math.min(width, x + step / 2) - from) });
  });
  if (line.length) lines.push(line.join(' '));
  return { lines, outages };
}
