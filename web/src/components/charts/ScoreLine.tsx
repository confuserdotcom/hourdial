import { CartesianGrid, Line, LineChart, ReferenceLine, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip } from "@/components/ui/chart";
import { Marks, plotInset } from "./Marks";
import { reducedMotion } from "@/lib/motion";

/** ScoreLine: the score per day; days that are not workdays are gaps.
 *  onPoint gets the index of a clicked day; with it the days can be
 *  picked with the keyboard too (Tab, then the arrows, Enter). */
export function ScoreLine({ points, height = 200, onPoint }: { points: { label: string; title?: string; score: number | null }[]; height?: number; onPoint?: (i: number) => void }) {
  const real = points.filter((p) => p.score !== null).map((p) => p.score as number);
  const mean = real.length ? Math.round(real.reduce((a, b) => a + b, 0) / real.length) : null;
  const name = (p: { label: string; title?: string; score: number | null }) => `${p.title ?? p.label}, ${p.score === null ? "no score" : `score ${p.score}`}`;
  return (
    <div className="relative">
    <ChartContainer config={{}} className="w-full" style={{ height }}>
      <LineChart data={points} margin={{ top: 8, right: 8, left: -24, bottom: 0 }} className={onPoint ? "cursor-pointer" : undefined}
        onClick={(e) => e?.activeTooltipIndex !== undefined && onPoint?.(e.activeTooltipIndex)}>
        <CartesianGrid vertical={false} strokeDasharray="3 3" stroke="hsl(var(--border))" />
        <XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: "hsl(var(--faint-text))", fontSize: 11 }} interval="preserveStartEnd" />
        <YAxis domain={[0, 100]} ticks={[0, 25, 50, 75, 100]} tickLine={false} axisLine={false} tick={{ fill: "hsl(var(--faint-text))", fontSize: 11 }} />
        {mean !== null && <ReferenceLine y={mean} stroke="hsl(var(--muted-foreground))" strokeDasharray="4 4" label={{ value: `mean ${mean}`, position: "insideTopRight", fill: "hsl(var(--muted-foreground))", fontSize: 11 }} />}
        <ChartTooltip
          cursor={{ stroke: "hsl(var(--border))" }}
          content={({ active, payload, label }) =>
            active && payload?.length && payload[0].value != null ? (
              <div className="rounded-md border bg-popover px-2.5 py-1.5 text-xs shadow">
                {label} · score <span className="num font-medium">{payload[0].value}</span>
              </div>
            ) : null
          }
        />
        <Line dataKey="score" stroke="hsl(var(--foreground))" strokeWidth={1.75} dot={{ r: 2.5, fill: "hsl(var(--foreground))" }} connectNulls={false} isAnimationActive={!reducedMotion()} animationDuration={720} animationEasing="ease-out" />
      </LineChart>
    </ChartContainer>
    {onPoint && (
      <Marks label="Focus score per day. The arrow keys move between days, Enter opens one." points
        names={points.map(name)} inset={plotInset(8, 8)} onPick={onPoint} readout={(i) => name(points[i])} />
    )}
    </div>
  );
}
