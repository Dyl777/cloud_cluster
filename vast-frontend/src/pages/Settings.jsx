import { useState } from "react";
import { Copy, Eye, EyeOff, KeyRound, Plus, Trash2 } from "lucide-react";
import { useAuth } from "../auth";
import { useStore } from "../store";
import { Badge, PageHead } from "../components/ui";

export default function Settings() {
  const { user, logout } = useAuth();
  const { notify, account } = useStore();
  const [keys, setKeys] = useState([
    { name: "Default API key", value: user?.apiKey || "abc123def456abc123def456abc123de", scopes: ["all"], lastUsed: "2h ago" },
  ]);
  const [show, setShow] = useState(false);
  const [newKey, setNewKey] = useState("");
  const [creating, setCreating] = useState(false);

  function createKey(e) {
    e.preventDefault();
    if (!newKey) return;
    const key = Array.from({ length: 36 }, () => "0123456789abcdef"[Math.floor(Math.random() * 16)]).join("");
    setKeys((k) => [{ name: newKey, value: key, scopes: ["rent", "instances"], lastUsed: "never" }, ...k]);
    setNewKey("");
    setCreating(false);
    notify("API key created — copy it now, it won't be shown again.");
  }

  function copy(v) {
    navigator.clipboard?.writeText(v);
    notify("Copied to clipboard.");
  }

  return (
    <div>
      <PageHead title="Settings" sub="Account, API keys and CLI access for {username}." />

      <div className="card">
        <div className="card-title"><h3>Profile</h3></div>
        <div className="row" style={{ gap: 14 }}>
          <div className="avatar" style={{ width: 44, height: 44, fontSize: 16 }}>{user?.name?.split(" ").map((w) => w[0]).join("").slice(0, 2)}</div>
          <div>
            <div style={{ fontWeight: 700 }}>{user?.name}</div>
            <div className="dim small">{user?.email}</div>
          </div>
          <div style={{ marginLeft: "auto" }}>
            <Badge tone="green">plan: {account.plan}</Badge>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-title">
          <h3>API keys</h3>
          <button className="btn btn-sm btn-primary" onClick={() => setCreating((c) => !c)}><Plus size={13} /> New key</button>
        </div>
        <p className="small dim" style={{ marginTop: -8 }}>Keys grant programmatic access to the Vast.ai REST API. Treat them like passwords.</p>

        {creating && (
          <form onSubmit={createKey} style={{ display: "flex", gap: 10, margin: "10px 0 16px" }}>
            <input className="input flex-1" placeholder="Key name (e.g. ci-pipeline)" value={newKey} onChange={(e) => setNewKey(e.target.value)} autoFocus />
            <button className="btn btn-green" type="submit">Generate</button>
          </form>
        )}

        {keys.map((k) => (
          <div key={k.name} className="row between" style={{ padding: "12px 0", borderBottom: "1px solid var(--border-soft)" }}>
            <div>
              <div className="row" style={{ gap: 8 }}>
                <KeyRound size={14} color="var(--accent)" />
                <span style={{ fontWeight: 600 }}>{k.name}</span>
              </div>
              <div className="row mt-sm" style={{ gap: 8 }}>
                <code>{show ? k.value : "•".repeat(32) + k.value.slice(-4)}</code>
                <button className="btn btn-ghost btn-sm" onClick={() => copy(k.value)}><Copy size={12} /> copy</button>
                <button className="btn btn-ghost btn-sm" onClick={() => setShow((s) => !s)}>{show ? <EyeOff size={12} /> : <Eye size={12} />}</button>
              </div>
            </div>
            <div className="row" style={{ gap: 10 }}>
              <div className="small dim" style={{ textAlign: "right" }}>
                <div>{k.scopes.join(", ")}</div>
                <div>used {k.lastUsed}</div>
              </div>
              <button className="btn btn-danger btn-sm" onClick={() => setKeys((ks) => ks.filter((x) => x.name !== k.name))} title="Revoke">
                <Trash2 size={13} />
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="card">
        <div className="card-title"><h3>CLI access</h3></div>
        <div className="small dim">The <code>vastai</code> CLI is a Python package. Configure it with your API key:</div>
        <code style={{ display: "block", margin: "12px 0" }}>pip install -U vastai && vastai set api-key YOUR_API_KEY</code>
        <div className="small dim">Then manage everything from the terminal: <code>vastai search offers 'gpu_name=RTX 4090'</code></div>
      </div>

      <div className="card">
        <div className="card-title"><h3>Danger zone</h3></div>
        <button className="btn btn-danger" onClick={logout}>Sign out of this device</button>
      </div>
    </div>
  );
}
