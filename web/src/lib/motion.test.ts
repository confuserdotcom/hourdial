import { expect, test } from "vitest";
import { easeOut, tween } from "./motion";

test("the ease starts at 0, ends at exactly 1 and never goes back", () => {
  expect(easeOut(0)).toBe(0);
  expect(easeOut(1)).toBe(1);
  expect(easeOut(2)).toBe(1);
  let last = 0;
  for (let t = 0; t <= 1; t += 0.05) {
    expect(easeOut(t)).toBeGreaterThanOrEqual(last);
    last = easeOut(t);
  }
  expect(easeOut(0.5)).toBeGreaterThan(0.9);
});

test("a tween lands on its target", () => {
  expect(tween(10, 50, 0)).toBe(10);
  expect(tween(10, 50, 1)).toBe(50);
  expect(tween(50, 10, 1)).toBe(10);
});
