// The sample a new install shows until it has recorded anything: the
// showcase day, shipped as sample.json beside the page. While it is on,
// the record's reads come from it and nothing of it can be changed,
// exported or invoiced; setup (Rules, Settings) stays the real one.

/** Sample is sample.json: the day it holds and the API's answers on it. */
export type Sample = { day: string; answers: Record<string, unknown> };

export const SAMPLE_NOTE = "Sample data. Your own day replaces it after your first hour.";
export const SAMPLE_LOCKED = "This is sample data: nothing here can be changed, exported or invoiced.";

const REAL = ["/api/status", "/api/settings", "/api/rules"];
const WRITABLE = ["/api/settings", "/api/library", "/api/rules", "/api/session"];

const under = (path: string, roots: string[]) => {
  const p = path.split("?")[0];
  return roots.some((r) => p === r || p.startsWith(r + "/"));
};

/** isReal is whether a read goes to the server even with the sample on. */
export const isReal = (path: string) => under(path, REAL);
/** writable is whether a POST is allowed with the sample on. */
export const writable = (path: string) => under(path, WRITABLE);

/** answer is the sample's answer to a GET, as a copy: the exact address,
 *  else the same endpoint with the same range and series (the sample has
 *  one day, whatever date was asked). Undefined when it has none. */
export function answer(s: Sample, path: string): unknown {
  let hit = s.answers[path];
  if (hit === undefined) {
    const [name, query = ""] = path.split("?");
    const want = new URLSearchParams(query);
    const same = (k: string, q: URLSearchParams) => (q.get(k) ?? "") === (want.get(k) ?? "");
    for (const [k, v] of Object.entries(s.answers)) {
      const [n, q = ""] = k.split("?");
      const has = new URLSearchParams(q);
      if (n === name && same("range", has) && same("series", has)) {
        hit = v;
        break;
      }
    }
  }
  return hit === undefined ? undefined : JSON.parse(JSON.stringify(hit));
}
