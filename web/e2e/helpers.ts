import { expect, type Page } from "@playwright/test";

/** TOKEN is the showcase server's fixed token (internal/web/showcase_test.go). */
export const TOKEN = "0".repeat(64);
/** DAY is the showcase's today: a Thursday, recorded up to 17:20. */
export const DAY = "2026-11-05";

/** open signs in with the token, then goes to the view in hash
 *  ("week/2026-11-05", "library/rules") and waits for it to load. */
export async function open(page: Page, hash: string) {
  await page.goto(`/#t=${TOKEN}`);
  await page.waitForFunction(() => sessionStorage.getItem("tock-token"));
  await page.goto(`/#${hash}`);
  await ready(page);
}

/** ready waits for the view's headings and for every skeleton to go. */
export async function ready(page: Page) {
  await expect(page.locator("main :is(h1, h2)").first()).toBeVisible();
  await expect(page.locator("main .animate-pulse")).toHaveCount(0);
}
