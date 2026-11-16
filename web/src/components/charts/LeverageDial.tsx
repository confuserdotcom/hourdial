import { useCountUp } from "@/lib/motion";
import { dur, num } from "@/lib/format";

const SIZE = 148, C = SIZE / 2;

/** ring is a full circle of radius r, starting at 12 o'clock, clockwise. */
function ring(r: number): string {
  return `M ${C} ${C - r} A ${r} ${r} 0 1 1 ${C - 0.01} ${C - r}`;
}

/** LeverageDial is two rings on one scale, the longer of the two times a
 *  full turn: the agents' outside, yours inside, and the ratio between
 *  them in the middle. Both draw themselves in. */
export function LeverageDial({ agentMS, youMS }: { agentMS: number; youMS: number }) {
  const top = Math.max(agentMS, youMS, 1);
  const lev = useCountUp(youMS > 0 ? agentMS / youMS : 0);
  const arc = (r: number, part: number, cls: string, i: number) => part > 0 && (
    <path d={ring(r)} pathLength={1} fill="none" strokeWidth={10} strokeLinecap="round" className={cls}
      style={{ strokeDasharray: `${Math.min(part, 0.999)} 1`, ["--i" as string]: i }} />
  );
  return (
    <svg viewBox={`0 0 ${SIZE} ${SIZE}`} className="h-[148px] w-[148px] shrink-0" role="img"
      aria-label={`Agents ${dur(agentMS)}, you ${dur(youMS)}${youMS > 0 ? `, ${num(agentMS / youMS, 1)} times your time` : ""}`}>
      <circle cx={C} cy={C} r={62} fill="none" strokeWidth={10} className="stroke-foreground/[0.07]" />
      <circle cx={C} cy={C} r={46} fill="none" strokeWidth={10} className="stroke-foreground/[0.07]" />
      <g className="anim-dial">
        {arc(62, agentMS / top, "stroke-agent", 0)}
        {arc(46, youMS / top, "stroke-foreground/70", 1)}
      </g>
      <text x={C} y={C + 1} textAnchor="middle" dominantBaseline="middle" className="num fill-foreground text-[22px] font-semibold">
        {youMS > 0 ? `${num(lev, 1)}×` : "–"}
      </text>
    </svg>
  );
}
