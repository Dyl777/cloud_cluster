import { useState } from "react";
import { UserCheck, UserX, Users } from "lucide-react";
import { api, fmtMoney, timeAgo, fmtDate } from "../../data/mock";
import { Badge, PageHead, StatCard } from "../../components/ui";
import AdminNav from "./AdminNav";

const initials = (name) =>
  name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

function statusBadge(status) {
  const tone = status === "active" ? "green" : status === "suspended" ? "red" : "amber";
  return <Badge tone={tone} dot>{status}</Badge>;
}

export default function AdminUsers() {
  const [rows, setRows] = useState(api.users);
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");
  const [role, setRole] = useState("all");

  const totals = {
    total: rows.length,
    active: rows.filter((u) => u.status === "active").length,
    suspended: rows.filter((u) => u.status === "suspended").length,
    spend: rows.reduce((s, u) => s + u.spent, 0),
  };

  const filtered = rows.filter((u) => {
    const matchQ =
      !q ||
      u.email.toLowerCase().includes(q.toLowerCase()) ||
      u.name.toLowerCase().includes(q.toLowerCase());
    const matchStatus = status === "all" || u.status === status;
    const matchRole = role === "all" || u.role === role;
    return matchQ && matchStatus && matchRole;
  });

  function toggleStatus(id) {
    setRows((prev) =>
      prev.map((u) => (u.id === id ? { ...u, status: u.status === "suspended" ? "active" : "suspended" } : u)),
    );
  }

  function setUserRole(id, next) {
    setRows((prev) => prev.map((u) => (u.id === id ? { ...u, role: next } : u)));
  }

  return (
    <div>
      <PageHead title="Users" sub="Manage registrations, roles and account state across the platform." />
      <AdminNav />

      <div className="grid cols-4">
        <StatCard label="Total users" value={totals.total} icon={Users} tone="var(--blue)" />
        <StatCard label="Active" value={totals.active} icon={UserCheck} tone="var(--green)" />
        <StatCard label="Suspended" value={totals.suspended} icon={UserX} tone="var(--red)" />
        <StatCard label="Lifetime spend" value={fmtMoney(totals.spend)} icon={Users} tone="var(--amber)" />
      </div>

      <div className="filters mt">
        <label className="field grow">
          <input className="input" placeholder="Search by name or email…" value={q} onChange={(e) => setQ(e.target.value)} />
        </label>
        <label className="field">Status
          <select className="select" value={status} onChange={(e) => setStatus(e.target.value)}>
            {["all", "active", "suspended", "pending"].map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
        </label>
        <label className="field">Role
          <select className="select" value={role} onChange={(e) => setRole(e.target.value)}>
            {["all", "user", "host", "admin", "superadmin"].map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
        </label>
      </div>

      <div className="card mt" style={{ padding: 0 }}>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>User</th>
                <th>Role</th>
                <th>Status</th>
                <th style={{ textAlign: "right" }}>Balance</th>
                <th style={{ textAlign: "right" }}>Spent</th>
                <th style={{ textAlign: "right" }}>GPU-hours</th>
                <th>Joined</th>
                <th>Last seen</th>
                <th style={{ textAlign: "right" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((u) => (
                <tr key={u.id}>
                  <td>
                    <div className="row gap-sm">
                      <span className="avatar" style={{ width: 26, height: 26, fontSize: 11 }}>{initials(u.name)}</span>
                      <span>
                        {u.name}
                        <div className="small dim mono">{u.email}</div>
                      </span>
                    </div>
                  </td>
                  <td>
                    <select
                      className="select"
                      style={{ width: 120, padding: "4px 8px", fontSize: 12 }}
                      value={u.role}
                      onChange={(e) => setUserRole(u.id, e.target.value)}
                      disabled={u.role === "superadmin"}
                    >
                      {u.role === "superadmin"
                        ? <option value="superadmin">superadmin</option>
                        : ["user", "host", "admin"].map((r) => <option key={r} value={r}>{r}</option>)}
                    </select>
                  </td>
                  <td>{statusBadge(u.status)}</td>
                  <td style={{ textAlign: "right", fontWeight: 600 }}>{fmtMoney(u.balance)}</td>
                  <td style={{ textAlign: "right" }}>{fmtMoney(u.spent)}</td>
                  <td style={{ textAlign: "right" }}>{u.gpu_hours.toLocaleString()}</td>
                  <td className="small dim">{fmtDate(u.created)}</td>
                  <td className="small dim">{timeAgo(u.last_seen)}</td>
                  <td style={{ textAlign: "right" }}>
                    <button
                      className={`btn btn-sm ${u.status === "suspended" ? "btn-green" : "btn-danger"}`}
                      disabled={u.role === "superadmin"}
                      onClick={() => toggleStatus(u.id)}
                    >
                      {u.status === "suspended" ? "Activate" : "Suspend"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}