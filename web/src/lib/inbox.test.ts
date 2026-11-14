import { test, expect } from "vitest";
import { guessText, topMS } from "./inbox";

test("topMS sums the n biggest groups, whatever the order", () => {
  const it = (ms: number) => ({ key: String(ms), app: "", ms, spans: 1, example: "", private: false });
  expect(topMS([it(1), it(9), it(4), it(7)], 2)).toBe(16);
  expect(topMS([it(3)], 5)).toBe(3);
  expect(topMS([], 5)).toBe(0);
});

test("guessText says the label, the reason and a split vote", () => {
  expect(guessText({ label: "Study", reason: "same site", share: 0.8 })).toBe("Study: same site, 80 % of the evidence");
  expect(guessText({ label: "Code", reason: "same app", share: 1 })).toBe("Code: same app");
  expect(guessText({ label: "Code" })).toBe("");
  expect(guessText(undefined)).toBe("");
});
