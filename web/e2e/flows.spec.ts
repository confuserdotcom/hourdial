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
  await open(page, "rules");
  await page.getByRole("button", { name: "Add", exact: true }).click();
  const sheet = page.getByRole("dialog", { name: "Add rule" });
  await sheet.getByLabel("Title").fill("^Quarterly report");
  await sheet.getByLabel("Files it as").fill("Writing");
  await sheet.getByRole("button", { name: "Add", exact: true }).click();
  await expect(sheet).toBeHidden();
  await expect(page.locator("main").getByText("^Quarterly report")).toBeVisible();
});

test("the date control steps to the day before", async ({ page }) => {
  await open(page, `timeline/${DAY}`);
  await page.getByRole("button", { name: /^Previous day/ }).click();
  await expect(page).toHaveURL(/#timeline\/2026-11-04$/);
  await ready(page);
  await expect(page.locator('button[title="Pick a date"]')).toContainText("4 Nov");
});

test("the focus score opens its explanation", async ({ page }) => {
  await open(page, `timeline/${DAY}`);
  await page.getByRole("button", { name: "How it's scored" }).click();
  const why = page.getByRole("dialog", { name: /^Focus score \d+$/ });
  await expect(why).toBeVisible();
  await expect(why.getByRole("heading", { name: /^Focus score \d+$/ })).toBeVisible();
});

test("an hour picked on the minutes-per-hour chart with the keyboard zooms the timeline", async ({ page }) => {
  await open(page, `timeline/${DAY}`);
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
  await open(page, `projects/week/${DAY}`);
  const days = page.getByRole("group", { name: /^Hours per day/ });
  await days.locator('[tabindex="0"]').focus();
  await page.keyboard.press("ArrowRight");
  await expect(days.locator(":focus")).toHaveAccessibleName(/^Tue 3 Nov, \d+h\d\d: /);
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(/#timeline\/2026-11-03$/);
});

test("an invoice for a rated project previews, issues after a confirmation, and joins the list", async ({ page }) => {
  await open(page, `projects/week/${DAY}`);
  await page.getByRole("button", { name: "Export" }).click();
  await page.getByRole("button", { name: /^Invoice…/ }).click();
  await expect(page).toHaveURL(/#invoices\/2026-11-02\.\.2026-11-08$/);
  const draft = page.getByRole("region", { name: "Draft" });
  await expect(draft.getByLabel("Project")).toHaveValue(/.+/);
  await expect(draft.getByLabel("First day")).toHaveValue("2026-11-02");
  const preview = draft.locator("[data-preview] span");
  await expect(preview).toHaveText(/\d[\d,.]*\.\d\d EUR/);
  const total = (await preview.innerText()).trim();
  await draft.getByRole("button", { name: "Issue…" }).click();
  const confirm = page.getByRole("dialog", { name: "Issue this invoice?" });
  await expect(confirm).toContainText("permanent and numbered");
  await confirm.getByRole("button", { name: "Issue it" }).click();
  await expect(confirm).toBeHidden();
  // The toast names the number the invoice got and the total it was issued for.
  const said = page.getByText(/^Issued \S+ · /);
  await expect(said).toContainText(total);
  const number = /^Issued (\S+) · /.exec(await said.innerText())![1];
  const row = page.getByRole("region", { name: "Issued" }).locator("li").filter({ hasText: number });
  await expect(row).toContainText("issued");
  await expect(row).toContainText(total);
  const dl = page.waitForEvent("download");
  await row.getByRole("button", { name: `Save the PDF of ${number}` }).click();
  expect((await dl).suggestedFilename()).toBe(`${number}.pdf`);
});

test("the old tabs' addresses open the tabs that took them over", async ({ page }) => {
  await open(page, `week/${DAY}`);
  await expect(page).toHaveURL(new RegExp(`#projects/week/${DAY}$`));
  await expect(page.getByRole("heading", { level: 1, name: "Projects" })).toBeVisible();
  await page.goto(`/#today/${DAY}`);
  await expect(page).toHaveURL(new RegExp(`#timeline/${DAY}$`));
  await page.goto("/#library/goals");
  await expect(page).toHaveURL(/#rules\/goals$/);
});

test("the sidebar leads with agents, and the range carries over to projects", async ({ page }) => {
  await open(page, "");
  await expect(page.getByRole("heading", { level: 1, name: "Agents" })).toBeVisible();
  await page.getByRole("radio", { name: "Week" }).click();
  await expect(page).toHaveURL(/#agents\/week/);
  await page.getByRole("navigation", { name: "Views" }).first().getByRole("button", { name: /^Projects/ }).click();
  await expect(page).toHaveURL(/#projects\/week/);
});

test("the agent ledger shows leverage and your time, and a project opens Projects", async ({ page }) => {
  await open(page, `agents/${DAY}`);
  const rows = page.locator("main tbody tr");
  await expect(rows).toHaveCount(3);
  await expect(page.getByRole("region", { name: "Totals" })).toContainText(/Leverage\d+[.,]\d×/);
  await expect(rows.filter({ hasText: "tock" }).getByRole("img", { name: /^You \d+h\d\d$/ })).toBeVisible();
  await rows.filter({ hasText: "tock" }).getByRole("button").click();
  await expect(page).toHaveURL(new RegExp(`#projects/${DAY}/@tock$`));
  await expect(page.locator("main button[aria-controls]").filter({ hasText: "tock" })).toHaveAttribute("aria-expanded", "true");
});

test("a custom range shows on its button and in the address", async ({ page }) => {
  await open(page, `agents/${DAY}`);
  await page.getByRole("button", { name: "Custom" }).click();
  await page.getByLabel("From", { exact: true }).fill("2026-10-20");
  await page.getByLabel("To", { exact: true }).fill("2026-11-05");
  await page.getByRole("button", { name: "Show" }).click();
  await expect(page).toHaveURL(/#agents\/2026-10-20\.\.2026-11-05$/);
  await expect(page.getByRole("button", { name: /20 Oct – .*5 Nov/ })).toBeVisible();
});

test("a project opens to its days and hands its range to an invoice draft", async ({ page }) => {
  await open(page, `projects/week/${DAY}`);
  const row = page.locator("main button[aria-controls]").filter({ hasText: "Atlas" });
  await expect(row).toContainText(/\d[\d,.]*\.\d\d EUR/);
  await row.click();
  await expect(page).toHaveURL(/#projects\/week\/2026-11-05\/@Atlas$/);
  await expect(page.getByRole("img", { name: /^Atlas per day: Mon 2 Nov / })).toBeVisible();
  await page.getByRole("button", { name: "Draft invoice" }).click();
  await expect(page).toHaveURL(/#invoices\/2026-11-02\.\.2026-11-05\/@Atlas$/);
});
