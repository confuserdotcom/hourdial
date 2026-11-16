import type { AgentSpan } from "@/lib/api";
import { lanes } from "@/lib/agents";
import { dur, hhmm, hourLabel } from "@/lib/format";

/** AgentLanes is when in the day the agents ran: a lane per project, the
 *  agent's stretches solid and yours with it outlined beneath, on the
 *  hours they span. */
export function AgentLanes({ spans, order }: { spans: AgentSpan[]; order: string[] }) {
  const l = lanes(spans, order);
  if (!l) return null;
  const hours = Array.from({ length: l.toH - l.fromH + 1 }, (_, i) => l.fromH + i);
  const stamp = (a: number, b: number) => {
    const s = l.t0 + ((l.t1 - l.t0) * a) / 100, e = l.t0 + ((l.t1 - l.t0) * b) / 100;
    return `${hhmm(s)} to ${hhmm(e)}, ${dur(e - s)}`;
  };
  return (
    <div className="grid grid-cols-[minmax(84px,16%)_1fr] items-center gap-x-4 gap-y-3 text-sm">
      {l.lanes.map((lane, i) => (
        <div key={lane.project} className="contents">
          <div className="truncate font-medium" title={lane.project}>{lane.project}</div>
          <div className="relative h-7" role="img"
            aria-label={`${lane.project}: agent ${lane.agent.map(([a, b]) => stamp(a, b)).join("; ") || "none"}. You ${lane.you.map(([a, b]) => stamp(a, b)).join("; ") || "none"}.`}>
            <div className="absolute inset-0 flex justify-between" aria-hidden>
              {hours.map((h) => <span key={h} className="h-full w-px bg-border/70" />)}
            </div>
            {lane.agent.map(([a, b]) => (
              <div key={`a${a}`} className="absolute top-1 h-2.5" style={{ left: `${a}%`, width: `${b - a}%` }} title={`Agent ${stamp(a, b)}`}>
                <div className="anim-grow h-full rounded-[3px] bg-agent" style={{ ["--i" as string]: i }} />
              </div>
            ))}
            {lane.you.map(([a, b]) => (
              <div key={`y${a}`} className="absolute bottom-1 h-2" style={{ left: `${a}%`, width: `${b - a}%` }} title={`You ${stamp(a, b)}`}>
                <div className="anim-grow h-full rounded-[3px] border border-foreground/70" style={{ ["--i" as string]: i + 1 }} />
              </div>
            ))}
          </div>
        </div>
      ))}
      <div aria-hidden />
      <div className="num relative h-4 text-[11px] text-faint" aria-hidden>
        {hours.map((h, i) => (
          <span key={h} className={i === 0 ? "absolute" : i === hours.length - 1 ? "absolute -translate-x-full" : "absolute -translate-x-1/2"}
            style={{ left: `${(i / (hours.length - 1)) * 100}%` }}>{hourLabel(h % 24)}</span>
        ))}
      </div>
    </div>
  );
}
