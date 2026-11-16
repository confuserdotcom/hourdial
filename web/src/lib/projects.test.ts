import { expect, test } from "vitest";
import type { Agents, Budget, Library, Top } from "./api";
import { projectRows, rangeDays } from "./projects";

const H = 3_600_000;
const top = {
  from: "2026-11-02", to: "2026-11-04", totalMS: 20 * H,
  sections: { project: [{ name: "Atlas", ms: 8 * H, kind: "", mostly: "" }, { name: "tock", ms: 4 * H, kind: "", mostly: "" }] },
  projectDays: [{ name: "Atlas", ms: [4 * H, 0, 4 * H] }, { name: "tock", ms: [H, H, 2 * H] }],
} as unknown as Top;
const lib = {
  billing: true, currency: "EUR",
  projects: [{ name: "Atlas", rate: 95, client: "Atlas Media", archived: false }, { name: "tock", rate: null, archived: false }],
} as unknown as Library;
const budgets = [{ project: "tock", budgetMS: 10 * H, budget: "10h", spentMS: 4 * H, perDayMS: H, pct: 40, over: false }] as Budget[];
const ag = { rows: [{ project: "tock", agentMS: 6 * H, youMS: H, prompts: 3, sessions: 2 }] } as unknown as Agents;

test("each project gets its share, days, money, budget and agents", () => {
  const p = projectRows(top, lib, budgets, ag);
  expect(p.rows.map((r) => [r.name, r.pct, r.share])).toEqual([["Atlas", 100, 8 / 12 * 100], ["tock", 50, 4 / 12 * 100]]);
  expect(p.rows[0]).toMatchObject({ client: "Atlas Media", rate: 95, value: 760, budget: null, agentMS: 0, days: [4 * H, 0, 4 * H] });
  expect(p.rows[1]).toMatchObject({ rate: null, value: null, agentMS: 6 * H });
  expect(p.rows[1].budget?.pct).toBe(40);
  expect(p).toMatchObject({ projectMS: 12 * H, billableMS: 8 * H, value: 760, currency: "EUR", otherMS: 8 * H });
});

test("with billing off nothing is priced", () => {
  const p = projectRows(top, { ...lib, billing: false }, null, null);
  expect(p.value).toBeNull();
  expect(p.billableMS).toBe(0);
  expect(p.rows.every((r) => r.value === null)).toBe(true);
});

test("the days of a range, in order", () => {
  expect(rangeDays("2026-10-30", "2026-11-02")).toEqual(["2026-10-30", "2026-10-31", "2026-11-01", "2026-11-02"]);
  expect(rangeDays("2026-11-02", "2026-11-02")).toEqual(["2026-11-02"]);
});
