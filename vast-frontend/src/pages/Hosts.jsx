import { useState } from "react";
import { DollarSign, HardDrive, Plus, Server, ShieldCheck, Cpu, MemoryStick, Gauge } from "lucide-react";
import { useStore } from "../store";
import { fmtMoney } from "../data/mock";
import { Badge, PageHead, Spark, StatCard } from "../components/ui";

export default function Hosts() {
  const { hosts, account } = useStore();
  const [registering, setRegistering] = useState(false);

  const totals = hosts.reduce(
    (acc, h) => {
      acc.gpus += h.num_gpus_total;
      acc.rented += h.num_gpus;
      acc.rate += h.dph_total;
      acc.verified += h.verified ? 1 : 0;
      return acc;
    },
    { gpus: 0, rented: 0, rate: 0, verified: 0 },
  );

  return (
    <div>
      <PageHead title="My Hosts" sub="Offer your GPUs to the marketplace and earn per-hour.">
        <button className="btn btn-primary" onClick={() => setRegistering(true)}><Plus size={15} /> Register machine</button>
      </PageHead>

      <div className="grid cols-4">
        <StatCard label="Machines" value={hosts.length} icon={Server} />
        <StatCard label="Total GPUs" value={totals.gpus} sub={`${totals.rented} rented right now`} icon={Gauge} tone="var(--accent-2)" />
        <StatCard label="Hourly earnings" value={fmtMoney(totals.rate)} sub="across all machines" icon={DollarSign} tone="var(--green)" />
        <StatCard label="Lifetime earned" value={fmtMoney(account.total_earned)} icon={DollarSign} tone="var(--amber)" />
      </div>

      <div className="mt">
        {hosts.map((h) => (
          <div className="card" key={h.id}>
            <div className="row between wrap" style={{ gap: 12 }}>
              <div>
                <div className="row" style={{ gap: 8 }}>
                  <h3>{h.gpu_name}</h3>
                  <Badge tone={h.verified ? "green" : "neutral"} dot>{h.verified ? "verified" : "unverified"}</Badge>
                  {h.datacenter && <Badge tone="blue">datacenter</Badge>}
                </div>
                <div className="small dim mono mt-sm">host #{h.id} · {h.machine_id}</div>
              </div>
              <div className="row" style={{ gap: 20, textAlign: "right" }}>
                <div>
                  <div className="muted small" style={{ fontSize: 11 }}>Rented / Total</div>
                  <div style={{ fontWeight: 700 }}>{h.num_gpus}<span className="muted">/{h.num_gpus_total}</span> GPU</div>
                </div>
                <div>
                  <div className="muted small" style={{ fontSize: 11 }}>Hourly</div>
                  <div style={{ fontWeight: 700, color: "var(--green)" }}>{fmtMoney(h.dph_total)}/hr</div>
                </div>
                <div>
                  <div className="muted small" style={{ fontSize: 11 }}>Rate per GPU</div>
                  <div style={{ fontWeight: 700 }}>{fmtMoney(h.dph)}</div>
                </div>
                <div>
                  <div className="muted small" style={{ fontSize: 11 }}>Uptime</div>
                  <div style={{ fontWeight: 700 }}>{h.uptime}%</div>
                </div>
              </div>
            </div>

            <div className="row between wrap mt" style={{ gap: 14 }}>
              <div className="row wrap" style={{ gap: 18, color: "var(--text-dim)", fontSize: 12.5 }}>
                <span className="row gap-sm"><MemoryStick size={13} />{h.ram} GB RAM</span>
                <span className="row gap-sm"><Cpu size={13} />{h.cpu_cores} cores</span>
                <span className="row gap-sm"><HardDrive size={13} />{h.total_storage} TB ({(h.total_storage - h.used_storage).toFixed(1)} TB free)</span>
                <span className="row gap-sm"><ShieldCheck size={13} />rel {Math.round(h.reliability * 100)}%</span>
              </div>
              <div className="row" style={{ gap: 8, alignItems: "flex-end" }}>
                <div className="row gap-sm" style={{ alignItems: "flex-end" }}>
                  <span className="muted small">12-month earnings</span>
                  <Spark values={h.months} height={34} />
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {registering && <RegisterModal onClose={() => setRegistering(false)} />}
    </div>
  );
}

function RegisterModal({ onClose }) {
  const { addHost } = useStore();
  const [gpu, setGpu] = useState("RTX 4090");
  const [count, setCount] = useState(8);
  const [dph, setDph] = useState(0.24);
  const [cpu, setCpu] = useState(16);
  const [ram, setRam] = useState(256);
  const [disk, setDisk] = useState(8);
  const [reliability, setReliability] = useState(99.1);

  function register() {
    addHost({
      id: 20000 + Math.floor(Math.random() * 9000),
      machine_id: "m" + Math.random().toString(16).slice(2, 10),
      gpu_name: gpu,
      num_gpus_total: count,
      num_gpus: 0,
      num_rentable_gpus: count,
      dph,
      dph_total: 0,
      total_storage: disk,
      used_storage: 0.1,
      ram,
      cpu_cores: cpu,
      reliability: reliability / 100,
      uptime: 100,
      verified: false,
      datacenter: false,
      months: Array(12).fill(0),
      earn_rate: 0,
    });
    onClose();
  }

  return (
    <div style={{
      position: "fixed", inset: 0, background: "var(--scrim)", backdropFilter: "blur(3px)",
      display: "grid", placeItems: "center", zIndex: 60, padding: 20,
    }} onClick={onClose}>
      <div className="card" style={{ width: "min(520px, 100%)" }} onClick={(e) => e.stopPropagation()}>
        <div className="card-title"><h3>Register a machine</h3><button className="btn btn-ghost btn-sm" onClick={onClose}>×</button></div>
        <p className="small dim" style={{ marginTop: -6 }}>
          Define the machine you're connecting to the network. You'll get a one-time setup command after registration.
        </p>
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <div className="field-row">
            <label className="field">GPU model
              <select className="select" value={gpu} onChange={(e) => setGpu(e.target.value)}>
                {["RTX 4090", "RTX 3090", "RTX 3080", "RTX 4070", "A5000", "A6000", "A100 SXM4", "H100 SXM"].map((g) => <option key={g}>{g}</option>)}
              </select>
            </label>
            <label className="field">GPU count
              <input className="input" type="number" min="1" value={count} onChange={(e) => setCount(+e.target.value)} />
            </label>
          </div>
          <div className="field-row">
            <label className="field">Price per GPU ($/hr)
              <input className="input" type="number" step="0.01" min="0.01" value={dph} onChange={(e) => setDph(+e.target.value)} />
            </label>
            <label className="field">CPU cores
              <input className="input" type="number" min="1" value={cpu} onChange={(e) => setCpu(+e.target.value)} />
            </label>
          </div>
          <div className="field-row">
            <label className="field">RAM (GB)
              <input className="input" type="number" min="8" value={ram} onChange={(e) => setRam(+e.target.value)} />
            </label>
            <label className="field">Disk (TB)
              <input className="input" type="number" step="0.1" min="0.1" value={disk} onChange={(e) => setDisk(+e.target.value)} />
            </label>
          </div>
          <label className="field">Expected reliability (%)
            <input className="input" type="number" step="0.1" min="50" max="100" value={reliability} onChange={(e) => setReliability(+e.target.value)} />
          </label>
          <div className="row" style={{ justifyContent: "flex-end", gap: 10 }}>
            <button className="btn" onClick={onClose}>Cancel</button>
            <button className="btn btn-green" onClick={register}><Plus size={14} /> Register</button>
          </div>
        </div>
      </div>
    </div>
  );
}
