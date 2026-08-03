import { useEffect, useRef } from "react";
import { Link, useLocation } from "react-router-dom";
import {
  Box, CreditCard, Cpu, Gauge, LayoutDashboard, LogOut, Moon, Server, Settings,
  ShoppingCart, Sun, Wallet, Zap,
} from "lucide-react";
import { useStore } from "../store";
import { useAuth } from "../auth";
import { useTheme } from "../theme";
import { fmtMoney } from "../data/mock";

export function NavLink({ to, icon: Icon, children }) {
  const { pathname } = useLocation();
  const active = pathname === to || (to !== "/" && pathname.startsWith(to));
  return (
    <Link to={to} className={active ? "active" : ""}>
      <Icon size={15} />
      {children}
    </Link>
  );
}

export function Topbar() {
  const { account, user } = useStore();
  const { logout } = useAuth();
  const { theme, toggle } = useTheme();
  const initials = user?.name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase() || "VA";

  return (
    <header className="topbar">
      <Link to="/" className="logo">
        <span className="logo-mark">v</span>
        vast<span style={{ color: "var(--text-mute)", fontWeight: 600 }}>.ai</span>
      </Link>
      <nav className="nav">
        <NavLink to="/" icon={ShoppingCart}>Marketplace</NavLink>
        <NavLink to="/instances" icon={Box}>Instances</NavLink>
        <NavLink to="/hosts" icon={Server}>Hosts</NavLink>
        <NavLink to="/billing" icon={CreditCard}>Billing</NavLink>
        <NavLink to="/settings" icon={Settings}>Settings</NavLink>
      </nav>
      <div className="topbar-right">
        <button className="btn btn-ghost btn-sm" onClick={toggle} title={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}>
          {theme === "dark" ? <Sun size={15} /> : <Moon size={15} />}
        </button>
        <Link to="/billing" className="balance-chip" title="Account balance">
          <Wallet size={14} />
          <span className="label">balance</span>
          {fmtMoney(account.balance)}
        </Link>
        <div className="avatar" title={user?.name}>{initials}</div>
        <button className="btn btn-ghost btn-sm" onClick={logout} title="Sign out">
          <LogOut size={14} />
        </button>
      </div>
    </header>
  );
}

export function Toast() {
  const { toast } = useStore();
  if (!toast) return null;
  return (
    <div style={{
      position: "fixed", bottom: 22, left: "50%", transform: "translateX(-50%)",
      zIndex: 100, background: "var(--bg-soft)", border: "1px solid var(--green)",
      color: "var(--text)", padding: "10px 18px", borderRadius: 8, fontSize: 13,
      boxShadow: "var(--shadow-lg)", display: "flex", gap: 8, alignItems: "center",
    }}>
      <Zap size={15} color="var(--green)" />
      {toast}
    </div>
  );
}

export function StatCard({ label, value, sub, icon: Icon, tone }) {
  return (
    <div className="card stat">
      <span className="stat-label">{label}</span>
      <div className="row" style={{ gap: 8, margin: 0 }}>
        <span className="stat-value">{value}</span>
        {Icon && <Icon size={18} color={tone || "var(--accent)"} />}
      </div>
      {sub && <span className="stat-sub">{sub}</span>}
    </div>
  );
}

export function Badge({ children, tone = "neutral", dot }) {
  return (
    <span className={`badge badge-${tone}`}>
      {dot && <span className="dot" />}
      {children}
    </span>
  );
}

export function StatusBadge({ status }) {
  const map = {
    running: { tone: "green", label: "Running" },
    pending: { tone: "amber", label: "Provisioning" },
    stopped: { tone: "neutral", label: "Stopped" },
    off: { tone: "red", label: "Offline" },
    paused: { tone: "amber", label: "Paused" },
    error: { tone: "red", label: "Error" },
  };
  const s = map[status] || { tone: "neutral", label: status };
  return <Badge tone={s.tone} dot>{s.label}</Badge>;
}

export function UtilBar({ value, title }) {
  const hot = value > 75;
  return (
    <div className="util-bar" title={title}>
      <div className={`fill ${hot ? "hot" : ""}`} style={{ width: `${Math.min(100, value)}%` }} />
    </div>
  );
}

export function Spark({ values, height = 40, color = "var(--accent)" }) {
  const max = Math.max(...values, 1);
  const bars = useRef(values).current;
  return (
    <div className="mini-chart" style={{ height }}>
      {bars.map((v, i) => (
        <div key={i} className="bar" style={{ height: `${Math.max(8, (v / max) * 100)}%`, background: color }} title={v} />
      ))}
    </div>
  );
}

export function PageHead({ title, sub, children }) {
  return (
    <div className="page-head">
      <div>
        <h1>{title}</h1>
        {sub && <div className="sub">{sub}</div>}
      </div>
      {children && <div className="row wrap">{children}</div>}
    </div>
  );
}

export function KeyValue({ k, v }) {
  return (
    <div className="kv" style={{ display: "contents" }}>
      <span className="k">{k}</span>
      <span className="v">{v}</span>
    </div>
  );
}

export function CopyButton({ text, label = "Copy" }) {
  async function copy() {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      /* ignore */
    }
  }
  return (
    <button className="btn btn-sm btn-ghost" onClick={copy} title="Copy to clipboard">
      {label}
    </button>
  );
}

export function useInterval(cb, ms) {
  const saved = useRef(cb);
  useEffect(() => {
    saved.current = cb;
  }, [cb]);
  useEffect(() => {
    if (!ms) return;
    const t = setInterval(() => saved.current(), ms);
    return () => clearInterval(t);
  }, [ms]);
}

// eslint-disable-next-line react-refresh/only-export-components
export { Gauge, Cpu, LayoutDashboard };
