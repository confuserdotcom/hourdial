import { useState } from "react";
import { InvoiceDialog } from "@/components/InvoiceDialog";
import { Empty, Panel } from "@/components/Panel";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useApi, type Library } from "@/lib/api";
import { isoDate, weekStartOf } from "@/lib/format";
import { go } from "@/lib/route";

/** Invoices bills a project with a rate for a period. */
export default function Invoices({ tick }: { tick: number }) {
  const { data: lib, error, loading } = useApi<Library>("/api/library", tick);
  const [open, setOpen] = useState(false);
  if (error) return <Panel label="Invoices"><Empty>Couldn't load: {error}</Empty></Panel>;
  if (loading && !lib) return <Skeleton className="h-48" />;
  const rated = lib?.projects.some((p) => p.rate != null && !p.archived);
  return (
    <Panel label="Invoices">
      {!lib?.billing ? (
        <div className="space-y-3 text-sm">
          <p className="max-w-prose text-muted-foreground">
            Billing is off. With it on, a project can carry a rate and a client, and its recorded time becomes a numbered invoice.
          </p>
          <Button variant="outline" size="sm" onClick={() => go({ view: "settings" })}>Turn billing on in Settings</Button>
        </div>
      ) : (
        <div className="space-y-3 text-sm">
          <p className="max-w-prose text-muted-foreground">
            {rated
              ? "Draft an invoice from a project's recorded time. Issuing gives it the next number, for good."
              : "No project has a rate yet. Give one a rate, then draft its invoice here."}
          </p>
          {rated
            ? <Button size="sm" onClick={() => setOpen(true)}>Draft an invoice</Button>
            : <Button variant="outline" size="sm" onClick={() => go({ view: "rules", sub: "projects" })}>Set a rate</Button>}
        </div>
      )}
      <InvoiceDialog open={open} monday={isoDate(weekStartOf(new Date()))} onClose={() => setOpen(false)} />
    </Panel>
  );
}
