// The agent ledger as the Agents screen draws it: bars on one scale, and
// your time priced where the project has a rate.
import type { AgentRow, Agents, Library } from "./api";

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
