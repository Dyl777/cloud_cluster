import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ShieldCheck, Cpu, Gauge, Wallet, Server } from "lucide-react";
import { useAuth } from "../auth";
import { useStore } from "../store";

function AuthShell({ title, subtitle, children, side }) {
  return (
    <div className="auth-wrap">
      <aside className="auth-side">
        <Link to="/" className="logo">
          <span className="logo-mark">v</span>
          vast<span style={{ color: "var(--text-mute)" }}>.ai</span>
        </Link>
        <h1>{side.title}</h1>
        <p>{side.body}</p>
        <div className="row wrap" style={{ gap: 22, marginTop: 8 }}>
          {side.features.map((f) => (
            <div key={f.label} className="row gap-sm" style={{ color: "var(--text-dim)" }}>
              <f.icon size={16} color="var(--accent-2)" />
              {f.label}
            </div>
          ))}
        </div>
      </aside>
      <main className="auth-form">
        <div className="auth-card">
          <h2>{title}</h2>
          <div className="hint">{subtitle}</div>
          {children}
        </div>
      </main>
    </div>
  );
}

export default function AuthPage({ mode }) {
  const { login } = useAuth();
  const { account } = useStore();
  const navigate = useNavigate();
  const [email, setEmail] = useState(account.user);
  const [password, setPassword] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [error, setError] = useState("");

  const isLogin = mode === "login";

  function submit(e) {
    e.preventDefault();
    if (!email || (!isLogin && (!password || !apiKey)) || (isLogin && !password)) {
      setError("Please fill in all fields.");
      return;
    }
    if (isLogin) {
      if (password.length < 4) {
        setError("Invalid email or password.");
        return;
      }
    } else if (!/^[0-9a-f]{32,}$/i.test(apiKey.replace(/\s/g, ""))) {
      setError("API key should be 32+ hex characters.");
      return;
    }
    login(email, apiKey.replace(/\s/g, "") || "abcdef0123456789abcdef0123456789");
    navigate("/");
  }

  const side = {
    title: isLogin ? "Welcome back to your GPU fleet" : "Rent the world's GPUs in seconds",
    body: isLogin
      ? "Manage instances, watch utilization, and control spend from a single dashboard — whether you rent compute or host machines."
      : "Sign up with an API key from console.vast.ai and instantly search thousands of H100, A100, RTX 4090 and more across 40+ datacenters.",
    features: [
      { label: "Pay as you go", icon: Wallet },
      { label: "H100 · A100 · RTX", icon: Cpu },
      { label: "Datacenter or consumer GPUs", icon: Server },
      { label: "Real-time telemetry", icon: Gauge },
    ],
  };

  return (
    <AuthShell title={isLogin ? "Sign in" : "Create account"} subtitle={isLogin ? "Enter your credentials to continue." : "Generate an API key at console.vast.ai/keys and paste it below."} side={side}>
      <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        <label className="field">
          Email
          <input className="input" type="email" placeholder="you@example.com" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" />
        </label>
        <label className="field">
          Password
          <input className="input" type="password" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete={isLogin ? "current-password" : "new-password"} />
        </label>
        {!isLogin && (
          <label className="field">
            Vast.ai API key
            <input className="input mono" placeholder="abc123... (32+ hex chars)" value={apiKey} onChange={(e) => setApiKey(e.target.value)} autoComplete="off" />
          </label>
        )}
        {error && <div className="badge badge-red" style={{ borderRadius: 6, padding: "7px 10px" }}>{error}</div>}
        <button className="btn btn-primary btn-block" type="submit">
          <ShieldCheck size={15} />
          {isLogin ? "Sign in" : "Create account"}
        </button>
      </form>
      <div className="divider">or</div>
      <div className="auth-footer">
        {isLogin ? "New to Vast?" : "Already have an account?"}{" "}
        <Link to={isLogin ? "/signup" : "/login"}>{isLogin ? "Create an account" : "Sign in"}</Link>
      </div>
    </AuthShell>
  );
}
