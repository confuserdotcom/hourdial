import { test, expect } from "vitest";
import { readFileSync } from "node:fs";
import { answer, isReal, writable, type Sample } from "./sample";

const s: Sample = {
  day: "2026-11-05",
  answers: {
    "/api/day?date=2026-11-05": { day: 1 },
    "/api/agents?date=2026-11-05": { range: "day" },
    "/api/agents?date=2026-11-05&range=week": { range: "week" },
    "/api/budgets": [1],
  },
};

test("a sampled path answers by its exact address, else by the same endpoint and range", () => {
  expect(answer(s, "/api/agents?date=2026-11-05&range=week")).toEqual({ range: "week" });
  expect(answer(s, "/api/agents?date=2026-01-01&range=week")).toEqual({ range: "week" });
  expect(answer(s, "/api/agents?date=2026-01-01")).toEqual({ range: "day" });
  expect(answer(s, "/api/day?date=1999-01-01")).toEqual({ day: 1 });
  expect(answer(s, "/api/budgets")).toEqual([1]);
  expect(answer(s, "/api/why?id=3")).toBeUndefined();
});

test("an answer is a copy, so a view can't change the sample", () => {
  const a = answer(s, "/api/budgets") as number[];
  a.push(2);
  expect(answer(s, "/api/budgets")).toEqual([1]);
});

test("setup stays real and writable; the record is read-only", () => {
  expect(isReal("/api/status")).toBe(true);
  expect(isReal("/api/settings")).toBe(true);
  expect(isReal("/api/rules?x=1")).toBe(true);
  expect(isReal("/api/day?date=2026-11-05")).toBe(false);
  expect(writable("/api/settings/hide")).toBe(true);
  expect(writable("/api/library/project")).toBe(true);
  expect(writable("/api/session")).toBe(true);
  expect(writable("/api/tag")).toBe(false);
  expect(writable("/api/invoice/issue")).toBe(false);
});

test("the shipped sample answers every screen's first load", () => {
  const shipped = JSON.parse(readFileSync(new URL("../../public/sample.json", import.meta.url), "utf8")) as Sample;
  const d = shipped.day;
  for (const p of [`/api/day?date=${d}`, `/api/week?date=${d}`, "/api/hours", `/api/month?month=${d.slice(0, 7)}`,
    `/api/agents?date=${d}`, `/api/agents?date=${d}&range=week`, `/api/agents?date=${d}&range=month`,
    `/api/top?date=${d}&range=day&series=project`, `/api/top?date=${d}&range=week&series=project`, `/api/top?date=${d}&range=month&series=project`,
    `/api/top?date=${d}&range=day`, "/api/budgets", "/api/library", "/api/inbox?days=7", "/api/labels", "/api/invoices"]) {
    expect(shipped.answers[p], p).toBeDefined();
  }
  const day = shipped.answers[`/api/day?date=${d}`] as { firstRun: boolean };
  expect(day.firstRun).toBe(false);
  expect((shipped.answers["/api/invoices"] as { invoices: unknown[] }).invoices.length).toBeGreaterThan(0);
});
