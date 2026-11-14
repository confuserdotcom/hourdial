// Pure helpers for the Inbox.
import type { Item } from "./api";

/** topMS is the time the n biggest groups hold: what labelling them files. */
export function topMS(items: Item[], n: number): number {
  return items.map((i) => i.ms).sort((a, b) => b - a).slice(0, n).reduce((s, ms) => s + ms, 0);
}

/** guessText says why the inbox guesses a label, inline under the item:
 *  "Study: same site, 80 % of the evidence". "" without a reason. */
export function guessText(s: Item["suggestion"]): string {
  if (!s?.reason) return "";
  const share = s.share && s.share < 1 ? `, ${Math.round(s.share * 100)} % of the evidence` : "";
  return `${s.label}: ${s.reason}${share}`;
}
