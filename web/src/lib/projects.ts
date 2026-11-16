// The Projects screen's table: each project's time in the range with what
// the library, the budgets and the agent ledger know about it.
import type { Agents, Budget, Library, Top } from "./api";
import { addDays, isoDate, num, parseDate } from "./format";

export type ProjectRow = {
  name: string; ms: number;
  /** pct is the bar: 0–100 of the longest project; share is of all project time. */
  pct: number; share: number;
  /** days is its time on each day of the range, when the series was asked for. */
  days: number[];
  client: string; rate: number | null;
  /** value is its time at its rate, or null without one. */
  value: number | null;
  budget: Budget | null;
  agentMS: number; youMS: number;
};
export type Projects = {
  rows: ProjectRow[];
  projectMS: number; billableMS: number; otherMS: number;
  value: number | null; currency: string;
};

export function projectRows(top: Top, lib: Library | null, budgets: Budget[] | null, ag: Agents | null): Projects {
  const list = top.sections.project ?? [];
  const longest = Math.max(1, ...list.map((r) => r.ms));
  const projectMS = list.reduce((n, r) => n + r.ms, 0);
  const rows = list.map((r) => {
    const p = lib?.projects.find((x) => x.name === r.name);
    const rate = lib?.billing && p?.rate != null ? p.rate : null;
    const a = ag?.rows.find((x) => x.project === r.name);
    return {
      name: r.name, ms: r.ms, pct: (r.ms / longest) * 100, share: projectMS ? (r.ms / projectMS) * 100 : 0,
      days: top.projectDays?.find((d) => d.name === r.name)?.ms ?? [],
      client: p?.client ?? "", rate, value: rate == null ? null : (r.ms / 3_600_000) * rate,
      budget: budgets?.find((b) => b.project === r.name) ?? null,
      agentMS: a?.agentMS ?? 0, youMS: a?.youMS ?? 0,
    };
  });
  const priced = rows.filter((r) => r.value != null);
  return {
    rows, projectMS,
    billableMS: priced.reduce((n, r) => n + r.ms, 0),
    otherMS: Math.max(0, top.totalMS - projectMS),
    value: priced.length ? priced.reduce((n, r) => n + (r.value as number), 0) : null,
    currency: lib?.currency ?? "",
  };
}

/** rangeDays are the days from from to to, both included. */
export function rangeDays(from: string, to: string): string[] {
  const out: string[] = [];
  for (let d = parseDate(from); isoDate(d) <= to && out.length < 400; d = addDays(d, 1)) out.push(isoDate(d));
  return out;
}

/** money is an amount as invoices write it: "1,234.50 EUR". */
export function money(v: number, currency: string): string {
  return `${num(v, 2)} ${currency}`.trim();
}
