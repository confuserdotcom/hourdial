import { useRef, useState, type ReactNode } from "react";
import { activates, nextIndex, type MoveOpts } from "@/lib/roving";

/** useRoving makes n marks one tab stop: the arrows move the focus between
 *  them (see nextIndex), Enter or Space picks one. shown is the focused
 *  mark, for a readout; start is the mark Tab lands on first. */
export function useRoving(n: number, onPick: (i: number) => void, opts: MoveOpts & { start?: number } = {}) {
  const [cur, setCur] = useState<number | null>(null);
  const [shown, setShown] = useState<number | null>(null);
  const refs = useRef<(HTMLElement | SVGElement | null)[]>([]);
  const stop = Math.min(n - 1, Math.max(0, cur ?? opts.start ?? 0));
  const item = (i: number) => ({
    ref: (el: HTMLElement | SVGElement | null) => { refs.current[i] = el; },
    tabIndex: i === stop ? 0 : -1,
    onFocus: () => { setCur(i); setShown(i); },
    onBlur: () => setShown((s) => (s === i ? null : s)),
    onKeyDown: (e: React.KeyboardEvent) => {
      const j = nextIndex(e.key, i, n, opts);
      if (j !== null) {
        e.preventDefault();
        refs.current[j]?.focus();
      } else if (activates(e.key) && !(e.currentTarget instanceof HTMLButtonElement)) {
        // a button clicks on Enter and Space by itself
        e.preventDefault();
        onPick(i);
      }
    },
  });
  return { item, shown };
}

/** Marks lays a row of focusable bands over a recharts plot, one per bar
 *  (or per point, when points is set), for the keyboard: the chart itself
 *  stays for the mouse, so the bands take no pointer events. inset is the
 *  plot's place in the chart, as CSS lengths. readout is what a focused
 *  band shows, as the mouse tooltip would. */
export function Marks({ label, names, onPick, inset, points, readout, start }: {
  label: string;
  names: string[];
  onPick: (i: number) => void;
  inset: { left: string; right: string; top: string; bottom: string };
  points?: boolean;
  readout?: (i: number) => ReactNode;
  start?: number;
}) {
  const n = names.length;
  const { item, shown } = useRoving(n, onPick, { start });
  if (!n) return null;
  const at = (i: number) =>
    points
      ? { left: `${n > 1 ? (i / (n - 1)) * 100 : 50}%`, width: `${100 / Math.max(1, n - 1)}%`, transform: "translateX(-50%)" }
      : { left: `${(i / n) * 100}%`, width: `${100 / n}%` };
  return (
    <div role="group" aria-label={label} className="pointer-events-none absolute" style={inset}>
      {names.map((name, i) => (
        <button key={i} type="button" aria-label={name} {...item(i)} onClick={() => onPick(i)}
          className="absolute inset-y-0 rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          style={at(i)}>
          {shown === i && readout && (
            <span className={`absolute top-0 z-10 -translate-y-[calc(100%+4px)] whitespace-nowrap rounded-md border bg-popover px-2.5 py-1.5 text-left text-xs text-popover-foreground shadow ${i < n / 2 ? "left-0" : "right-0"}`}>
              {readout(i)}
            </span>
          )}
        </button>
      ))}
    </div>
  );
}

/** plotInset is where a chart with margin left -24 and the default axes draws
 *  its plot: a 60px y axis less the margin on the left, a 30px x axis
 *  below. */
export function plotInset(top: number, right: number) {
  return { left: "36px", right: `${right}px`, top: `${top}px`, bottom: "30px" };
}
