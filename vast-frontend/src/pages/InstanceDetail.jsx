import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Box, Play, Power, Terminal, Trash2, Wifi } from "lucide-react";
import { useStore } from "../store";
import { fmtBytes, fmtMoney, fmtUptime, timeAgo } from "../data/mock";
import { Badge, CopyButton, KeyValue, PageHead, StatusBadge, UtilBar } from "../components/ui";

export default function InstanceDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { instances, setInstanceStatus, destroyInstance, notify } = useStore();
  const instance = instances.find((i) => i.id === Number(id));

  if (!instance) {
    return (
      <div>
        <PageHead title="Instance not found" />
        <div className="card empty"><Box size={40} /><div>This instance was destroyed or never existed.</div></div>
        <Link to="/instances" className="btn mt">Back to instances</Link>
      </div>
    );
  }

  const isRunning = instance.actual_status === "running";

  const sshCmd = `ssh -p ${instance.ssh_port} -o StrictHostKeyChecking=no root@${instance.ssh_host}`;

  return (
    <div>
      <button className="btn btn-ghost btn-sm" onClick={() => navigate(-1)} style={{ marginBottom: 14 }}>
        <ArrowLeft size={14} /> Back
      </button>

      <PageHead
        title={instance.label}
        sub={<><span className="mono">#{instance.id}</span> · {instance.gpu_name} × {instance.num_gpus} · {instance.display_location} · started {timeAgo(instance.start_date)}</>}
      >
        {isRunning && <button className="btn" onClick={() => { setInstanceStatus(instance.id, "stopped"); notify("Instance stopped."); }}><Power size={14} /> Stop</button>}
        {!isRunning && instance.actual_status !== "off" && <button className="btn btn-green" onClick={() => { setInstanceStatus(instance.id, "running"); notify("Instance starting…"); }}><Play size={14} /> Start</button>}
        <button className="btn btn-danger" onClick={() => { destroyInstance(instance.id); navigate("/instances"); }}>
          <Trash2 size={14} /> Destroy
        </button>
      </PageHead>

      <div className="grid cols-4">
        <MiniStat label="Status" value={<StatusBadge status={instance.actual_status} />} />
        <MiniStat label="Cost" value={<><span style={{ color: "var(--green)", fontWeight: 700 }}>{fmtMoney(instance.dph_total)}</span><span className="muted small">/hr</span></>} sub={instance.bid ? "bid pricing" : "on-demand"} />
        <MiniStat label="Uptime" value={isRunning ? fmtUptime(instance.duration) : "—"} sub={`total ${fmtUptime(instance.duration_total)}`} />
        <MiniStat label="Public IP" value={<span className="mono" style={{ fontSize: 13 }}>{instance.public_ipaddr || "provisioning…"}</span>} sub={`ssh :${instance.ssh_port}`} />
      </div>

      <div className="grid cols-2" style={{ gridTemplateColumns: "1.4fr 1fr", marginTop: 16 }}>
        <div className="card">
          <div className="card-title"><h3>Usage</h3><span className="small dim">live telemetry · 2s refresh</span></div>
          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            <UsageRow label="GPU utilization" value={instance.gpu_util} />
            <UsageRow label="CPU utilization" value={instance.cpu_util} />
            <UsageRow label="Memory" value={instance.mem_util} />
            <UsageRow label="Disk usage" value={instance.disk_usage} />
          </div>
          <div className="row mt" style={{ gap: 18, color: "var(--text-dim)", fontSize: 12.5 }}>
            <span className="row gap-sm"><Wifi size={13} />{Math.round(instance.inet_down / 1000)} Gbps down</span>
            <span>↑ {Math.round(instance.inet_up / 1000)} Gbps up</span>
          </div>
        </div>

        <div className="card">
          <div className="card-title"><h3>Hardware</h3></div>
          <div className="kv">
            <KeyValue k="GPU" v={instance.gpu_name} />
            <KeyValue k="GPU count" v={instance.num_gpus} />
            <KeyValue k="VRAM" v={fmtBytes(24576)} />
            <KeyValue k="CPU" v={`${instance.cpu_cores} cores`} />
            <KeyValue k="RAM" v={`${instance.cpu_ram} GB`} />
            <KeyValue k="Disk" v={`${instance.disk_space} TB NVMe`} />
            <KeyValue k="Region" v={instance.display_location} />
            <KeyValue k="Machine" v={<span className="mono">{instance.machine_id}</span>} />
          </div>
        </div>
      </div>

      <div className="grid cols-2" style={{ gridTemplateColumns: "1fr 1fr", marginTop: 16 }}>
        <div className="card">
          <div className="card-title"><h3><Terminal size={14} style={{ verticalAlign: -2 }} /> SSH connection</h3><CopyButton text={sshCmd} label="Copy" /></div>
          <code style={{ display: "block", padding: "10px 12px", marginBottom: 10, fontSize: 12.5 }}>{sshCmd}</code>
          <div className="small dim">Jupyter token: <span className="mono">{instance.jupyter_token}</span></div>
          <div className="small dim mt-sm">Image: <span className="mono">{instance.image_name}</span></div>
        </div>

        <div className="card">
          <div className="card-title"><h3>Startup command</h3><Badge tone="blue">{instance.onstart?.runtype || "ssh"}</Badge></div>
          <div className="console" style={{ minHeight: 90 }}>
            <span className="dim">$ cat /root/onstart.sh</span>{"\n"}
            {instance.onstart?.startup_script || `# SSH into the instance to configure.`}
          </div>
        </div>
      </div>

      <div className="card mt">
        <div className="card-title"><h3>Live console</h3><span className="row gap-sm small dim"><span className="badge badge-green dot" style={{ width: 8, height: 8, borderRadius: "50%", display: "inline-block" }} /> connected</span></div>
        <Console id={instance.id} />
      </div>
    </div>
  );
}

