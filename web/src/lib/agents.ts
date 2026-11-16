// The agent ledger as the Agents screen draws it: bars on one scale, and
// your time priced where the project has a rate.
import type { AgentRow, Agents, AgentSpan, Library } from "./api";

export type LedgerRow = AgentRow & {
  /** agentPct and youPct are bar lengths, 0–100 of the longest bar. */
  agentPct: number; youPct: number;
  rate: number | null;
  /** value is your time in the agents' windows at the rate, or null. */
  value: number | null;
};
export type Ledger = { rows: LedgerRow[]; leverage: number; sessions: number; value: number | null; currency: string };

export function ledger(ag: Agents, lib: Library | null): Ledger {
  const top = Math.max(1, ...ag.rows.flatMap((r) => [r.agentMS, r.youMS]));
  const rates = new Map((lib?.billing ? lib.projects : []).filter((p) => p.rate != null).map((p) => [p.name, p.rate as number]));
  const rows = ag.rows.map((r) => {
    const rate = rates.get(r.project) ?? null;
    return { ...r, agentPct: (r.agentMS / top) * 100, youPct: (r.youMS / top) * 100, rate, value: rate == null ? null : (r.youMS / 3_600_000) * rate };
  });
  const priced = rows.filter((r) => r.value != null);
  return {
    rows,
    leverage: ag.youMS > 0 ? ag.agentMS / ag.youMS : 0,
    sessions: ag.rows.reduce((n, r) => n + r.sessions, 0),
    value: priced.length ? priced.reduce((n, r) => n + (r.value as number), 0) : null,
    currency: lib?.currency ?? "",
  };
}

export type Lane = { project: string; agent: [number, number][]; you: [number, number][] };
/** lanes lays a day's spans out for the "when they ran" strip: whole hours
 *  from the first span to the last, and each span as 0–100 of that width,
 *  one lane per project in the ledger's order. */
export function lanes(spans: AgentSpan[], order: string[]): { fromH: number; toH: number; lanes: Lane[]; t0: number; t1: number } | null {
  if (!spans.length) return null;
  const HOUR = 3_600_000;
  const lo = Math.min(...spans.map((s) => s.start)), hi = Math.max(...spans.map((s) => s.end));
  const d = new Date(lo);
  const day0 = new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
  const fromH = Math.floor((lo - day0) / HOUR), toH = Math.max(fromH + 1, Math.ceil((hi - day0) / HOUR));
  const t0 = day0 + fromH * HOUR, t1 = day0 + toH * HOUR;
  const at = (t: number) => ((t - t0) / (t1 - t0)) * 100;
  const names = [...order.filter((p) => spans.some((s) => s.project === p)), ...new Set(spans.map((s) => s.project).filter((p) => !order.includes(p)))];
  return {
    fromH, toH, t0, t1,
    lanes: names.map((project) => {
      const mine = spans.filter((s) => s.project === project);
      const part = (you: boolean) => mine.filter((s) => !!s.you === you).map((s) => [at(s.start), at(s.end)] as [number, number]);
      return { project, agent: part(false), you: part(true) };
    }),
  };
}
