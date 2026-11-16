import { expect, test } from "vitest";
import type { Agents, Library } from "./api";
import { ledger } from "./agents";

const H = 3_600_000;
const ag: Agents = {
  from: "2026-11-05", to: "2026-11-05", agentMS: 5 * H, wallMS: 4 * H, youMS: 2 * H, peak: 2,
  rows: [
    { project: "tock", agentMS: 3 * H, youMS: 0.75 * H, prompts: 13, sessions: 2 },
    { project: "Atlas", agentMS: 1.5 * H, youMS: 1.5 * H, prompts: 6, sessions: 1 },
  ],
};
const lib = (billing: boolean) => ({
  billing, currency: "EUR",
  projects: [{ name: "Atlas", rate: 95, archived: false }, { name: "tock", rate: null, archived: false }],
}) as unknown as Library;

test("bars share one scale, set by the longest of either kind", () => {
  const l = ledger(ag, null);
  expect(l.rows.map((r) => [r.agentPct, r.youPct])).toEqual([[100, 25], [50, 50]]);
  expect(l.leverage).toBe(2.5);
  expect(l.sessions).toBe(3);
  expect(l.value).toBeNull();
});

test("your time is priced only for a project with a rate, with billing on", () => {
  const on = ledger(ag, lib(true));
  expect(on.rows.map((r) => r.value)).toEqual([null, 142.5]);
  expect(on.value).toBe(142.5);
  expect(on.currency).toBe("EUR");
  expect(ledger(ag, lib(false)).value).toBeNull();
});

test("no time of yours gives no leverage, and nothing divides by zero", () => {
  const l = ledger({ ...ag, youMS: 0, rows: [] }, null);
  expect(l.leverage).toBe(0);
  expect(l.rows).toEqual([]);
});