function MiniStat({ label, value, sub }) {
  return (
    <div className="card">
      <div className="muted small" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: 0.04 }}>{label}</div>
      <div style={{ fontSize: 17, fontWeight: 700, marginTop: 6 }}>{value}</div>
      {sub && <div className="small dim mt-sm">{sub}</div>}
    </div>
  );
}

function UsageRow({ label, value }) {
  return (
    <div className="row between">
      <span className="dim" style={{ fontSize: 12.5 }}>{label}</span>
      <div className="row" style={{ gap: 10 }}>
        <UtilBar value={value} />
        <span className="mono" style={{ fontSize: 12, width: 34, textAlign: "right" }}>{value}%</span>
      </div>
    </div>
  );
}

function Console({ id }) {
  const [lines, setLines] = useState([
    { t: `vastai@instance-${id}:~$ systemctl status container`, c: "cmd" },
    { t: "● container.service - instance runtime", c: "dim" },
    { t: "   Active: active (running) since 3 days ago", c: "dim" },
    { t: "   Memory: 14.2G (57.4%)", c: "dim" },
    { t: "", c: "dim" },
    { t: "vastai@instance:~$ nvidia-smi", c: "cmd" },
    { t: "GPU  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC", c: "dim" },
    { t: "0    H100 SXM4    On           | 00000000:00:05.0 Off |                    0", c: "accent" },
    { t: "    41%   62C    P0   200W / 350W |  48172MiB / 81559MiB |    58%     Default", c: "accent" },
    { t: "", c: "dim" },
    { t: "vastai@instance:~$ tail -f /root/logs/train.log", c: "cmd" },
    { t: "Epoch 24/50  loss=1.2143  lr=2.4e-4  acc=0.814  (2.1s/step)", c: "ok" },
  ]);

  useEffect(() => {
    const busy = [
      { t: "Epoch 25/50  loss=1.1831  lr=2.3e-4  acc=0.822  (2.0s/step)", c: "ok" },
      { t: "Epoch 26/50  loss=1.1568  lr=2.3e-4  acc=0.831  (2.1s/step)", c: "ok" },
      { t: "Epoch 27/50  loss=1.1312  lr=2.2e-4  acc=0.839  (2.0s/step)", c: "ok" },
    ];
    const t = setInterval(() => {
      setLines((prev) => [...prev.slice(-40), ...busy]);
    }, 2600);
    return () => clearInterval(t);
  }, []);

  return (
    <div className="console">
      {lines.map((l, i) => (
        <div key={i} className={l.c === "accent" ? "accent" : l.c === "ok" ? undefined : l.c === "cmd" ? "" : "dim"} style={{ color: l.c === "ok" ? "var(--green)" : undefined, fontWeight: l.c === "cmd" ? 700 : undefined }}>
          {l.t || "\u00A0"}
        </div>
      ))}
      <span style={{ background: "var(--green)", display: "inline-block", width: 8, height: 14, marginLeft: 2 }} />
    </div>
  );
}
