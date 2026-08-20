import { useState } from "react";
import { AlertTriangle, CheckCircle2, Server, ServerOff, Gauge, DollarSign } from "lucide-react";
import { useStore } from "../../store";
import { fmtMoney } from "../../data/mock";
import { Badge, PageHead, Spark, StatCard } from "../../components/ui";
import AdminNav from "./AdminNav";

export default function AdminHosts() {
  const { hosts: initial } = useStore();
  const [rows, setRows] = useState(initial);
  const [flagged, setFlagged] = useState({});

  const totals = rows.reduce(
    (acc, h) => {
      acc.total += h.num_gpus_total;
      acc.rented += h.num_gpus;
      acc.rate += h.dph_total;
      acc.verified += h.verified ? 1 : 0;
      return acc;
    },
    { total: 0, rented: 0, rate: 0, verified: 0 },
  );

  function toggleVerified(id) {
    setRows((prev) => prev.map((h) => (h.id === id ? { ...h, verified: !h.verified } : h)));
  }

  function toggleFlag(id) {
    setFlagged((prev) => ({ ...prev, [id]: !prev[id] }));
  }

  return (
    <div>
      <PageHead title="Hosts" sub="Approve and monitor every machine renting GPUs to the marketplace." />
      <AdminNav />

      <div className="grid cols-4">
        <StatCard label="Machines" value={rows.length} icon={Server} tone="var(--blue)" />
        <StatCard label="Total GPUs" value={totals.total} sub={`${totals.rented} rented now`} icon={Gauge} tone="var(--purple)" />
        <StatCard label="Hourly platform rate" value={fmtMoney(totals.rate)} icon={DollarSign} tone="var(--green)" />
        <StatCard label="Verified" value={`${totals.verified}/${rows.length}`} icon={CheckCircle2} tone="var(--amber)" />
      </div>

      <div className="mt">
        {rows.map((h) => {
          const bad = !h.verified || h.uptime < 95 || flagged[h.id];
          return (
            <div className="card" key={h.id} style={bad ? { borderColor: "rgba(207, 34, 46, 0.4)" } : undefined}>
              <div className="row between wrap" style={{ gap: 12 }}>
                <div>
                  <div className="row" style={{ gap: 8 }}>
                    <h3>{h.gpu_name}</h3>
                    {h.verified ? <Badge tone="green" dot>verified</Badge> : <Badge tone="neutral">unverified</Badge>}
                    {flagged[h.id] && <Badge tone="red" dot>flagged</Badge>}
                    {h.datacenter && <Badge tone="purple">datacenter</Badge>}
                  </div>
                  <div className="small dim mono mt-sm">host #{h.id} · {h.machine_id}</div>
                </div>
                <div className="row" style={{ gap: 20, textAlign: "right" }}>
                  <div>
                    <div className="muted small" style={{ fontSize: 11 }}>Rented / Total</div>
                    <div style={{ fontWeight: 700 }}>{h.num_gpus}<span className="muted">/{h.num_gpus_total}</span> GPU</div>
                  </div>
                  <div>
                    <div className="muted small" style={{ fontSize: 11 }}>Hourly</div>
                    <div style={{ fontWeight: 700, color: "var(--green)" }}>{fmtMoney(h.dph_total)}/hr</div>
                  </div>
                  <div>
                    <div className="muted small" style={{ fontSize: 11 }}>Uptime</div>
                    <div style={{ fontWeight: 700, color: h.uptime < 95 ? "var(--red)" : "var(--text)" }}>{h.uptime}%</div>
                  </div>
                  <div>
                    <div className="muted small" style={{ fontSize: 11 }}>Reliability</div>
                    <div style={{ fontWeight: 700 }}>{Math.round((h.reliability ?? 0.95) * 100)}%</div>
                  </div>
                </div>
              </div>

              <div className="row between wrap mt" style={{ gap: 14 }}>
                <div className="row wrap" style={{ gap: 18, color: "var(--text-dim)", fontSize: 12.5 }}>
                  <span>{h.ram} GB RAM</span>
                  <span>{h.cpu_cores} cores</span>
                  <span>{h.total_storage} TB storage</span>
                  <span className="row gap-sm" style={{ alignItems: "flex-end" }}>
                    12-mo earnings <Spark values={h.months} height={30} />
                  </span>
                </div>
                <div className="row" style={{ gap: 8 }}>
                  <button className={`btn btn-sm ${h.verified ? "btn-ghost" : "btn-green"}`} onClick={() => toggleVerified(h.id)}>
                    <CheckCircle2 size={13} /> {h.verified ? "Unverify" : "Verify"}
                  </button>
                  <button className={`btn btn-sm ${flagged[h.id] ? "btn-green" : "btn-danger"}`} onClick={() => toggleFlag(h.id)}>
                    {flagged[h.id] ? <Server size={13} /> : <AlertTriangle size={13} />}
                    {flagged[h.id] ? "Clear flag" : "Flag"}
                  </button>
                </div>
              </div>
            </div>
          );
        })}
        {rows.length === 0 && (
          <div className="empty"><ServerOff size={40} />No machines registered.</div>
        )}
      </div>
    </div>
  );
}