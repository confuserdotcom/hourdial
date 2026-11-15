import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip } from "@/components/ui/chart";
import { SERIES, kindColor, kindFill, kindLabel } from "@/lib/kinds";
import { dur, hourLabel, minClock } from "@/lib/format";
import { hourName } from "@/lib/roving";
import { KindTable } from "./KindTable";
import { Marks, plotInset } from "./Marks";

/** HourBars: minutes per hour, stacked by kind, from hourFirst to hourLast. */
/** onHour gets the hour of a clicked bar; with it the bars can be picked
 *  with the keyboard too (Tab, then the arrows, Enter). */
export function HourBars({ hours, first, last, height = 220, onHour }: { hours: Record<string, number[]>; first: number; last: number; height?: number; onHour?: (h: number) => void }) {
  const rows = [];
  const byKind = (h: number) => Object.fromEntries(SERIES.map((k) => [k, hours[k]?.[h] ?? 0]));
  for (let h = first; h <= last; h++) {
    const r: Record<string, number | string> = { h: hourLabel(h), at: minClock(h * 60) };
    for (const k of SERIES) r[k || "unlabelled"] = Math.round((hours[k]?.[h] ?? 0) / 60_000);
    rows.push(r);
  }
  return (
    <>
    <div className="relative">
    <ChartContainer config={{}} className="w-full" style={{ height }} aria-hidden>
      <BarChart data={rows} margin={{ top: 4, right: 0, left: -24, bottom: 0 }} className={onHour ? "cursor-pointer" : undefined}
        onClick={(e) => e?.activeTooltipIndex !== undefined && onHour?.(first + e.activeTooltipIndex)}>
        <CartesianGrid vertical={false} strokeDasharray="3 3" stroke="hsl(var(--border))" />
        <XAxis dataKey="h" tickLine={false} axisLine={false} tick={{ fill: "hsl(var(--faint-text))", fontSize: 11 }} />
        <YAxis domain={[0, 60]} ticks={[0, 15, 30, 45, 60]} tickLine={false} axisLine={false} tick={{ fill: "hsl(var(--faint-text))", fontSize: 11 }} />
        <ChartTooltip
          cursor={{ fill: "hsl(var(--secondary))" }}
          content={({ active, payload, label }) =>
            active && payload?.length ? (
              <div className="rounded-md border bg-popover px-2.5 py-1.5 text-xs shadow">
                <div className="mb-1 font-medium">{label}:00</div>
                {onHour && <div className="mb-1 text-faint">Click to zoom the timeline</div>}
                {payload.filter((p) => Number(p.value) > 0).map((p) => (
                  <div key={String(p.dataKey)} className="flex items-center gap-2">
                    <i className="h-2 w-2 rounded-sm" style={{ background: kindColor(p.dataKey === "unlabelled" ? "" : String(p.dataKey)) }} />
                    <span className="flex-1">{kindLabel(p.dataKey === "unlabelled" ? "" : String(p.dataKey))}</span>
                    <span className="num">{dur(Number(p.value) * 60_000)}</span>
                  </div>
                ))}
              </div>
            ) : null
          }
        />
        {SERIES.map((k, i) => (
          <Bar key={k || "unlabelled"} dataKey={k || "unlabelled"} stackId="a" fill={kindFill(k)} radius={i === SERIES.length - 1 ? [2, 2, 0, 0] : 0} isAnimationActive={false} />
        ))}
      </BarChart>
    </ChartContainer>
    {onHour && (
      <Marks label="Minutes per hour. The arrow keys move between hours, Enter zooms the timeline to one."
        names={rows.map((_, i) => hourName(first + i, byKind(first + i)))} inset={plotInset(4, 0)}
        onPick={(i) => onHour(first + i)} readout={(i) => hourName(first + i, byKind(first + i))} />
    )}
    </div>
    <KindTable caption="Minutes per hour, by kind of time" rows={rows.map((r) => ({ label: String(r.at), byKind: Object.fromEntries(SERIES.map((k) => [k, Number(r[k || "unlabelled"]) * 60_000])) }))} />
    </>
  );
}
