import { defineConfig, devices } from "@playwright/test";

// Browser tests of the dashboard against the showcase server (e2e/serve.mjs):
// axe on every view, and a few real flows. npm run e2e; npm test does not
// run them (vitest only reads src/).
const port = process.env.TOCK_E2E_PORT ?? "7798";

export default defineConfig({
  testDir: "./e2e",
  // One server with one database: the flows change it, so one at a time.
  workers: 1,
  fullyParallel: false,
  retries: 0,
  forbidOnly: !!process.env.CI,
  reporter: process.env.CI ? [["list"], ["github"]] : "list",
  use: {
    ...devices["Desktop Chrome"],
    // TOCK_E2E_CHANNEL=chrome runs in the installed Chrome, with no browser
    // download (npx playwright install).
    channel: process.env.TOCK_E2E_CHANNEL || undefined,
    baseURL: `http://127.0.0.1:${port}`,
    // The showcase day is 5 Nov 2026 in UTC+2 (apptest.Zone).
    timezoneId: "Etc/GMT-2",
    locale: "en-GB",
    reducedMotion: "reduce",
    trace: "retain-on-failure",
  },
  webServer: {
    command: "node e2e/serve.mjs",
    url: `http://127.0.0.1:${port}/`,
    timeout: 300_000,
    reuseExistingServer: !process.env.CI,
    stdout: "pipe",
    // SIGTERM so e2e/serve.mjs can stop the server and remove its temp folder.
    gracefulShutdown: { signal: "SIGTERM", timeout: 10_000 },
  },
});
