// Routing lives in the fragment once the token is gone: #agents/week,
// #projects/month/2026-10-01, #agents/2026-10-01..2026-10-20,
// #timeline/2026-11-05, #rules/goals.
import { useEffect, useState } from "react";
import { holdsToday, isoDate, type Unit } from "./format";

/** VIEWS are the views in the nav, in order: 1–6 switch to them. */
export const VIEWS = ["agents", "projects", "invoices", "timeline", "inbox", "rules"] as const;
/** ALL_VIEWS adds the ones reached another way: Settings, from the gear. */
export const ALL_VIEWS = [...VIEWS, "settings"] as const;
export type View = (typeof ALL_VIEWS)[number];
/** Range is what Agents and Projects show: the day, week or month around
 *  date, or the days from date to `to`. */
export type Range = Unit | "custom";
export const UNITS = ["day", "week", "month"] as const;
export const LIB_SECTIONS = ["projects", "categories", "goals", "rules"] as const;
export type LibSection = (typeof LIB_SECTIONS)[number];
export type Route = { view: View; date: string; to: string; unit: Range; sub: LibSection };

const DAY = /^\d{4}-\d{2}-\d{2}$/;
const SPAN = /^(\d{4}-\d{2}-\d{2})\.\.(\d{4}-\d{2}-\d{2})$/;
/** OLD are the tabs before the agent ledger led, and where each one went. */
const OLD: Record<string, (rest: string[]) => string[]> = {
  today: (r) => ["timeline", ...r],
  week: (r) => ["projects", "week", ...r],
  trends: (r) => ["projects", "month", ...r],
  top: () => ["projects"],
  library: (r) => ["rules", r[0] ?? "projects"],
};

/** ranged views share the range control; the rest ignore the unit. */
export function ranged(v: View): boolean {
  return v === "agents" || v === "projects";
}

/** unitOf is what the date arrows step by in a view, if it has a date. */
export function unitOf(r: Route): Unit | undefined {
  if (r.view === "timeline") return "day";
  return ranged(r.view) && r.unit !== "custom" ? r.unit : undefined;
}

function segments(): string[] {
  let seg = location.hash.replace(/^#/, "").split("/");
  const views = ALL_VIEWS as readonly string[];
  if (!views.includes(seg[0]) && !OLD[seg[0]]) seg = [new URLSearchParams(location.search).get("v") ?? ""];
  return OLD[seg[0]] ? OLD[seg[0]](seg.slice(1)) : seg;
}

export function parse(): Route {
  const [v, a, b] = segments();
  const today = isoDate(new Date());
  const view = (ALL_VIEWS as readonly string[]).includes(v) ? (v as View) : "agents";
  const r: Route = { view, date: today, to: today, unit: "day", sub: "rules" };
  if (view === "timeline" && DAY.test(a ?? "")) r.date = r.to = a;
  if (view === "rules" && (LIB_SECTIONS as readonly string[]).includes(a ?? "")) r.sub = a as LibSection;
  if (ranged(view)) {
    const span = SPAN.exec(a ?? "");
    if (span && span[1] <= span[2]) [r.unit, r.date, r.to] = ["custom", span[1], span[2]];
    else if (DAY.test(a ?? "")) r.date = r.to = a;
    else if ((UNITS as readonly string[]).includes(a)) {
      r.unit = a as Unit;
      if (DAY.test(b ?? "")) r.date = r.to = b;
    }
  }
  return r;
}

/** format is the hash of a route, without what its view does not use. */
export function format(r: Route): string {
  if (r.view === "rules") return r.sub === "rules" ? "rules" : `rules/${r.sub}`;
  if (r.view === "timeline") return r.date === isoDate(new Date()) ? "timeline" : `timeline/${r.date}`;
  if (!ranged(r.view)) return r.view;
  if (r.unit === "custom") return `${r.view}/${r.date}..${r.to}`;
  const now = holdsToday(r.date, r.unit);
  if (r.unit === "day") return now ? r.view : `${r.view}/${r.date}`;
  return now ? `${r.view}/${r.unit}` : `${r.view}/${r.unit}/${r.date}`;
}

/** canonical is the hash an old route should be rewritten to, or null. */
export function canonical(): string | null {
  return OLD[location.hash.replace(/^#/, "").split("/")[0]] ? format(parse()) : null;
}

/** rangeQuery is a route's range as /api/agents and its like take it. */
export function rangeQuery(r: Pick<Route, "date" | "to" | "unit">): string {
  if (r.unit === "custom") return `from=${r.date}&to=${r.to}`;
  return r.unit === "day" ? `date=${r.date}` : `date=${r.date}&range=${r.unit}`;
}

export function go(r: Partial<Route>) {
  const cur = parse();
  const next = { ...cur, ...r };
  // A view of its own section when none was asked for, and one day unless
  // the range was set with it.
  if (r.view === "rules" && !r.sub) next.sub = "rules";
  if (r.date && !r.to) next.to = r.date;
  if (r.date && !r.unit && next.unit === "custom") next.unit = "day";
  if (next.view === "timeline" && cur.unit !== "day" && !r.date) next.date = isoDate(new Date());
  location.hash = format(next);
}

export function useRoute(): Route {
  const [r, setR] = useState(parse);
  useEffect(() => {
    const on = () => {
      const c = canonical();
      if (c) history.replaceState(null, "", `#${c}`);
      setR(parse());
    };
    on();
    window.addEventListener("hashchange", on);
    return () => window.removeEventListener("hashchange", on);
  }, []);
  return r;
}
