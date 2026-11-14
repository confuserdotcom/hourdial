import { expect, test } from "vitest";
import { firstDay } from "./firstday";

const M = 60_000;
const apps = [{ name: "Code", ms: 6 * M }, { name: "Safari", ms: 4 * M }, { name: "Slack", ms: 2 * M }, { name: "Mail", ms: M }];

test("a new install's day says what was recorded and what is labelled", () => {
  const f = firstDay({ today: true, usual: null, trackedMS: 13 * M, byKind: { deep: 6 * M, shallow: 3 * M, "": 4 * M }, apps });
  expect(f).toEqual({
    recorded: "13m", top: ["Code", "Safari", "Slack"], labelled: "9m", unlabelled: "4m", unlabelledMS: 4 * M,
  });
});

test("nothing to say once there is a usual day, on a past day, or with nothing tracked", () => {
  const base = { today: true, usual: null, trackedMS: 13 * M, byKind: { "": 13 * M }, apps };
  expect(firstDay({ ...base, usual: { days: 3 } })).toBeNull();
  expect(firstDay({ ...base, today: false })).toBeNull();
  expect(firstDay({ ...base, trackedMS: 0 })).toBeNull();
});

test("all labelled leaves nothing for the inbox", () => {
  const f = firstDay({ today: true, usual: null, trackedMS: 5 * M, byKind: { deep: 5 * M }, apps: apps.slice(0, 1) });
  expect(f).toMatchObject({ labelled: "5m", unlabelledMS: 0, top: ["Code"] });
});
