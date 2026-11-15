import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip } from "@/components/ui/chart";
import { SERIES, kindColor, kindFill, kindLabel } from "@/lib/kinds";
import { dur } from "@/lib/format";
import { dayName } from "@/lib/roving";
import { KindTable } from "./KindTable";
import { Marks, plotInset } from "./Marks";

/** title, when set, names the day in full for the keyboard ("Mon 3 Nov"). */
export type DayPoint = { label: string; title?: string; byKind: Record<string, number>; faint?: boolean };

/** DayBars: hours per day, stacked by kind. */
/** onDay gets the index of a clicked day; with it the days can be picked
 *  with the keyboard too (Tab, then the arrows, Enter). */
export function DayBars({ days, height = 240, onDay }: { days: DayPoint[]; height?: number; onDay?: (i: number) => void }) {
  const rows = days.map((d) => {
    const r: Record<string, number | string> = { label: d.label };
    for (const k of SERIES) r[k || "unlabelled"] = (d.byKind[k] ?? 0) / 3_600_000;
    return r;
  });
  return (
    <>
    <div className="relative">
    <ChartContainer config={{}} className="w-full" style={{ height }} aria-hidden>
      <BarChart data={rows} margin={{ top: 4, right: 0, left: -24, bottom: 0 }} className={onDay ? "cursor-pointer" : undefined}
        onClick={(e) => e?.activeTooltipIndex !== undefined && onDay?.(e.activeTooltipIndex)}>
        <CartesianGrid vertical={false} strokeDasharray="3 3" stroke="hsl(var(--border))" />
        <XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: "hsl(var(--faint-text))", fontSize: 11 }} />
        <YAxis tickLine={false} axisLine={false} tick={{ fill: "hsl(var(--faint-text))", fontSize: 11 }} tickFormatter={(v) => `${v}h`} allowDecimals={false} />
        <ChartTooltip
          cursor={{ fill: "hsl(var(--secondary))" }}
          content={({ active, payload, label }) =>
            active && payload?.length ? (
              <div className="rounded-md border bg-popover px-2.5 py-1.5 text-xs shadow">
                <div className="mb-1 font-medium">{label}</div>
                {onDay && <div className="mb-1 text-faint">Click to open the day</div>}
                {payload.filter((p) => Number(p.value) > 0).map((p) => (
                  <div key={String(p.dataKey)} className="flex items-center gap-2">
                    <i className="h-2 w-2 rounded-sm" style={{ background: kindColor(p.dataKey === "unlabelled" ? "" : String(p.dataKey)) }} />
                    <span className="flex-1">{kindLabel(p.dataKey === "unlabelled" ? "" : String(p.dataKey))}</span>
                    <span className="num">{dur(Number(p.value) * 3_600_000)}</span>
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
    {onDay && (
      <Marks label="Hours per day. The arrow keys move between days, Enter opens one."
        names={days.map((d) => dayName(d.title ?? d.label, d.byKind))} inset={plotInset(4, 0)}
        onPick={onDay} readout={(i) => dayName(days[i].title ?? days[i].label, days[i].byKind)} />
    )}
    </div>
    <KindTable caption="Hours per day, by kind of time" rows={days.map((d) => ({ label: d.label, byKind: d.byKind }))} />
    </>
  );
}
