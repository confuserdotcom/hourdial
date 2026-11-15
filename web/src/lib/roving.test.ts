import { afterEach, expect, test } from "vitest";
import { activates, dayName, hourName, nextIndex } from "./roving";
import { setClock } from "./format";

const M = 60_000;

afterEach(() => setClock(""));

test("arrows step between marks and stop at the ends", () => {
  expect(nextIndex("ArrowRight", 0, 5)).toBe(1);
  expect(nextIndex("ArrowDown", 0, 5)).toBe(1);
  expect(nextIndex("ArrowLeft", 3, 5)).toBe(2);
  expect(nextIndex("ArrowUp", 3, 5)).toBe(2);
  expect(nextIndex("ArrowLeft", 0, 5)).toBe(0);
  expect(nextIndex("ArrowRight", 4, 5)).toBe(4);
  expect(nextIndex("Home", 3, 5)).toBe(0);
  expect(nextIndex("End", 1, 5)).toBe(4);
});

test("the dial wraps round midnight", () => {
  expect(nextIndex("ArrowRight", 23, 24, { wrap: true })).toBe(0);
  expect(nextIndex("ArrowLeft", 0, 24, { wrap: true })).toBe(23);
});

test("in a calendar up and down move a week, and stay put past the edge", () => {
  expect(nextIndex("ArrowDown", 2, 30, { cols: 7 })).toBe(9);
  expect(nextIndex("ArrowUp", 9, 30, { cols: 7 })).toBe(2);
  expect(nextIndex("ArrowUp", 2, 30, { cols: 7 })).toBe(2);
  expect(nextIndex("ArrowDown", 27, 30, { cols: 7 })).toBe(27);
  expect(nextIndex("ArrowRight", 6, 30, { cols: 7 })).toBe(7);
});

test("other keys are not moves; Enter and Space pick", () => {
  expect(nextIndex("a", 1, 5)).toBeNull();
  expect(nextIndex("Tab", 1, 5)).toBeNull();
  expect(nextIndex("ArrowRight", 0, 0)).toBeNull();
  expect(activates("Enter")).toBe(true);
  expect(activates(" ")).toBe(true);
  expect(activates("Escape")).toBe(false);
});

test("an hour is named by its time and minutes, on the user's clock", () => {
  expect(hourName(9, { deep: 30 * M, shallow: 12 * M })).toBe("09:00, 42 minutes: Deep 30m, Shallow 12m");
  expect(hourName(13, {})).toBe("13:00, nothing tracked");
  expect(hourName(9, { deep: 1 * M })).toBe("09:00, 1 minute: Deep 1m");
  setClock("12h");
  expect(hourName(13, { deep: 42 * M })).toBe("1:00pm, 42 minutes: Deep 42m");
});

test("a day is named by its label and total", () => {
  expect(dayName("Mon", { deep: 3 * 60 * M, distract: 20 * M })).toBe("Mon, 3h20: Deep 3h00, Distracting 20m");
  expect(dayName("Sun", {})).toBe("Sun, nothing tracked");
});
