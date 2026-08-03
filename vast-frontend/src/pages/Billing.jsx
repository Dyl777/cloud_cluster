import { useState } from "react";
import { ArrowDownLeft, ArrowUpRight, CreditCard, DollarSign, Plus, Wallet } from "lucide-react";
import { useStore } from "../store";
import { fmtDate, fmtMoney } from "../data/mock";
import { Badge, PageHead, StatCard } from "../components/ui";

export default function Billing() {
  const { account, transactions, topUp } = useStore();
  const [amount, setAmount] = useState(25);

  const spent = transactions.filter((t) => t.amount < 0).reduce((s, t) => s - t.amount, 0);

  return (
    <div>
      <PageHead title="Billing" sub="Manage credits, view charges, and track your usage spend." />

      <div className="grid cols-4">
        <StatCard label="Balance" value={fmtMoney(account.balance)} sub="available credit" icon={Wallet} tone="var(--green)" />
        <StatCard label="Spent (30d)" value={fmtMoney(spent)} icon={ArrowUpRight} tone="var(--red)" />
        <StatCard label="Earned (hosting)" value={fmtMoney(account.total_earned)} icon={ArrowDownLeft} tone="var(--accent-2)" />
        <StatCard label="Reserved" value={fmtMoney(account.reserved)} sub="running instances" icon={DollarSign} tone="var(--amber)" />
      </div>

      <div className="grid cols-2" style={{ marginTop: 16 }}>
        <div className="card">
          <div className="card-title"><h3>Add credits</h3><CreditCard size={16} color="var(--accent)" /></div>
          <div className="row" style={{ gap: 10 }}>
            {[10, 25, 50, 100, 500].map((a) => (
              <button key={a} className={`btn btn-sm ${amount === a ? "btn-primary" : ""}`} onClick={() => setAmount(a)}>{a}</button>
            ))}
            <input className="input" style={{ width: 110 }} type="number" min="1" value={amount} onChange={(e) => setAmount(+e.target.value)} />
          </div>
          <button className="btn btn-green mt" style={{ width: "100%", justifyContent: "center" }} onClick={() => topUp(amount)}>
            <Plus size={15} /> Add {fmtMoney(amount)} to balance
          </button>
          <div className="small dim mt">Simulated checkout — no real payment is processed.</div>
        </div>

        <div className="card">
          <div className="card-title"><h3>Account</h3><Badge tone="blue">{account.plan}</Badge></div>
          <div className="kv">
            <div style={{ display: "contents" }}><span className="k">Member since</span><span className="v">{new Date(account.created).toLocaleDateString(undefined, { month: "long", year: "numeric" })}</span></div>
            <div style={{ display: "contents" }}><span className="k">Pending charges</span><span className="v">{fmtMoney(account.pending)}</span></div>
            <div style={{ display: "contents" }}><span className="k">Credits</span><span className="v">{account.credits.toFixed(2)}</span></div>
            <div style={{ display: "contents" }}><span className="k">Default currency</span><span className="v">USD</span></div>
          </div>
        </div>
      </div>

      <div className="card mt" style={{ padding: 0 }}>
        <div style={{ padding: "16px 18px", borderBottom: "1px solid var(--border-soft)" }}>
          <h3>Transactions</h3>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Date</th><th>Description</th><th>Type</th><th style={{ textAlign: "right" }}>Amount</th><th>Status</th></tr></thead>
            <tbody>
              {transactions.map((t) => (
                <tr key={t.id}>
                  <td className="small dim">{fmtDate(t.time)}</td>
                  <td>
                    {t.label}
                    <div className="small dim mono">#{t.id}</div>
                  </td>
                  <td>
                    <Badge tone={t.amount > 0 ? "green" : "neutral"}>
                      {t.amount > 0 ? <ArrowDownLeft size={11} /> : <ArrowUpRight size={11} />} {t.type}
                    </Badge>
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
      </div>
    </div>
  );
}
