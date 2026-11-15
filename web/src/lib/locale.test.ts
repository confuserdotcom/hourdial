import { afterEach, expect, test } from "vitest";
import { dayTitle, dayTitleLong, dayTitleYear, num, periodTitle, parseDate, setDateStyle, setNumbers } from "./format";

afterEach(() => {
  setDateStyle(undefined);
  setNumbers(undefined);
});

const thu = parseDate("2026-11-05");
const today = new Date(2026, 11, 20);

test("dates default to day first, as before the setting", () => {
  expect(dayTitle(thu)).toBe("Thu 5 Nov");
  expect(dayTitleYear(thu)).toBe("Thu 5 Nov 2026");
  expect(dayTitleLong(thu)).toBe("Thursday 5 November");
  expect(periodTitle("2026-11-05", "week", today)).toBe("2–8 Nov");
  expect(periodTitle("2026-10-28", "week", today)).toBe("26 Oct – 1 Nov");
});

test("an unknown date style is day first", () => {
  setDateStyle("ymd");
  expect(dayTitle(thu)).toBe("Thu 5 Nov");
});

test("mdy writes the month first", () => {
  setDateStyle("mdy");
  expect(dayTitle(thu)).toBe("Thu Nov 5");
  expect(dayTitleYear(thu)).toBe("Thu Nov 5, 2026");
  expect(dayTitleLong(thu)).toBe("Thursday, November 5");
  expect(periodTitle("2026-11-05", "day", today)).toBe("Thu Nov 5");
  expect(periodTitle("2026-11-05", "week", today)).toBe("Nov 2–8");
  expect(periodTitle("2026-10-28", "week", today)).toBe("Oct 26 – Nov 1");
});

test("iso writes year-month-day", () => {
  setDateStyle("iso");
  expect(dayTitle(thu)).toBe("Thu 2026-11-05");
  expect(dayTitleYear(thu)).toBe("Thu 2026-11-05");
  expect(dayTitleLong(thu)).toBe("Thursday 2026-11-05");
  expect(periodTitle("2026-11-05", "week", today)).toBe("2026-11-02 – 2026-11-08");
});

test("numbers default to a point and no grouping", () => {
  expect(num(0.7, 1)).toBe("0.7");
  expect(num(1234.5, 2)).toBe("1234.50");
  expect(num(-1234.5, 1)).toBe("-1234.5");
});

test("each numbers style, as [display] numbers names it", () => {
  const cases: [string, string, string][] = [
    ["1234.5", "1234567.50", "0.7"],
    ["1,234.5", "1,234,567.50", "0.7"],
    ["1.234,5", "1.234.567,50", "0,7"],
    ["1 234,5", "1 234 567,50", "0,7"],
    ["1234,5", "1234567,50", "0,7"],
  ];
  for (const [style, big, small] of cases) {
    setNumbers(style);
    expect(num(1234567.5, 2)).toBe(big);
    expect(num(0.66, 1)).toBe(small);
    expect(num(999, 0)).toBe("999");
  }
  setNumbers("de");
  expect(num(1234.5, 1)).toBe("1234.5");
});
