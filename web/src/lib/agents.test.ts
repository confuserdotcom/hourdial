import { expect, test } from "vitest";
import type { Agents, Library } from "./api";
import { lanes, ledger } from "./agents";

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

test("a day's spans become lanes over whole hours, in the ledger's order", () => {
  const at = (h: number, m = 0) => new Date(2026, 10, 5, h, m).getTime();
  const l = lanes([
    { project: "Atlas", start: at(10), end: at(11) },
    { project: "tock", start: at(9, 30), end: at(10, 30) },
    { project: "tock", start: at(12), end: at(12, 45), you: true },
  ], ["tock", "Atlas"])!;
  expect([l.fromH, l.toH]).toEqual([9, 13]);
  expect(l.lanes.map((x) => x.project)).toEqual(["tock", "Atlas"]);
  expect(l.lanes[0].agent).toEqual([[12.5, 37.5]]);
  expect(l.lanes[0].you).toEqual([[75, 93.75]]);
  expect(l.lanes[1].agent).toEqual([[25, 50]]);
  expect(lanes([], [])).toBeNull();
});
