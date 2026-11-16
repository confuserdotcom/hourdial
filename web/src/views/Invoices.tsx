import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Download, ReceiptText } from "lucide-react";
import { Amount } from "@/components/Figure";
import { Empty, Panel } from "@/components/Panel";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { get, post, useApi, useSample, type InvoiceIssued, type InvoicePreview, type IssuedInvoice, type Library } from "@/lib/api";
import { download } from "@/lib/download";
import { dur, isoDate, weekStartOf } from "@/lib/format";
import { useLive } from "@/lib/live";
import { go, type Route } from "@/lib/route";
import { periods } from "@/lib/rules";
import { cn } from "@/lib/utils";

const field = "h-8 w-full rounded-md border bg-background px-2.5 text-sm placeholder:text-faint focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring";

/** Invoices is the money path: what was issued on the left, and a draft
 *  from one project's time on the right, which Issue turns into the next
 *  number for good. With Billing off it says what Billing is and turns it on. */
export default function Invoices({ route, tick }: { route: Route; tick: number }) {
  const { reload: reloadApp } = useLive();
  const sample = useSample();
  const lib = useApi<Library>("/api/library", tick);
  const billing = !!lib.data?.billing;
  const list = useApi<{ invoices: IssuedInvoice[] }>(billing ? "/api/invoices" : null, tick);
  const [busy, setBusy] = useState(false);
  if (lib.error) return <Panel label="Invoices"><Empty>Couldn't load: {lib.error}</Empty></Panel>;
  if (!lib.data) return <Skeleton className="h-64" />;
  if (!billing) {
    const on = async () => {
      setBusy(true);
      try {
        await post("/api/settings", { features: { billing: true } });
        reloadApp();
        lib.reload();
      } catch (e) {
        toast.error((e as Error).message);
      } finally {
        setBusy(false);
      }
    };
    return (
      <section aria-labelledby="billing-off" className="anim-rise glow mx-auto max-w-2xl rounded-lg border bg-card p-8">
        <ReceiptText className="h-6 w-6 text-agent" aria-hidden />
        <h2 id="billing-off" className="mt-4 text-xl font-semibold tracking-tight">Billing is off</h2>
        <p className="mt-2 max-w-prose text-sm text-muted-foreground">
          With Billing on, a project can carry an hourly rate and a client. Its recorded time then becomes an invoice: a draft
          you can check day by day, and, once issued, a numbered PDF that stays exactly as it was issued.
        </p>
        <ul className="mt-4 space-y-1.5 text-sm">
          <li>Rates and clients appear on projects in Rules.</li>
          <li>Projects shows what each one is worth in a range.</li>
          <li>Nothing leaves this machine: invoices are files in your own folder.</li>
        </ul>
        <Button className="press mt-6" disabled={busy} onClick={on}>Turn billing on</Button>
      </section>
    );
  }
  return (
    <div className="grid items-start gap-4 lg:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)]">
      <Issued list={list.data?.invoices ?? null} error={list.error} />
      {sample
        ? <Panel label="Draft"><Empty>These are sample invoices. Drafting and issuing open with hours of your own; the sample is never invoiced.</Empty></Panel>
        : <Draft lib={lib.data} route={route} onIssued={list.reload} />}
    </div>
  );
}

