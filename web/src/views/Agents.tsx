import { ArrowUpRight } from "lucide-react";
import { AgentLanes } from "@/components/charts/AgentLanes";
import { LeverageDial } from "@/components/charts/LeverageDial";
import { Empty, Panel } from "@/components/Panel";
import { Skeleton } from "@/components/ui/skeleton";
import { ledger, type LedgerRow } from "@/lib/agents";
import { useApi, type Agents as AgentsT, type Library } from "@/lib/api";
import { dur, num } from "@/lib/format";
import { useCountUp } from "@/lib/motion";
import { go, rangeQuery, type Route } from "@/lib/route";

/** Agents is the agent ledger of the range: what the coding agents worked
 *  on per project, and how much of your own time went with them. */
export default function Agents({ route, tick }: { route: Route; tick: number }) {
  const { data: ag, error, loading } = useApi<AgentsT>(`/api/agents?${rangeQuery(route)}`, tick);
  const { data: lib } = useApi<Library>("/api/library", tick);
  if (error) return <Panel label="Agents"><Empty>Couldn't load: {error}</Empty></Panel>;
  if (loading && !ag) return <div className="space-y-4"><Skeleton className="h-44" /><Skeleton className="h-64" /></div>;
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
      <section aria-labelledby="agents-totals" className="anim-rise glow overflow-hidden rounded-lg border bg-card">
        <h2 id="agents-totals" className="sr-only">Totals</h2>
        <div className="flex flex-wrap items-center gap-x-10 gap-y-4 p-6">
          <LeverageDial agentMS={ag.agentMS} youMS={ag.youMS} />
          <div className="grid min-w-0 flex-1 gap-x-10 gap-y-5 sm:grid-cols-3">
            <Figure i={0} label="Agent time" ms={ag.agentMS} mark="agent"
              note={`${l.sessions} ${l.sessions === 1 ? "session" : "sessions"}${ag.peak > 1 ? ` · up to ${ag.peak} at once` : ""} · ${dur(ag.wallMS)} on the clock`} />
            <Figure i={1} label="Your time with them" ms={ag.youMS} mark="you"
              note={l.value ? `${num(l.value, 2)} ${l.currency} at your rates` : "with an agent's terminal in front of you"} />
            <Figure i={2} label="Leverage" text={l.leverage > 0 ? `${num(l.leverage, 1)}×` : "–"}
              note={l.leverage > 0 ? `${dur(ag.agentMS)} of agent work ÷ ${dur(ag.youMS)} of yours` : "none of your own time in their windows"} />
          </div>
        </div>
      </section>

      {ag.spans && ag.spans.length > 0 && (
        <Panel label="When they ran" aside={<Key />}>
          <AgentLanes spans={ag.spans} order={ag.rows.map((r) => r.project)} />
        </Panel>
      )}

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
            {l.rows.map((r, i) => <Row key={r.project} i={i} r={r} priced={priced} currency={l.currency} />)}
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

/** Figure is one headline number; a duration counts up to its value. */
function Figure({ i, label, ms, text, note, mark }: { i: number; label: string; ms?: number; text?: string; note: string; mark?: "agent" | "you" }) {
  const shown = useCountUp(ms ?? 0);
  return (
    <div className="anim-rise min-w-0" style={{ ["--i" as string]: i + 1 }}>
      <div className="lbl flex items-center gap-1.5">
        {mark && <i className={mark === "agent" ? "h-2 w-2 rounded-full bg-agent" : "h-2 w-2 rounded-full border border-foreground/70"} aria-hidden />}
        {label}
      </div>
      <div className="num mt-1.5 text-[32px] font-semibold leading-9 tracking-[-0.02em]">{text ?? dur(shown)}</div>
      <div className="num mt-1 text-xs text-muted-foreground">{note}</div>
    </div>
  );
}

/** Key says which mark is which: solid is the agent, outlined is you. */
function Key() {
  return (
    <span className="flex items-center gap-3">
      <span className="flex items-center gap-1.5"><i className="h-2 w-4 rounded-[2px] bg-agent" aria-hidden />Agent</span>
      <span className="flex items-center gap-1.5"><i className="h-2 w-4 rounded-[2px] border border-foreground/70" aria-hidden />You</span>
    </span>
  );
}

function Row({ i, r, priced, currency }: { i: number; r: LedgerRow; priced: boolean; currency: string }) {
  return (
    <tr className="group border-t align-middle transition-colors duration-150 hover:bg-accent/50">
      <td className="max-w-0 py-2.5 pr-3">
        <button onClick={() => go({ view: "projects" })} title={`Open ${r.project} in Projects`}
          className="flex max-w-full items-center gap-1 rounded-sm text-left font-medium underline-offset-2 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <span className="truncate">{r.project}</span>
          <ArrowUpRight className="h-3.5 w-3.5 shrink-0 -translate-x-1 text-muted-foreground opacity-0 transition duration-200 group-hover:translate-x-0 group-hover:opacity-100 group-focus-within:translate-x-0 group-focus-within:opacity-100" aria-hidden />
        </button>
      </td>
      <td className="py-2.5 pr-4">
        <Bar i={i} pct={r.agentPct} text={dur(r.agentMS)} name="Agent" solid />
        <Bar i={i + 1} pct={r.youPct} text={dur(r.youMS)} name="You" />
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
 *  it: solid for the agent, outlined for you. It grows from zero. */
function Bar({ i, pct, text, name, solid }: { i: number; pct: number; text: string; name: string; solid?: boolean }) {
  return (
    <div className="flex h-5 items-center gap-2" role="img" aria-label={`${name} ${text}`}>
      <div className="h-2.5 min-w-0 flex-1 rounded-[3px] bg-foreground/[0.04]">
        <div className="h-full" style={{ width: `${pct}%`, minWidth: pct > 0 ? 3 : 0 }}>
          <div className={solid ? "anim-grow h-full rounded-[3px] bg-agent transition-[filter] group-hover:brightness-110" : "anim-grow h-full rounded-[3px] border border-foreground/70"}
            style={{ ["--i" as string]: i }} />
        </div>
      </div>
      <span className="num w-12 shrink-0 text-right text-xs">{text}</span>
    </div>
  );
}
