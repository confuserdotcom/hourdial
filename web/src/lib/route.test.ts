import { test, expect, vi, afterEach } from "vitest";

afterEach(() => (vi.unstubAllGlobals(), vi.useRealTimers()));

async function at(hash: string, search = "") {
  vi.resetModules();
  vi.useFakeTimers({ now: new Date(2026, 10, 5, 12) });
  vi.stubGlobal("location", { hash, search });
  return await import("./route");
}
const parseAt = async (hash: string, search = "") => (await at(hash, search)).parse();

test("agents is where an empty or unknown hash lands", async () => {
  expect(await parseAt("")).toMatchObject({ view: "agents", unit: "day", date: "2026-11-05" });
  expect(await parseAt("#nonsense")).toMatchObject({ view: "agents" });
  expect(await parseAt("", "?v=inbox")).toMatchObject({ view: "inbox" });
  expect(await parseAt("", "?v=today")).toMatchObject({ view: "timeline" });
});

test("agents and projects carry a range: a unit and a day in it", async () => {
  expect(await parseAt("#agents/week")).toMatchObject({ view: "agents", unit: "week", date: "2026-11-05" });
  expect(await parseAt("#projects/month/2026-10-01")).toMatchObject({ view: "projects", unit: "month", date: "2026-10-01" });
  expect(await parseAt("#agents/2026-11-03")).toMatchObject({ view: "agents", unit: "day", date: "2026-11-03" });
  expect(await parseAt("#agents/2026-10-01..2026-10-20")).toMatchObject({ view: "agents", unit: "custom", date: "2026-10-01", to: "2026-10-20" });
  expect(await parseAt("#agents/2026-10-20..2026-10-01")).toMatchObject({ unit: "day", date: "2026-11-05" });
  expect(await parseAt("#timeline/2026-11-03")).toMatchObject({ view: "timeline", date: "2026-11-03" });
});

test("the old routes land on the tab that took them over", async () => {
  expect(await parseAt("#today/2026-11-03")).toMatchObject({ view: "timeline", date: "2026-11-03" });
  expect(await parseAt("#week/2026-11-02")).toMatchObject({ view: "projects", unit: "week", date: "2026-11-02" });
  expect(await parseAt("#trends/2026-10-01")).toMatchObject({ view: "projects", unit: "month", date: "2026-10-01" });
  expect(await parseAt("#top")).toMatchObject({ view: "projects", unit: "day" });
  expect(await parseAt("#library")).toMatchObject({ view: "rules", sub: "projects" });
  expect(await parseAt("#library/goals")).toMatchObject({ view: "rules", sub: "goals" });
});

test("rules sections ride in the second segment", async () => {
  expect(await parseAt("#rules")).toMatchObject({ view: "rules", sub: "rules" });
  expect(await parseAt("#rules/goals")).toMatchObject({ view: "rules", sub: "goals" });
  expect(await parseAt("#rules/nonsense")).toMatchObject({ view: "rules", sub: "rules" });
  expect(await parseAt("#settings")).toMatchObject({ view: "settings" });
});

test("a route writes the shortest hash that reads back the same", async () => {
  const { format, parse } = await at("");
  const base = parse();
  expect(format(base)).toBe("agents");
  expect(format({ ...base, unit: "week" })).toBe("agents/week");
  expect(format({ ...base, unit: "week", date: "2026-10-20" })).toBe("agents/week/2026-10-20");
  expect(format({ ...base, view: "projects", unit: "month", date: "2026-10-01" })).toBe("projects/month/2026-10-01");
  expect(format({ ...base, unit: "custom", date: "2026-10-01", to: "2026-10-20" })).toBe("agents/2026-10-01..2026-10-20");
  expect(format({ ...base, view: "timeline", date: "2026-11-03" })).toBe("timeline/2026-11-03");
  expect(format({ ...base, view: "timeline" })).toBe("timeline");
  expect(format({ ...base, view: "rules", sub: "rules" })).toBe("rules");
  expect(format({ ...base, view: "rules", sub: "goals" })).toBe("rules/goals");
  expect(format({ ...base, view: "inbox", unit: "week" })).toBe("inbox");
});

test("an old hash is rewritten, a current one is left alone", async () => {
  expect((await at("#week/2026-10-26")).canonical()).toBe("projects/week/2026-10-26");
  expect((await at("#library/rules")).canonical()).toBe("rules");
  expect((await at("#agents/week")).canonical()).toBeNull();
  expect((await at("#t=abc")).canonical()).toBeNull();
});

test("the range of a route, as the API takes it", async () => {
  const { rangeQuery, parse } = await at("");
  const base = parse();
  expect(rangeQuery(base)).toBe("date=2026-11-05");
  expect(rangeQuery({ ...base, unit: "week" })).toBe("date=2026-11-05&range=week");
  expect(rangeQuery({ ...base, unit: "month", date: "2026-10-01" })).toBe("date=2026-10-01&range=month");
  expect(rangeQuery({ ...base, unit: "custom", date: "2026-10-01", to: "2026-10-20" })).toBe("from=2026-10-01&to=2026-10-20");
});
