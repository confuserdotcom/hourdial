import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { DAY, open } from "./helpers";

// Every view, as wide as a laptop and as narrow as a small window, in both
// themes: axe-core must find nothing.
const VIEWS = [
  ["Today", `today/${DAY}`],
  ["Week", `week/${DAY}`],
  ["Trends", `trends/${DAY}`],
  ["Top", "top"],
  ["Inbox", "inbox"],
  ["Library projects", "library"],
  ["Library categories", "library/categories"],
  ["Library goals", "library/goals"],
  ["Library rules", "library/rules"],
  ["Settings", "settings"],
] as const;

for (const colorScheme of ["dark", "light"] as const) {
  test.describe(`${colorScheme} theme`, () => {
    test.use({ colorScheme });
    for (const width of [1440, 800]) {
      for (const [name, hash] of VIEWS) {
        test(`${name} at ${width}px`, async ({ page }) => {
          await page.setViewportSize({ width, height: 900 });
          await open(page, hash);
          const { violations } = await new AxeBuilder({ page }).analyze();
          const found = violations.map((v) => `${v.id} (${v.impact}): ${v.help}\n  ${v.nodes.slice(0, 5).map((n) => n.target.join(" ")).join("\n  ")}`);
          expect(found).toEqual([]);
        });
      }
    }
  });
}
