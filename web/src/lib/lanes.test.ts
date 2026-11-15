import { expect, test } from "vitest";
import { labelWidth, legendWidth, LABEL_MIN_REM } from "./lanes";

const rem = (w: string) => Number(/, ([\d.]+)rem,/.exec(w)?.[1]);

test("the lane names get room for the longest one", () => {
  const short = labelWidth(["Code", "Away"]);
  const long = labelWidth(["Code", "Communication"]);
  expect(rem(long)).toBeGreaterThan(rem(short));
  // "Communication" and a duration fit without an ellipsis: about 13 chars
  // of 13px Inter and "3h05" beside them.
  expect(rem(long)).toBeGreaterThanOrEqual(10.5);
});

test("the column never shrinks below the old width, nor takes the track", () => {
  expect(labelWidth([])).toBe(`clamp(${LABEL_MIN_REM}rem, ${LABEL_MIN_REM}rem, max(${LABEL_MIN_REM}rem, 30%))`);
  expect(labelWidth(["A very long category name that would eat the timeline"])).toMatch(/max\(9rem, 30%\)\)$/);
});

test("a donut legend gets room for its longest name", () => {
  const remOf = (w: string) => Number(/^([\d.]+)rem$/.exec(w)?.[1]);
  // "Northwind" (9 chars), its duration and its share fit without an ellipsis.
  expect(remOf(legendWidth(["Northwind", "Acme"]))).toBeGreaterThanOrEqual(11.5);
  expect(remOf(legendWidth(["Northwind Traders"]))).toBeGreaterThan(remOf(legendWidth(["Northwind"])));
  expect(remOf(legendWidth([]))).toBeGreaterThan(0);
});
