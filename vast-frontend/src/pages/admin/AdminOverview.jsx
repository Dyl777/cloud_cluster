import { useStore } from "../../store";
import { api, fmtMoney, timeAgo } from "../../data/mock";
import { Badge, PageHead, Spark, StatCard } from "../../components/ui";
import { Activity, Box, Cpu, DollarSign, UserPlus, Users, ShieldCheck } from "lucide-react";
import AdminNav from "./AdminNav";

export default function AdminOverview() {
  const { offers, instances, hosts, transactions } = useStore();
  const users = api.users;
  const activity = api.activity;

  const active = (s) => s === "running" || s === "pending";
  const runningInstances = instances.filter((i) => i.actual_status === "running");
  const runningGPUs = runningInstances.reduce((s, i) => s + i.num_gpus, 0);
  const marketGPUs = offers.reduce((s, o) => s + o.num_gpus, 0);
  const rentedOffers = offers.filter((o) => o.rented).length;
  const revenue30d = activity.reduce((s, d) => s + d.revenue, 0);
  const totalGpuHours = activity.reduce((s, d) => s + d.gpu_hours, 0);
  const credits = transactions.filter((t) => t.amount > 0 && t.status === "completed").reduce((s, t) => s + t.amount, 0);

  const gpuPool = {};
  offers.forEach((o) => {
    gpuPool[o.gpu_name] = (gpuPool[o.gpu_name] || 0) + o.num_gpus;
  });
  const gpuRows = Object.entries(gpuPool).sort((a, b) => b[1] - a[1]).slice(0, 8);

  const regions = {};
  offers.forEach((o) => {
    const k = o.country || o.region;
    regions[k] = (regions[k] || 0) + o.num_gpus;
  });
  const regionRows = Object.entries(regions).sort((a, b) => b[1] - a[1]);

  const revenueSeries = activity.map((d) => Math.round(d.revenue));
  const signupSeries = activity.map((d) => d.signups);

  return (
    <div>
      <PageHead title="Admin" sub={`Platform-wide control room · ${users.length} registered users`}>
        <Badge tone="purple" dot>superadmin</Badge>
      </PageHead>
      <AdminNav />

      <div className="grid cols-4">
        <StatCard label="Users" value={users.length} sub={`${users.filter((u) => u.status === "active").length} active`} icon={Users} tone="var(--blue)" />
        <StatCard label="Market GPUs" value={marketGPUs.toLocaleString()} sub={`${rentedOffers}/${offers.length} offers rented`} icon={Cpu} tone="var(--purple)" />
        <StatCard label="Running instances" value={instances.filter(active).length} sub={`${runningGPUs} GPUs in use`} icon={Box} tone="var(--green)" />
        <StatCard label="Revenue (30d)" value={fmtMoney(revenue30d)} sub={`${totalGpuHours.toLocaleString()} GPU-hours`} icon={DollarSign} tone="var(--amber)" />
      </div>

      <div className="grid cols-2 mt">
        <div className="card">
          <div className="card-title"><h3>Revenue — last 28 days</h3><Activity size={16} color="var(--green)" /></div>
          <Spark values={revenueSeries} height={110} color="var(--green)" />
        </div>
        <div className="card">
          <div className="card-title"><h3>New signups — last 28 days</h3><UserPlus size={16} color="var(--blue)" /></div>
          <Spark values={signupSeries} height={110} color="var(--blue)" />
        </div>
      </div>

      <div className="grid cols-2 mt">
        <div className="card">
          <div className="card-title"><h3>GPU pool by model</h3><Cpu size={16} color="var(--accent-2)" /></div>
          <table className="admin-table">
            <tbody>
              {gpuRows.map(([name, n]) => {
                const pct = Math.round((n / marketGPUs) * 100);
                return (
                  <tr key={name}>
                    <td className="mono">{name}</td>
                    <td style={{ width: 40, textAlign: "right" }}>{n}</td>
                    <td style={{ width: "40%" }}><div className="util-bar"><div className="fill" style={{ width: `${pct}%` }} /></div></td>
                    <td className="small dim" style={{ width: 40 }}>{pct}%</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div className="card">
          <div className="card-title"><h3>Capacity by country</h3><ShieldCheck size={16} color="var(--purple)" /></div>
          <table className="admin-table">
            <tbody>
              {regionRows.map(([region, n]) => {
                const pct = Math.round((n / marketGPUs) * 100);
                return (
                  <tr key={region}>
                    <td>{region}</td>
                    <td style={{ width: 40, textAlign: "right" }}>{n}</td>
                    <td style={{ width: "40%" }}><div className="util-bar"><div className="fill" style={{ width: `${pct}%`, background: "var(--purple)" }} /></div></td>
                    <td className="small dim" style={{ width: 40 }}>{pct}%</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      <div className="grid cols-2 mt">
        <div className="card">
          <div className="card-title"><h3>Recent signups</h3><Users size={16} color="var(--accent)" /></div>
          <div className="kv">
            {users.slice(0, 5).map((u) => (
              <div key={u.id} style={{ display: "contents" }}>
                <span className="k">{u.name}</span>
                <span className="v small dim">{u.email} · {timeAgo(u.created)}</span>
              </div>
            ))}
          </div>
        </div>
        <div className="card">
          <div className="card-title"><h3>Ledger health</h3><Activity size={16} color="var(--amber)" /></div>
          <div className="kv">
            <div style={{ display: "contents" }}><span className="k">Completed credits</span><span className="v">{fmtMoney(credits)}</span></div>
            <div style={{ display: "contents" }}><span className="k">Total GPU-hours (28d)</span><span className="v">{totalGpuHours.toLocaleString()}</span></div>
            <div style={{ display: "contents" }}><span className="k">Host machines</span><span className="v">{hosts.length}</span></div>
            <div style={{ display: "contents" }}><span className="k">All-time spend</span><span className="v">{fmtMoney(transactions.reduce((s, t) => s + Math.abs(t.amount), 0))}</span></div>
          </div>
        </div>
      </div>
    </div>
  );
}