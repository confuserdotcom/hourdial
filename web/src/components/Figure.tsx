import type { ReactNode } from "react";
import { dur, num } from "@/lib/format";
import { useCountUp } from "@/lib/motion";

/** Figure is one headline number of a page's lead card, with its label
 *  above and what it is made of beneath. i staggers its entrance. */
export function Figure({ i, label, note, mark, children }: { i: number; label: string; note: ReactNode; mark?: "agent" | "you"; children: ReactNode }) {
  return (
    <div className="anim-rise min-w-0" style={{ ["--i" as string]: i + 1 }}>
      <div className="lbl flex items-center gap-1.5">
        {mark && <i className={mark === "agent" ? "h-2 w-2 rounded-full bg-agent" : "h-2 w-2 rounded-full border border-foreground/70"} aria-hidden />}
        {label}
      </div>
      <div className="num mt-1.5 text-[32px] font-semibold leading-9 tracking-[-0.02em]">{children}</div>
      <div className="num mt-1 text-xs text-muted-foreground">{note}</div>
    </div>
  );
}

/** Dur counts up to a duration. */
export function Dur({ ms }: { ms: number }) {
  return <>{dur(useCountUp(ms))}</>;
}

/** Amount counts up to a number with digits decimals, then its unit. */
export function Amount({ v, digits = 2, unit }: { v: number; digits?: number; unit?: string }) {
  return <>{num(useCountUp(v), digits)}{unit && <span className="ml-1.5 text-base font-medium text-muted-foreground">{unit}</span>}</>;
}
