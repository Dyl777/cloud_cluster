import { useState } from "react";
import { Box, Power, PowerOff, Hourglass, Trash2 } from "lucide-react";
import { useStore } from "../../store";
import { fmtMoney, fmtDate, fmtUptime } from "../../data/mock";
import { PageHead, StatCard, StatusBadge } from "../../components/ui";
import AdminNav from "./AdminNav";

export default function AdminInstances() {
  const { instances, destroyInstance, notify } = useStore();
  const [filter, setFilter] = useState("all");

  const running = instances.filter((i) => i.actual_status === "running");
  const pending = instances.filter((i) => i.actual_status === "pending");
  const burn = instances.reduce((s, i) => s + i.dph_total, 0);

  const shown = instances.filter((i) => filter === "all" || (filter === "running" && i.actual_status === "running") || (filter === "pending" && i.actual_status === "pending") || (filter === "stopped" && (i.actual_status === "stopped" || i.actual_status === "off")));

  function kill(id) {
    destroyInstance(id);
    notify(`Instance #${id} destroyed by admin.`);
  }

  return (
    <div>
      <PageHead title="Instances" sub="Every instance running across the platform, with lifecycle control." />
      <AdminNav />

      <div className="grid cols-4">
        <StatCard label="Total" value={instances.length} icon={Box} tone="var(--blue)" />
        <StatCard label="Running" value={running.length} icon={Power} tone="var(--green)" />
        <StatCard label="Provisioning" value={pending.length} icon={Hourglass} tone="var(--amber)" />
        <StatCard label="Burn rate" value={fmtMoney(burn)} sub="/hr across all instances" icon={Box} tone="var(--purple)" />
      </div>

      <div className="filters mt">
        <label className="field">Show
          <select className="select" value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="all">All instances</option>
            <option value="running">Running</option>
            <option value="pending">Provisioning</option>
            <option value="stopped">Stopped / off</option>
          </select>
        </label>
      </div>

      <div className="card mt" style={{ padding: 0 }}>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Instance</th>
                <th>GPU</th>
                <th>Status</th>
                <th style={{ textAlign: "right" }}>Util</th>
                <th style={{ textAlign: "right" }}>Uptime</th>
                <th style={{ textAlign: "right" }}>$ / hr</th>
                <th>Region</th>
                <th>Started</th>
                <th style={{ textAlign: "right" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {shown.map((i) => (
                <tr key={i.id}>
                  <td>
                    <div style={{ fontWeight: 600 }}>{i.label}</div>
                    <div className="small dim mono">#{i.id} · {i.image_name}</div>
                  </td>
                  <td className="mono">{i.gpu_name} ×{i.num_gpus}</td>
                  <td><StatusBadge status={i.actual_status} /></td>
                  <td style={{ textAlign: "right" }}>
                    <div className="row" style={{ justifyContent: "flex-end" }}>
                      <div className="util-bar"><div className={`fill ${i.gpu_util > 75 ? "hot" : ""}`} style={{ width: `${i.gpu_util}%` }} /></div>
                      <span className="small dim">{i.gpu_util}%</span>
                    </div>
                  </td>
                  <td style={{ textAlign: "right" }} className="mono">{i.actual_status === "running" ? fmtUptime(i.duration) : "—"}</td>
                  <td style={{ textAlign: "right", fontWeight: 700, color: "var(--green)" }}>{fmtMoney(i.dph_total)}</td>
                  <td className="small">{i.display_location || i.location}</td>
                  <td className="small dim">{fmtDate(i.start_date)}</td>
                  <td style={{ textAlign: "right" }}>
                    <button className="btn btn-sm btn-danger" title="Destroy instance" onClick={() => kill(i.id)}>
                      <Trash2 size={13} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {instances.length === 0 && (
          <div className="empty"><PowerOff size={40} />No instances on the platform.</div>
        )}
      </div>
    </div>
  );
}