import { useMemo, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { Box, Clock, Cpu, DollarSign, Play, Pause, Power, Trash2 } from "lucide-react";
import { useStore } from "../store";
import { fmtMoney, fmtUptime } from "../data/mock";
import { Badge, PageHead, StatCard, StatusBadge, UtilBar } from "../components/ui";

export default function Instances() {
  const { instances, setInstanceStatus, destroyInstance } = useStore();
  const location = useLocation();
  const highlight = location.state?.highlight;
  const [tab, setTab] = useState("all");

  const stats = useMemo(() => {
    const running = instances.filter((i) => i.actual_status === "running");
    const cost = instances.filter((i) => i.actual_status !== "off").reduce((s, i) => s + i.dph_total, 0);
    return { total: instances.length, running: running.length, cost, gpus: running.reduce((s, i) => s + i.num_gpus, 0) };
  }, [instances]);

  const filtered = tab === "all" ? instances : instances.filter((i) => i.actual_status === tab);

  return (
    <div>
      <PageHead title="My Instances" sub={`${stats.running} running · ${stats.cost.toFixed(2)} $/hr burn rate`}>
        <Link to="/" className="btn btn-primary"><Box size={15} /> Rent more</Link>
      </PageHead>

      <div className="grid cols-4">
        <StatCard label="Instances" value={stats.total} icon={Box} />
        <StatCard label="Running" value={stats.running} icon={Play} tone="var(--green)" />
        <StatCard label="Total GPUs" value={stats.gpus} icon={Cpu} tone="var(--accent-2)" />
        <StatCard label="Burn rate" value={fmtMoney(stats.cost)} sub="per hour" icon={DollarSign} tone="var(--amber)" />
      </div>

      <div className="tabs mt">
        {["all", "running", "pending", "stopped"].map((t) => (
          <button key={t} className={`tab ${tab === t ? "active" : ""}`} onClick={() => setTab(t)}>
            {t[0].toUpperCase() + t.slice(1)}
            <span className="muted small"> ({t === "all" ? instances.length : instances.filter((i) => i.actual_status === t).length})</span>
          </button>
        ))}
      </div>

      <div className="card" style={{ padding: 0 }}>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Instance</th>
                <th>Status</th>
                <th>GPU</th>
                <th>GPU util</th>
                <th>CPU</th>
                <th>RAM</th>
                <th>Location</th>
                <th>Uptime</th>
                <th>$/hr</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((i) => {
                const running = i.actual_status === "running";
                return (
                  <tr key={i.id} className="clickable" style={{ background: highlight === i.id ? "rgba(9,105,218,0.06)" : undefined }}>
                    <td>
                      <Link to={`/instances/${i.id}`} style={{ fontWeight: 600 }}>{i.label}</Link>
                      <div className="small dim mono">#{i.id}</div>
                    </td>
                    <td><StatusBadge status={i.actual_status} /></td>
                    <td>
                      {i.gpu_name}
                      <div className="small dim">{i.num_gpus}×</div>
                    </td>
                    <td><div className="row" style={{ gap: 8 }}><UtilBar value={i.gpu_util} /><span className="small dim">{i.gpu_util}%</span></div></td>
                    <td><div className="row" style={{ gap: 8 }}><UtilBar value={i.cpu_util} /><span className="small dim">{i.cpu_util}%</span></div></td>
                    <td><div className="row" style={{ gap: 8 }}><UtilBar value={i.mem_util} /><span className="small dim">{i.mem_util}%</span></div></td>
                    <td><div className="small">{i.display_location}</div></td>
                    <td><div className="row gap-sm small dim"><Clock size={12} />{running ? fmtUptime(i.duration) : "—"}</div></td>
                    <td><span style={{ color: "var(--green)", fontWeight: 600 }}>{fmtMoney(i.dph_total)}</span></td>
                    <td>
                      <div className="row" style={{ gap: 6 }} onClick={(e) => e.stopPropagation()}>
                        {running && <InstanceAction icon={<Pause size={13} />} label="Pause" onClick={() => setInstanceStatus(i.id, "stopped")} title="Pause (stop billing GPU)" />}
                        {running && <InstanceAction icon={<Power size={13} />} label="Stop" onClick={() => setInstanceStatus(i.id, "stopped")} />}
                        {i.actual_status === "stopped" && <InstanceAction icon={<Play size={13} />} label="Start" onClick={() => setInstanceStatus(i.id, "running")} />}
                        <InstanceAction icon={<Trash2 size={13} />} label="Destroy" danger onClick={() => destroyInstance(i.id)} />
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          {filtered.length === 0 && (
            <div className="empty">
              <Box size={40} />
              <div>No instances here yet.</div>
              <Link to="/" className="btn btn-primary mt-sm">Browse the marketplace</Link>
            </div>
          )}
        </div>
      </div>

      <div className="card mt">
        <div className="card-title"><h3>Storage volumes</h3><Badge tone="neutral">1 volume</Badge></div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Label</th><th>Size</th><th>Instance</th><th>Location</th><th>Created</th></tr></thead>
            <tbody>
              <tr>
                <td><span className="mono">data-lab</span></td>
                <td>200 GB</td>
                <td className="dim">not attached</td>
                <td>US - California</td>
                <td className="small dim">2 weeks ago</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function InstanceAction({ icon, label, onClick, title, danger }) {
  return (
    <button className={`btn btn-sm ${danger ? "btn-danger" : "btn-ghost"}`} onClick={onClick} title={title}>
      {icon}
      <span className="small">{label}</span>
    </button>
  );
}
