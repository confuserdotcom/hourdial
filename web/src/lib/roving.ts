// Picking a mark on a chart with the keyboard: the chart is one tab stop,
// the arrows move between its marks (a roving tabindex), and Enter or
// Space does what a click does. Names read the mark out with its numbers.

import { dur, minClock } from "./format";
import { byKindRows, kindLabel } from "./kinds";

export type MoveOpts = {
  /** wrap: past the last mark comes the first (the dial's hours). */
  wrap?: boolean;
  /** cols: marks laid out in rows this long (a calendar); Up and Down move a row. */
  cols?: number;
};

/** nextIndex is the mark key moves to from mark i of n, or null when key
 *  is not a move. Right and Down step forward, Left and Up back (a row at a
 *  time when cols is set), Home and End go to the ends. Without wrap the
 *  focus stays at the ends. */
export function nextIndex(key: string, i: number, n: number, { wrap = false, cols = 0 }: MoveOpts = {}): number | null {
  if (n <= 0) return null;
  let step: number;
  switch (key) {
    case "ArrowRight": step = 1; break;
    case "ArrowLeft": step = -1; break;
    case "ArrowDown": step = cols || 1; break;
    case "ArrowUp": step = -(cols || 1); break;
    case "Home": return 0;
    case "End": return n - 1;
    default: return null;
  }
  const j = i + step;
  if (wrap) return ((j % n) + n) % n;
  if (cols && Math.abs(step) === cols && (j < 0 || j >= n)) return i;
  return Math.min(n - 1, Math.max(0, j));
}

/** activates says whether key picks the focused mark. */
export function activates(key: string): boolean {
  return key === "Enter" || key === " ";
}

function kinds(byKind: Record<string, number>): string {
  return byKindRows(byKind).map((k) => `${kindLabel(k.kind)} ${dur(k.ms)}`).join(", ");
}

/** hourName names hour h of the day: "09:00, 42 minutes: Deep 30m, …". */
export function hourName(h: number, byKind: Record<string, number>): string {
  const ms = Object.values(byKind).reduce((s, v) => s + v, 0);
  const min = Math.round(ms / 60_000);
  if (min <= 0) return `${minClock(h * 60)}, nothing tracked`;
  return `${minClock(h * 60)}, ${min} minute${min === 1 ? "" : "s"}: ${kinds(byKind)}`;
}

/** dayName names a day's bar: "Mon, 3h20: Deep 3h00, Distracting 20m". */
export function dayName(label: string, byKind: Record<string, number>): string {
  const ms = Object.values(byKind).reduce((s, v) => s + v, 0);
  if (Math.round(ms / 60_000) <= 0) return `${label}, nothing tracked`;
  return `${label}, ${dur(ms)}: ${kinds(byKind)}`;
}
