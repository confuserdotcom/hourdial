import { ArrowUpRight } from "lucide-react";
import { Empty, Panel } from "@/components/Panel";
import { Skeleton } from "@/components/ui/skeleton";
import { ledger, type LedgerRow } from "@/lib/agents";
import { useApi, type Agents as AgentsT, type Library } from "@/lib/api";
import { dur, num } from "@/lib/format";
import { go, rangeQuery, type Route } from "@/lib/route";

/** Agents is the agent ledger of the range: what the coding agents worked
 *  on per project, and how much of your own time went with them. */
export default function Agents({ route, tick }: { route: Route; tick: number }) {
  const { data: ag, error, loading } = useApi<AgentsT>(`/api/agents?${rangeQuery(route)}`, tick);
  const { data: lib } = useApi<Library>("/api/library", tick);
  if (error) return <Panel label="Agents"><Empty>Couldn't load: {error}</Empty></Panel>;
  if (loading && !ag) return <div className="space-y-4"><Skeleton className="h-28" /><Skeleton className="h-64" /></div>;
  if (!ag || ag.rows.length === 0) {
    return (
      <Panel label="Agents">
        <Empty>
          No agent sessions in this range. hourdial looks in <code className="num">~/.claude/projects</code> and reads
          only times, folders and titles, never prompts.
        </Empty>
      </Panel>
    );
  }
  const l = ledger(ag, lib);
  // The money column shows once some of your time is on a project with a rate.
  const priced = !!l.value;
  return (
    <div className="space-y-4">
      <section aria-labelledby="agents-totals" className="rounded-lg border bg-card">
        <h2 id="agents-totals" className="sr-only">Totals</h2>
        <div className="grid sm:grid-cols-3 sm:divide-x max-sm:divide-y">
        <Figure label="Agent time" value={dur(ag.agentMS)}
          note={`${l.sessions} ${l.sessions === 1 ? "session" : "sessions"}${ag.peak > 1 ? ` · up to ${ag.peak} at once · ${dur(ag.wallMS)} on the clock` : ""}`} />
        <Figure label="Your time with them" value={dur(ag.youMS)}
          note={l.value ? `${num(l.value, 2)} ${l.currency} at your rates` : "with an agent's terminal in front of you"} />
        <Figure label="Leverage" value={l.leverage > 0 ? `${num(l.leverage, 1)}×` : "–"}
          note={l.leverage > 0 ? `${dur(ag.agentMS)} of agent work ÷ ${dur(ag.youMS)} of yours` : "none of your own time in their windows"} />
        </div>
      </section>

      <Panel label="By project" aside={<Key />}>
        <table className="w-full text-sm">
          <thead>
            <tr className="lbl text-left">
              <th className="w-[22%] pb-2 font-medium">Project</th>
              <th className="pb-2 font-medium">Time</th>
              <th className="w-20 pb-2 text-right font-medium">Prompts</th>
              <th className="w-20 pb-2 text-right font-medium">Sessions</th>
              {priced && <th className="w-36 pb-2 text-right font-medium">Your time at rate</th>}
            </tr>
          </thead>
          <tbody>
            {l.rows.map((r) => <Row key={r.project} r={r} priced={priced} currency={l.currency} />)}
          </tbody>
        </table>
        <p className="mt-4 max-w-prose text-xs text-muted-foreground">
          From Claude Code's session logs in <code className="num">~/.claude/projects</code>: the time a session worked, and the time
          you had its terminal in front of you. Only times, folders and titles are read, never prompts.
        </p>
      </Panel>
    </div>
  );
}

function Figure({ label, value, note }: { label: string; value: string; note: string }) {
  return (
    <div className="p-5">
      <div className="lbl">{label}</div>
      <div className="num mt-1.5 text-[28px] font-semibold leading-8 tracking-[-0.01em]">{value}</div>
      <div className="num mt-1 text-xs text-muted-foreground">{note}</div>
    </div>
  );
}

/** Key says which bar is which: solid is the agent, outlined is you. */
function Key() {
  return (
    <span className="flex items-center gap-3">
      <span className="flex items-center gap-1.5"><i className="h-2 w-4 rounded-[2px] bg-foreground/80" aria-hidden />Agent</span>
      <span className="flex items-center gap-1.5"><i className="h-2 w-4 rounded-[2px] border border-foreground/70" aria-hidden />You</span>
    </span>
  );
}

function Row({ r, priced, currency }: { r: LedgerRow; priced: boolean; currency: string }) {
  return (
    <tr className="group border-t align-middle transition-colors duration-150 hover:bg-accent/50">
      <td className="max-w-0 py-2.5 pr-3">
        <button onClick={() => go({ view: "projects" })} title={`Open ${r.project} in Projects`}
          className="flex max-w-full items-center gap-1 rounded-sm text-left font-medium underline-offset-2 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <span className="truncate">{r.project}</span>
          <ArrowUpRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground opacity-0 transition-opacity duration-150 group-hover:opacity-100 group-focus-within:opacity-100" aria-hidden />
        </button>
      </td>
      <td className="py-2.5 pr-4">
        <Bar pct={r.agentPct} text={dur(r.agentMS)} name="Agent" solid />
        <Bar pct={r.youPct} text={dur(r.youMS)} name="You" />
      </td>
      <td className="num py-2.5 text-right text-muted-foreground">{r.prompts}</td>
      <td className="num py-2.5 text-right text-muted-foreground">{r.sessions}</td>
      {priced && (
        <td className="num py-2.5 text-right">
          {r.value == null ? <span className="text-muted-foreground">no rate</span> : r.value === 0 ? <span className="text-muted-foreground">–</span> : (
            <>
              {num(r.value, 2)} {currency}
              <div className="text-xs text-muted-foreground">{dur(r.youMS)} at {num(r.rate as number, 0)}/h</div>
            </>
          )}
        </td>
      )}
    </tr>
  );
}

/** Bar is one duration on the table's shared scale, with its number beside
 *  it: solid for the agent, outlined for you. */
function Bar({ pct, text, name, solid }: { pct: number; text: string; name: string; solid?: boolean }) {
  return (
    <div className="flex h-5 items-center gap-2" role="img" aria-label={`${name} ${text}`}>
      <div className="h-2 min-w-0 flex-1">
        <div className={solid ? "h-full rounded-[2px] bg-foreground/80" : "h-full rounded-[2px] border border-foreground/70"}
          style={{ width: `${pct}%`, minWidth: pct > 0 ? 3 : 0 }} />
      </div>
      <span className="num w-12 shrink-0 text-right text-xs">{text}</span>
    </div>
  );
}
