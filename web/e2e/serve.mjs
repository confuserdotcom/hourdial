// The showcase server for the browser tests: builds TestShowcase
// (internal/web, demo tag) and serves its month of made-up data from a temp
// folder, with HOME pointed at an empty folder there too, so nothing reads
// or writes the real tock data. TOCK_SHOWCASE_BIN skips the build.
import { execFileSync, spawn } from "node:child_process";
import { mkdirSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const port = process.env.TOCK_E2E_PORT ?? "7798";
const tmp = mkdtempSync(join(tmpdir(), "tock-e2e-"));
const home = join(tmp, "home");
mkdirSync(home);

let bin = process.env.TOCK_SHOWCASE_BIN;
if (!bin) {
  bin = join(tmp, "showcase");
  execFileSync("go", ["test", "-c", "-tags", "demo", "-o", bin, "./internal/web"], { cwd: root, stdio: "inherit" });
}

// TMPDIR too: the test's own t.TempDir lands in tmp, so it goes with it
// even when the server is stopped before the test can clean up.
const env = { ...process.env, HOME: home, USERPROFILE: home, TMPDIR: tmp, TEMP: tmp, TMP: tmp, TOCK_DEMO_PORT: port };
for (const k of ["XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME"]) env[k] = join(home, k.toLowerCase());

const child = spawn(bin, ["-test.run", "^TestShowcase$", "-test.timeout", "0"], { cwd: tmp, env, stdio: "inherit" });
const stop = () => child.kill("SIGTERM");
process.on("SIGINT", stop);
process.on("SIGTERM", stop);
child.on("exit", (code, signal) => {
  rmSync(tmp, { recursive: true, force: true });
  process.exit(signal ? 0 : code ?? 0);
});
