import { expect, test } from "@playwright/test";
import { DAY, open, ready } from "./helpers";

// Real flows against the showcase server. They change its data, so they
// run one at a time (playwright.config.ts) on their own items.

test("labelling an inbox item takes it out of the inbox", async ({ page }) => {
  await open(page, "inbox");
  const items = page.locator("main li");
  const item = items.first();
  const key = (await item.locator(".font-medium").first().innerText()).trim();
  await item.getByRole("button", { name: "Label…", exact: true }).click();
  const sheet = page.getByRole("dialog", { name: key });
  await sheet.getByRole("option", { name: "Research", exact: true }).click();
  await expect(sheet).toBeHidden();
  await expect(items.filter({ hasText: key })).toHaveCount(0);
});

test("a rule added in the rule sheet shows in the rules", async ({ page }) => {
  await open(page, "library/rules");
  await page.getByRole("button", { name: "Add", exact: true }).click();
  const sheet = page.getByRole("dialog", { name: "Add rule" });
  await sheet.getByLabel("Title").fill("^Quarterly report");
  await sheet.getByLabel("Files it as").fill("Writing");
  await sheet.getByRole("button", { name: "Add", exact: true }).click();
  await expect(sheet).toBeHidden();
  await expect(page.locator("main").getByText("^Quarterly report")).toBeVisible();
});

test("the date control steps to the day before", async ({ page }) => {
  await open(page, `today/${DAY}`);
  await page.getByRole("button", { name: /^Previous day/ }).click();
  await expect(page).toHaveURL(/#today\/2026-11-04$/);
  await ready(page);
  await expect(page.locator('button[title="Pick a date"]')).toContainText("4 Nov");
});

test("the focus score opens its explanation", async ({ page }) => {
  await open(page, `today/${DAY}`);
  await page.getByRole("button", { name: "How it's scored" }).click();
  const why = page.getByRole("dialog", { name: /^Focus score \d+$/ });
  await expect(why).toBeVisible();
  await expect(why.getByRole("heading", { name: /^Focus score \d+$/ })).toBeVisible();
});

test("an hour picked on the minutes-per-hour chart with the keyboard zooms the timeline", async ({ page }) => {
  await open(page, `today/${DAY}`);
  const whole = page.getByRole("button", { name: "Show the whole day (0)" });
  await expect(whole).toBeDisabled();
  const hours = page.getByRole("group", { name: /^Minutes per hour/ });
  await hours.locator('[tabindex="0"]').focus();
  await page.keyboard.press("Home");
  await page.keyboard.press("ArrowRight");
  const focused = hours.locator(":focus");
  await expect(focused).toHaveAccessibleName(/^09:00, \d+ minutes: /);
  await page.keyboard.press("Enter");
  await expect(whole).toBeEnabled();
});

test("a day picked on the week chart with the keyboard opens it", async ({ page }) => {
  await open(page, `week/${DAY}`);
  const days = page.getByRole("group", { name: /^Hours per day/ });
  await days.locator('[tabindex="0"]').focus();
  await page.keyboard.press("ArrowRight");
  await expect(days.locator(":focus")).toHaveAccessibleName(/^Tue 3 Nov, \d+h\d\d: /);
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(/#today\/2026-11-03$/);
});

test("an invoice for a rated project previews and issues with its number and total", async ({ page }) => {
  await open(page, `week/${DAY}`);
  await page.getByRole("button", { name: "Export" }).click();
  await page.getByRole("button", { name: /^Invoice…/ }).click();
  const dialog = page.getByRole("dialog", { name: "Invoice" });
  await expect(dialog.getByLabel("Project")).toHaveValue(/.+/);
  const preview = dialog.locator("p.num");
  await expect(preview).toContainText(/\d[\d,.]*\.\d\d EUR/);
  const total = /(\d[\d,.]*\.\d\d EUR)/.exec(await preview.innerText())![1];
  await dialog.getByRole("button", { name: "Issue", exact: true }).click();
  await expect(dialog).toBeHidden();
  // The toast names the number the invoice got and the total it was issued for.
  const issued = page.getByText(/^Issued \S+ · /);
  await expect(issued).toHaveText(new RegExp(`^Issued [A-Za-z0-9-]*\\d[A-Za-z0-9-]* · ${total.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}$`));});
