import { AgentsPanel } from "@/components/AgentsBudgets";
import { Empty, Panel } from "@/components/Panel";
import { Skeleton } from "@/components/ui/skeleton";
import { useApi, type Agents as AgentsT } from "@/lib/api";
import { rangeQuery, type Route } from "@/lib/route";

/** Agents is the agent ledger of the range: what the coding agents worked
 *  on per project, and how much of your own time went with them. */
export default function Agents({ route, tick }: { route: Route; tick: number }) {
  const { data: ag, error, loading } = useApi<AgentsT>(`/api/agents?${rangeQuery(route)}`, tick);
  if (error) return <Panel label="Agents"><Empty>Couldn't load: {error}</Empty></Panel>;
  if (loading && !ag) return <Skeleton className="h-64" />;
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
  return <AgentsPanel ag={ag} rows={ag.rows.length} />;
}
