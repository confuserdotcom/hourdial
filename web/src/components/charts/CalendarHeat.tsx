import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import type { MonthDay } from "@/lib/api";
import { dur, parseDate, dayTitle, weekHeads, weekStartOf } from "@/lib/format";
import { useRoving } from "./Marks";

const HEAD = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]; // Sunday first, see weekHeads

/** CalendarHeat: the month as a calendar, each day shaded by deep time.
 *  It is one tab stop: the arrows move a day or a week, Enter opens one. */
export function CalendarHeat({ first, days, onDay }: { first: string; days: MonthDay[]; onDay?: (date: string) => void }) {
  const start = parseDate(first);
  const lead = Math.round((start.getTime() - weekStartOf(start).getTime()) / 86_400_000);
  const max = Math.max(1, ...days.map((d) => d.deepMS));
  const cells: (MonthDay | null)[] = [...Array(lead).fill(null), ...days];
  const { item } = useRoving(days.length, (i) => onDay?.(days[i].date), { cols: 7 });
  const say = (d: MonthDay) => d.score < 0 ? "Not a workday" : `${dur(d.deepMS)} deep · score ${d.score} · ${dur(d.trackedMS)} active`;
  return (
    <div className="grid grid-cols-7 gap-1.5" role="group" aria-label="Deep time by day. The arrow keys move a day or a week, Enter opens one.">
      {weekHeads(HEAD).map((h) => (
        <span key={h} className="pb-1 text-center text-[11px] text-faint">{h}</span>
      ))}
      {cells.map((d, i) =>
        d ? (
          <Tooltip key={d.date}>
            <TooltipTrigger asChild>
              <button
                type="button"
                {...item(i - lead)}
                onClick={() => onDay?.(d.date)}
                aria-label={`${dayTitle(parseDate(d.date))}, ${say(d).replaceAll(" · ", ", ")}`}
                className={`${d.score < 0 ? "hatch " : ""}flex aspect-[1.6] items-start justify-end rounded-md p-1.5 transition-[filter] hover:brightness-125 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring`}
                style={d.score < 0 ? undefined : { background: `hsl(var(--k-deep) / ${0.12 + 0.88 * (d.deepMS / max)})` }}
              >
                <span className="num text-[11px] text-foreground/80">{parseDate(d.date).getDate()}</span>
              </button>
            </TooltipTrigger>
            <TooltipContent className="num text-xs">
              <div className="font-medium">{dayTitle(parseDate(d.date))}</div>
              {say(d)}
            </TooltipContent>
          </Tooltip>
        ) : (
          <span key={`x${i}`} />
        ),
      )}
    </div>
  );
}
