// First day: what a fresh install has recorded so far, in words, for the
// Today view before there is a usual day to compare with. Pure, so tested.
import { dur } from "./format";

export type FirstDay = { recorded: string; top: string[]; labelled: string; unlabelled: string; unlabelledMS: number };

/** firstDay sums up today until tock has a usual day; null otherwise. */
export function firstDay({ today, usual, trackedMS, byKind, apps }: {
  today: boolean; usual: unknown; trackedMS: number; byKind: Record<string, number>; apps: { name: string; ms: number }[];
}): FirstDay | null {
  if (!today || usual || trackedMS <= 0) return null;
  const unlabelledMS = Math.min(byKind[""] ?? 0, trackedMS);
  return {
    recorded: dur(trackedMS),
    top: apps.slice(0, 3).map((a) => a.name),
    labelled: dur(trackedMS - unlabelledMS),
    unlabelled: dur(unlabelledMS),
    unlabelledMS,
  };
}
