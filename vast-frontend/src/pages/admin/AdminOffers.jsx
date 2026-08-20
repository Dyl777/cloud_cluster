import { useState } from "react";
import { RotateCcw, ShieldCheck, ShieldOff, Trash2, ShoppingCart, Gauge } from "lucide-react";
import { useStore } from "../../store";
import { fmtMoney } from "../../data/mock";
import { Badge, PageHead, StatCard } from "../../components/ui";
import AdminNav from "./AdminNav";

export default function AdminOffers() {
  const { offers: initial } = useStore();
  const [rows, setRows] = useState(initial);
  const [filter, setFilter] = useState("all");

  const totalGpus = rows.reduce((s, o) => s + o.num_gpus, 0);
  const rented = rows.filter((o) => o.rented).length;
  const verified = rows.filter((o) => o.verified).length;
  const avgPrice = totalGpus ? rows.reduce((s, o) => s + o.dph_total, 0) / rows.length : 0;

  const shown = rows.filter((o) => filter === "all" || (filter === "rented" && o.rented) || (filter === "free" && !o.rented));

  function toggleVerified(id) {
    setRows((prev) => prev.map((o) => (o.id === id ? { ...o, verified: !o.verified } : o)));
  }

  function delist(id) {
    setRows((prev) => prev.filter((o) => o.id !== id));
  }

  return (
    <div>
      <PageHead title="Offers" sub="Moderate the GPU marketplace: verify sellers, delist bad offers, watch utilization." />
      <AdminNav />

      <div className="grid cols-4">
        <StatCard label="Listed offers" value={rows.length} icon={ShoppingCart} tone="var(--blue)" />
        <StatCard label="Total GPUs" value={totalGpus.toLocaleString()} icon={Gauge} tone="var(--purple)" />
        <StatCard label="Rented" value={rented} sub={`${rows.length - rented} free`} icon={Gauge} tone="var(--green)" />
        <StatCard label="Avg $/hr" value={fmtMoney(avgPrice)} sub={`${verified}/${rows.length} verified`} icon={ShieldCheck} tone="var(--amber)" />
      </div>

      <div className="filters mt">
        <label className="field">Show
          <select className="select" value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="all">All offers</option>
            <option value="rented">Rented</option>
            <option value="free">Available</option>
          </select>
        </label>
      </div>

      <div className="card mt" style={{ padding: 0 }}>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Machine</th>
                <th>GPU</th>
                <th>GPUs</th>
                <th>Region</th>
                <th style={{ textAlign: "right" }}>Reliability</th>
                <th style={{ textAlign: "right" }}>$ / hr</th>
                <th>Status</th>
                <th style={{ textAlign: "right" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {shown.map((o) => (
                <tr key={o.id}>
                  <td>
                    <div className="mono">#{o.host_id}</div>
                    <div className="small dim mono">{o.machine_id}</div>
                  </td>
                  <td className="mono">{o.gpu_name}</td>
                  <td>{o.num_gpus}</td>
                  <td className="small">{o.display_location || o.location}</td>
                  <td style={{ textAlign: "right" }}>{Math.round((o.reliability2 ?? 0.95) * 100)}%</td>
                  <td style={{ textAlign: "right", fontWeight: 700, color: "var(--green)" }}>{fmtMoney(o.dph_total)}</td>
                  <td>
                    <div className="row" style={{ gap: 6 }}>
                      {o.verified ? <Badge tone="green" dot>verified</Badge> : <Badge tone="neutral">unverified</Badge>}
                      {o.rented ? <Badge tone="blue">rented</Badge> : <Badge tone="green">free</Badge>}
                      {o.datacenter && <Badge tone="purple">dc</Badge>}
                    </div>
                  </td>
                  <td style={{ textAlign: "right", whiteSpace: "nowrap" }}>
                    <button className="btn btn-sm btn-ghost" title={o.verified ? "Unverify" : "Verify"} onClick={() => toggleVerified(o.id)}>
                      {o.verified ? <ShieldOff size={13} /> : <ShieldCheck size={13} />}
                    </button>
                    <button className="btn btn-sm btn-ghost" title="Delist" onClick={() => delist(o.id)}>
                      <Trash2 size={13} color="var(--red)" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      <p className="small dim mt-sm">
        <RotateCcw size={12} /> Changes are session-local; refresh to reload the mock marketplace.
      </p>
    </div>
  );
}