function Issued({ list, error }: { list: IssuedInvoice[] | null; error: string | null }) {
  const sample = useSample();
  async function pdf(n: string) {
    try {
      const r = await download(`/api/invoices/${encodeURIComponent(n)}.pdf`, `${n}.pdf`);
      if (r.error) toast.error(`Couldn't save ${r.name}: ${r.error}`);
      else if (r.browser) toast(`Downloading ${r.name}`);
      else if (r.saved) toast(`Saved ${r.name}`);
    } catch (e) {
      toast.error((e as Error).message);
    }
  }
  return (
    <Panel label="Issued" aside={list && list.length > 0 ? `${list.length} ${list.length === 1 ? "invoice" : "invoices"}` : undefined}>
      {error ? <Empty>Couldn't load: {error}</Empty> : !list ? <Skeleton className="h-40" /> : list.length === 0 ? (
        <Empty>Nothing issued yet. A draft becomes an invoice when you issue it; it gets the next number.</Empty>
      ) : (
        <ul>
          {list.map((i, n) => (
            <li key={i.number} className="anim-rise group flex items-center gap-4 border-t py-3 first:border-t-0" style={{ ["--i" as string]: Math.min(n, 8) }}>
              <div className={cn("min-w-0 flex-1", i.status === "void" && "opacity-60")}>
                <div className="flex items-center gap-2">
                  <span className="num font-medium">{i.number}</span>
                  <span className={cn("rounded-[4px] border px-1.5 py-px text-[11px] font-medium", i.status === "void" ? "border-bad/40 text-bad" : "border-agent/40 text-agent")}>{i.status}</span>
                  <span className="num text-xs text-muted-foreground">{i.issued}</span>
                </div>
                <div className="mt-0.5 truncate text-sm">
                  {i.project}{i.client && <span className="text-muted-foreground"> · {i.client}</span>}
                </div>
                <div className="num text-xs text-muted-foreground">
                  {i.from} to {i.to} · {dur(i.seconds * 1000)}
                  {i.voidReason && ` · void: ${i.voidReason}`}
                  {i.supersedes && ` · replaces ${i.supersedes}`}
                  {i.replacedBy && ` · replaced by ${i.replacedBy}`}
                </div>
              </div>
              <div className={cn("num text-right font-medium", i.status === "void" && "line-through opacity-60")}>{i.total}</div>
              {!sample && <button onClick={() => pdf(i.number)} aria-label={`Save the PDF of ${i.number}`} title="Save the stored PDF"
                className="press grid h-8 w-8 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                <Download className="h-4 w-4" />
              </button>}
            </li>
          ))}
        </ul>
      )}
    </Panel>
  );
}

