import { ChevronDown, ReceiptText, Tags } from "lucide-react";
import { ramp } from "@/components/charts/Donut";
import { Amount, Dur, Figure } from "@/components/Figure";
import { Empty, Panel } from "@/components/Panel";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useApi, type Agents, type Budget, type Library, type Top as TopT } from "@/lib/api";
import { dayTitle, dur, num, parseDate, pct } from "@/lib/format";
import { money, projectRows, rangeDays, type ProjectRow } from "@/lib/projects";
import { go, rangeQuery, type Route } from "@/lib/route";
import { cn } from "@/lib/utils";
import Top from "@/views/Top";
import Trends from "@/views/Trends";
import Week from "@/views/Week";

/** Projects is where the time went per project in the range: what of it is
 *  billable and worth, the budget left, and under that the charts of the
 *  week, the month or the day. */
export default function Projects({ route, tick }: { route: Route; tick: number }) {
  const q = rangeQuery(route) + (route.unit === "day" ? "&range=day" : "");
  const { data: top, error, loading } = useApi<TopT>(`/api/top?${q}&series=project`, tick);
  const { data: lib } = useApi<Library>("/api/library", tick);
  const { data: budgets } = useApi<Budget[]>("/api/budgets", tick);
  const { data: ag } = useApi<Agents>(`/api/agents?${rangeQuery(route)}`, tick);
  return (
    <div className="space-y-4">
      {error ? <Panel label="Projects"><Empty>Couldn't load: {error}</Empty></Panel>
        : loading && !top ? <><Skeleton className="h-40" /><Skeleton className="h-64" /></>
          : top && <Ledger top={top} lib={lib} budgets={budgets} ag={ag} open={route.project} />}
      <div className="pt-4">
        {route.unit === "week" ? <Week date={route.date} tick={tick} />
          : route.unit === "month" ? <Trends date={route.date} />
            : <Top tick={tick} query={q} />}
      </div>
    </div>
  );
}

