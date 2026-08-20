import { useState } from "react";
import { ArrowDownLeft, ArrowUpRight, CheckCircle2, DollarSign, Wallet, Landmark } from "lucide-react";
import { useStore } from "../../store";
import { api, fmtDate, fmtMoney } from "../../data/mock";
import { Badge, PageHead, StatCard } from "../../components/ui";
import AdminNav from "./AdminNav";

export default function AdminPayments() {
  const { transactions } = useStore();
  const [paid, setPaid] = useState({});
  const [credits, setCredits] = useState({});

  const users = api.users;
  const revenue30d = api.activity.reduce((s, d) => s + d.revenue, 0);
  const gross = users.reduce((s, u) => s + u.spent + u.balance, 0);
  const outstanding = users.reduce((s, u) => s + u.balance, 0);
  const payouts = users.reduce((s, u) => s + u.earned, 0);

  function payOut(id) {
    setPaid((prev) => ({ ...prev, [id]: true }));
  }

  function credit(id) {
    setCredits((prev) => ({ ...prev, [id]: (prev[id] || 0) + 25 }));
  }

  const ledgerTotal = transactions.reduce((s, t) => s + Math.abs(t.amount), 0);

  return (
    <div>
      <PageHead title="Payments" sub="Platform revenue, balances, payouts and the full transaction ledger." />
      <AdminNav />

      <div className="grid cols-4">
        <StatCard label="Revenue (30d)" value={fmtMoney(revenue30d)} icon={DollarSign} tone="var(--green)" />
        <StatCard label="Gross volume" value={fmtMoney(gross)} icon={ArrowUpRight} tone="var(--blue)" />
        <StatCard label="Outstanding balances" value={fmtMoney(outstanding)} icon={Wallet} tone="var(--amber)" />
        <StatCard label="Host payouts due" value={fmtMoney(payouts)} icon={Landmark} tone="var(--purple)" />
      </div>

      <div className="grid cols-2 mt">
        <div className="card">
          <div className="card-title"><h3>Platform ledger</h3><Wallet size={16} color="var(--accent)" /></div>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Date</th><th>Description</th><th style={{ textAlign: "right" }}>Amount</th><th>Status</th></tr></thead>
              <tbody>
                {transactions.map((t) => (
                  <tr key={t.id}>
                    <td className="small dim">{fmtDate(t.time)}</td>
                    <td>
                      {t.label}
                      <div className="small dim mono">#{t.id}</div>
                    </td>
                    <td style={{ textAlign: "right", fontWeight: 700, color: t.amount > 0 ? "var(--green)" : "var(--text)" }}>
                      {t.amount > 0 ? "+" : ""}{fmtMoney(t.amount)}
                    </td>
                    <td><Badge tone={t.status === "completed" ? "green" : "amber"} dot>{t.status}</Badge></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="small dim mt-sm">Ledger volume (this session): {fmtMoney(ledgerTotal)}</div>
        </div>

        <div className="card">
          <div className="card-title"><h3>Customer balances & payouts</h3><Landmark size={16} color="var(--purple)" /></div>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Customer</th><th style={{ textAlign: "right" }}>Balance</th><th style={{ textAlign: "right" }}>Earned</th><th style={{ textAlign: "right" }}>Actions</th></tr></thead>
              <tbody>
                {users.slice(0, 8).map((u) => (
                  <tr key={u.id}>
                    <td>
                      <div style={{ fontWeight: 600 }}>{u.name}</div>
                      <div className="small dim mono">{u.email}</div>
                    </td>
                    <td style={{ textAlign: "right", fontWeight: 700 }}>
                      {fmtMoney(u.balance + (credits[u.id] || 0))}
                    </td>
                    <td style={{ textAlign: "right" }} className={u.earned ? "mono" : "small dim"}>
                      {fmtMoney(u.earned)}
                    </td>
                    <td style={{ textAlign: "right", whiteSpace: "nowrap" }}>
                      <button className="btn btn-sm btn-ghost" title="Manual credit $25" onClick={() => credit(u.id)}>
                        <ArrowDownLeft size={13} color="var(--green)" />
                      </button>
                      <button
                        className="btn btn-sm btn-ghost"
                        disabled={paid[u.id] || u.earned === 0}
                        title={paid[u.id] ? "Payout sent" : "Send payout"}
                        onClick={() => payOut(u.id)}
                      >
                        {paid[u.id] ? <CheckCircle2 size={13} color="var(--green)" /> : <ArrowUpRight size={13} />}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}