function Draft({ lib, route, onIssued }: { lib: Library; route: Route; onIssued: () => void }) {
  const rated = lib.projects.filter((p) => p.rate != null && !p.archived);
  const ps = periods(isoDate(weekStartOf(new Date())));
  const given = route.unit === "custom";
  const [project, setProject] = useState(rated.some((p) => p.name === route.project) ? route.project : rated[0]?.name ?? "");
  const [from, setFrom] = useState(given ? route.date : ps[0].from);
  const [to, setTo] = useState(given ? route.to : ps[0].to);
  const [client, setClient] = useState("");
  const [number, setNumber] = useState("");
  const [prev, setPrev] = useState<InvoicePreview | null>(null);
  const [err, setErr] = useState("");
  const [okModel, setOkModel] = useState(false);
  const [okOverlap, setOkOverlap] = useState(false);
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const [n, setN] = useState(0);

  // A project and range handed over in the address (from Projects) win.
  useEffect(() => {
    if (route.project && rated.some((p) => p.name === route.project)) setProject(route.project);
    if (route.unit === "custom") (setFrom(route.date), setTo(route.to));
  }, [route.project, route.unit, route.date, route.to]);

  const q = () => {
    const u = new URLSearchParams({ project, from, to });
    if (client.trim()) u.set("client", client.trim());
    if (number.trim()) u.set("number", number.trim());
    return u.toString();
  };

  useEffect(() => {
    if (!project || !from || !to) return;
    let live = true;
    const t = setTimeout(() => {
      get<InvoicePreview>(`/api/invoice?${q()}&preview=1`)
        .then((p) => live && (setPrev(p), setErr(""), setOkModel(false), setOkOverlap(false)))
        .catch((e: Error) => live && (setPrev(null), setErr(e.message)));
    }, 250);
    return () => {
      live = false;
      clearTimeout(t);
    };
  }, [project, from, to, number, n]);

  async function save() {
    try {
      const r = await download(`/api/invoice?${q()}`, "invoice.pdf");
      if (r.error) toast.error(`Couldn't save ${r.name}: ${r.error}`);
      else if (r.browser) toast(`Downloading ${r.name}`);
      else if (r.saved) toast(`Saved ${r.name}`);
    } catch (e) {
      toast.error((e as Error).message);
    }
  }

  async function issue() {
    setBusy(true);
    try {
      const r = await post<InvoiceIssued>("/api/invoice/issue", {
        project, from, to, client: client.trim(), number: number.trim(), allowOverlap: okOverlap, confirmModel: okModel,
      });
      toast(`Issued ${r.number} · ${r.totalText}`, {
        description: "Stored with its PDF; it stays as issued. Its PDF is in the list.",
      });
      setConfirm(false);
      setNumber("");
      setN((x) => x + 1);
      onIssued();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  if (!rated.length) {
    return (
      <Panel label="Draft">
        <p className="text-sm text-muted-foreground">No project has an hourly rate yet. An invoice bills one project's time at its rate.</p>
        <Button variant="outline" size="sm" className="press mt-3" onClick={() => go({ view: "rules", sub: "projects" })}>Set a rate in Rules</Button>
      </Panel>
    );
  }
  const canIssue = !!prev && !busy && (prev.modelOnlyS === 0 || okModel) && (prev.overlaps.length === 0 || okOverlap);
  const named = prev?.number || number.trim();
  const picked = ps.findIndex((p) => p.from === from && p.to === to);
  return (
    <section aria-labelledby="draft-h" className="anim-rise glow rounded-lg border bg-card p-5">
      <h2 id="draft-h" className="lbl mb-3">Draft</h2>
      <form className="space-y-4" onSubmit={(e) => (e.preventDefault(), save())}>
        <label className="block space-y-1.5">
          <span className="lbl">Project</span>
          <select className={field} value={project} onChange={(e) => setProject(e.target.value)}>
            {rated.map((p) => <option key={p.name} value={p.name}>{p.name}</option>)}
          </select>
        </label>
        <div className="space-y-1.5">
          <span className="lbl">Period</span>
          <div className="flex flex-wrap gap-1.5">
            {ps.map((p, i) => (
              <button key={p.label} type="button" aria-pressed={picked === i} onClick={() => (setFrom(p.from), setTo(p.to))}
                className={cn("press h-7 rounded-md border px-2.5 text-xs text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", picked === i && "bg-accent font-medium text-foreground")}>
                {p.label}
              </button>
            ))}
          </div>
          <div className="grid grid-cols-2 gap-3">
            <label className="block space-y-1 text-xs text-muted-foreground">First day<input type="date" className={cn(field, "num")} value={from} onChange={(e) => setFrom(e.target.value)} /></label>
            <label className="block space-y-1 text-xs text-muted-foreground">Last day<input type="date" className={cn(field, "num")} value={to} onChange={(e) => setTo(e.target.value)} /></label>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <label className="block space-y-1.5">
            <span className="lbl">To (client)</span>
            <input className={field} value={client} onChange={(e) => setClient(e.target.value)} placeholder="the project's client" />
          </label>
          <label className="block space-y-1.5">
            <span className="lbl">Number</span>
            <input className={cn(field, "num")} value={number} onChange={(e) => setNumber(e.target.value)} placeholder={prev?.number || "next in line"} />
          </label>
        </div>

        <div className="rounded-md border bg-background/60 p-4" aria-live="polite">
          {err ? <p className="text-sm text-muted-foreground">{err}</p> : !prev ? <Skeleton className="h-14" /> : (
            <>
              <div className="lbl">Total</div>
              <div className="num mt-1 text-[28px] font-semibold leading-8 tracking-[-0.02em]"><Amount v={prev.total} unit={prev.currency} /></div>
              <p className="num mt-1 text-xs text-muted-foreground" data-preview>
                {dur(prev.seconds * 1000)} over {prev.days} {prev.days === 1 ? "day" : "days"} · <span>{prev.totalText}</span> · {named ? `as ${named}` : "takes the next number"}
              </p>
            </>
          )}
        </div>
        {prev && prev.modelOnlyS > 0 && (
          <label className="flex items-start gap-2 text-sm">
            <input type="checkbox" className="mt-1" checked={okModel} onChange={(e) => setOkModel(e.target.checked)} />
            <span>{dur(prev.modelOnlyS * 1000)} of it only the local model labelled. I've looked at it.</span>
          </label>
        )}
        {prev && prev.overlaps.length > 0 && (
          <label className="flex items-start gap-2 text-sm">
            <input type="checkbox" className="mt-1" checked={okOverlap} onChange={(e) => setOkOverlap(e.target.checked)} />
            <span>{prev.overlaps.join(", ")} already bills some of these days. Bill them again.</span>
          </label>
        )}
        <div className="flex flex-wrap justify-end gap-2">
          <Button type="submit" variant="outline" className="press" disabled={!prev}>Save DRAFT</Button>
          <Button type="button" className="press" disabled={!canIssue} onClick={() => setConfirm(true)}>Issue…</Button>
        </div>
      </form>

      <Dialog open={confirm} onOpenChange={setConfirm}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{named ? `Issue ${named}?` : "Issue this invoice?"}</DialogTitle>
            <DialogDescription>
              This is permanent and numbered. {named ? `${named} is` : "It takes the next number, is"} stored with its PDF and stays exactly as
              issued: {prev?.totalText} for {project}, {from} to {to}. A mistake is fixed by voiding it (<code>tock invoice void</code>) and
              issuing a new one, never by editing.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setConfirm(false)}>Not yet</Button>
            <Button className="press" disabled={!canIssue} onClick={issue}>{named ? `Issue ${named}` : "Issue it"}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
}