function Ledger({ top, lib, budgets, ag, open }: { top: TopT; lib: Library | null; budgets: Budget[] | null; ag: Agents | null; open: string }) {
  const p = projectRows(top, lib, budgets, ag);
  const colors = ramp(p.rows.length);
  const days = rangeDays(top.from, top.to);
  if (p.rows.length === 0) {
    return (
      <Panel label="Projects">
        <Empty>No time on a project in this range. A project gets its time from rules: a folder, a window title, a site.</Empty>
        <div className="flex justify-center"><Button variant="outline" size="sm" className="press" onClick={() => go({ view: "rules", sub: "projects" })}>Set up projects</Button></div>
      </Panel>
    );
  }
  return (
    <>
      <section aria-labelledby="projects-totals" className="anim-rise glow overflow-hidden rounded-lg border bg-card">
        <h2 id="projects-totals" className="sr-only">Totals</h2>
        <div className="grid gap-x-10 gap-y-5 p-6 sm:grid-cols-3">
          <Figure i={0} label="On projects" note={`${p.rows.length} ${p.rows.length === 1 ? "project" : "projects"} · ${dur(p.otherMS)} on none`}><Dur ms={p.projectMS} /></Figure>
          <Figure i={1} label="Billable" note={p.value == null ? "no project has a rate" : `${pct(p.billableMS, p.projectMS)} of project time`}><Dur ms={p.billableMS} /></Figure>
          <Figure i={2} label="Value" note={p.value == null ? (lib?.billing ? "set a rate on a project in Rules" : "billing is off in Settings") : "billable time at each project's rate"}>
            {p.value == null ? "–" : <Amount v={p.value} unit={p.currency} />}
          </Figure>
        </div>
        <div className="px-6 pb-6">
          <div className="anim-grow flex h-3 gap-0.5 overflow-hidden rounded-[4px]" role="img"
            aria-label={`Share of project time: ${p.rows.map((r) => `${r.name} ${num(r.share, 0)}%`).join(", ")}`}>
            {p.rows.map((r, i) => <div key={r.name} className="h-full min-w-[3px] transition-[filter] hover:brightness-125" style={{ width: `${r.share}%`, background: colors[i] }} title={`${r.name} ${dur(r.ms)}`} />)}
          </div>
          <ul className="mt-3 flex flex-wrap gap-x-5 gap-y-1 text-xs text-muted-foreground">
            {p.rows.slice(0, 8).map((r, i) => (
              <li key={r.name} className="flex items-center gap-1.5">
                <i className="h-2 w-2 rounded-[2px]" style={{ background: colors[i] }} aria-hidden />{r.name}<span className="num text-foreground">{num(r.share, 0)}%</span>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <Panel label="By project" aside={`${dayTitle(parseDate(top.from))}${top.to !== top.from ? ` – ${dayTitle(parseDate(top.to))}` : ""}`}>
        <ul>
          {p.rows.map((r, i) => (
            <Row key={r.name} i={i} r={r} color={colors[i]} currency={p.currency} days={days} open={open === r.name} from={top.from} to={top.to} />
          ))}
        </ul>
      </Panel>
    </>
  );
}

function Row({ i, r, color, currency, days, open, from, to }: { i: number; r: ProjectRow; color: string; currency: string; days: string[]; open: boolean; from: string; to: string }) {
  const id = `project-${i}`;
  return (
    <li className="border-t first:border-t-0">
      <button aria-expanded={open} aria-controls={id} onClick={() => go({ project: open ? "" : r.name })}
        className="group grid w-full grid-cols-[minmax(0,1.2fr)_minmax(0,2fr)_auto] items-center gap-x-5 rounded-md px-2 py-3 text-left text-sm transition-colors duration-150 hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring md:grid-cols-[minmax(0,1.2fr)_96px_minmax(0,2fr)_120px_150px_20px]">
        <span className="flex min-w-0 items-center gap-2.5">
          <i className="h-2.5 w-2.5 shrink-0 rounded-[3px]" style={{ background: color }} aria-hidden />
          <span className="min-w-0">
            <span className="block truncate font-medium">{r.name}</span>
            {r.client && <span className="block truncate text-xs text-muted-foreground">{r.client}</span>}
          </span>
        </span>
        <span className="hidden md:block">{r.days.length > 1 && <Spark days={r.days} i={i} color={color} />}</span>
        <span className="flex items-center gap-2">
          <span className="h-2.5 min-w-0 flex-1 rounded-[3px] bg-foreground/[0.04]">
            <span className="block h-full" style={{ width: `${r.pct}%`, minWidth: 3 }}>
              <span className="anim-grow block h-full rounded-[3px] transition-[filter] group-hover:brightness-125" style={{ background: color, ["--i" as string]: i }} />
            </span>
          </span>
          <span className="num w-12 shrink-0 text-right">{dur(r.ms)}</span>
        </span>
        <span className="num hidden text-right md:block">
          {r.value == null ? <span className="text-muted-foreground">no rate</span> : <>{money(r.value, currency)}<span className="block text-xs text-muted-foreground">at {num(r.rate as number, 0)}/h</span></>}
        </span>
        <span className="hidden md:block">{r.budget ? <Meter b={r.budget} i={i} /> : <span className="text-xs text-muted-foreground">no budget</span>}</span>
        <ChevronDown className={cn("hidden h-4 w-4 text-muted-foreground transition-transform duration-200 md:block", open && "rotate-180")} aria-hidden />
      </button>
      {open && (
        <div id={id} className="anim-rise grid gap-6 px-2 pb-5 pt-1 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
          {r.days.length > 1
            ? <DayChart days={r.days} dates={days} color={color} name={r.name} />
            : <p className="text-sm text-muted-foreground">{dur(r.ms)} on {dayTitle(parseDate(from))}. Pick Week or Month above for its days.</p>}
          <div className="space-y-3 text-sm">
            <dl className="space-y-1.5">
              <Fact k="Agents">{r.agentMS > 0 ? `${dur(r.agentMS)} of agent work, ${dur(r.youMS)} of yours with them` : "no agent sessions in this range"}</Fact>
              <Fact k="Budget">{r.budget ? `${dur(r.budget.spentMS)} of ${r.budget.budget} spent${r.budget.over ? `, over by ${dur(r.budget.spentMS - r.budget.budgetMS)}` : ""}${r.budget.since ? `, since ${r.budget.since}` : ""}` : "none set"}</Fact>
              <Fact k="Rate">{r.rate == null ? "none" : `${num(r.rate, 0)} ${currency} an hour${r.client ? `, billed to ${r.client}` : ""}`}</Fact>
            </dl>
            <div className="flex flex-wrap gap-2 pt-1">
              {r.rate != null && (
                <Button size="sm" className="press gap-1.5" onClick={() => go({ view: "invoices", project: r.name, unit: "custom", date: from, to })}>
                  <ReceiptText className="h-3.5 w-3.5" />Draft invoice
                </Button>
              )}
              <Button variant="outline" size="sm" className="press gap-1.5" onClick={() => go({ view: "rules", sub: "projects" })}>
                <Tags className="h-3.5 w-3.5" />Rate, client and rules
              </Button>
            </div>
          </div>
        </div>
      )}
    </li>
  );
}

function Fact({ k, children }: { k: string; children: React.ReactNode }) {
  return <div className="flex gap-3"><dt className="lbl w-14 shrink-0 pt-0.5">{k}</dt><dd className="num min-w-0">{children}</dd></div>;
}

/** Spark is a project's days as small bars, in its row. */
function Spark({ days, i, color }: { days: number[]; i: number; color: string }) {
  const top = Math.max(1, ...days);
  return (
    <span className="flex h-6 items-end gap-px" aria-hidden>
      {days.map((ms, d) => (
        <span key={d} className="anim-grow-y min-w-0 flex-1 rounded-[1px]" style={{ height: `${Math.max(ms ? 8 : 0, (ms / top) * 100)}%`, background: color, ["--i" as string]: i + d * 0.4 }} />
      ))}
    </span>
  );
}

/** DayChart is the opened project's time per day, bars from zero. */
function DayChart({ days, dates, color, name }: { days: number[]; dates: string[]; color: string; name: string }) {
  const top = Math.max(1, ...days);
  const best = days.indexOf(Math.max(...days));
  const label = (d: number) => (dates[d] ? dayTitle(parseDate(dates[d])) : "");
  return (
    <div role="img" aria-label={`${name} per day: ${days.map((ms, d) => `${label(d)} ${dur(ms)}`).join(", ")}`}>
      <div className="flex h-28 items-end gap-[3px] border-b">
        {days.map((ms, d) => (
          <div key={d} className="group/bar relative flex h-full min-w-0 flex-1 items-end" title={`${label(d)} · ${dur(ms)}`}>
            <div className="anim-grow-y w-full rounded-t-[3px] transition-[filter] group-hover/bar:brightness-125"
              style={{ height: `${(ms / top) * 100}%`, minHeight: ms ? 2 : 0, background: color, ["--i" as string]: d }} />
          </div>
        ))}
      </div>
      <div className="num mt-1.5 flex justify-between text-[11px] text-faint" aria-hidden>
        <span>{label(0)}</span>
        <span className="text-muted-foreground">most on {label(best)}: {dur(days[best])}</span>
        <span>{label(days.length - 1)}</span>
      </div>
    </div>
  );
}

/** Meter is a budget: spent of the whole, red once over. */
function Meter({ b, i }: { b: Budget; i: number }) {
  return (
    <span className="block" title={`${dur(b.spentMS)} of ${b.budget}`}>
      <span className="num flex justify-between text-xs"><span className={cn(b.over && "text-bad")}>{b.pct}%</span><span className="text-muted-foreground">of {b.budget}</span></span>
      <span className="mt-1 block h-1.5 overflow-hidden rounded-[2px] bg-foreground/[0.07]">
        <span className="block h-full" style={{ width: `${Math.min(b.pct, 100)}%` }}>
          <span className="anim-grow block h-full rounded-[2px]" style={{ ["--i" as string]: i + 1, background: b.over ? "hsl(var(--k-distract))" : "hsl(var(--foreground) / 0.6)" }} />
        </span>
      </span>
    </span>
  );
}
