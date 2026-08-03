import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Check, Cpu, HardDrive, MemoryStick, Network, X } from "lucide-react";
import { useStore } from "../store";
import { fmtMoney } from "../data/mock";
import { Badge } from "./ui";

export default function RentModal({ offer, templates, onClose }) {
  const { rentOffer } = useStore();
  const navigate = useNavigate();
  const [template, setTemplate] = useState(templates[0]);
  const [label, setLabel] = useState(`${offer.gpu_name.toLowerCase().replace(/\s+/g, "-")}-rental`);
  const [bid, setBid] = useState(false);
  const [runtype, setRuntype] = useState("ssh");
  const [script, setScript] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const onKey = (e) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const price = bid ? offer.dph_bid : offer.dph_total;

  function rent() {
    setBusy(true);
    const id = rentOffer(offer, {
      label,
      template,
      bid,
      onstart: runtype === "script" ? { runtype: "script", startup_script: script } : { runtype: "ssh" },
    });
    setTimeout(() => {
      onClose();
      navigate("/instances", { state: { highlight: id } });
    }, 500);
  }

  return (
    <div style={{
      position: "fixed", inset: 0, background: "rgba(31,35,40,0.45)", backdropFilter: "blur(3px)",
      display: "grid", placeItems: "center", zIndex: 60, padding: 20,
    }} onClick={onClose}>
      <div className="card" style={{ width: "min(760px, 100%)", maxHeight: "88vh", overflow: "auto", padding: 0 }} onClick={(e) => e.stopPropagation()}>
        <div className="row between" style={{ padding: "18px 20px", borderBottom: "1px solid var(--border-soft)" }}>
          <div>
            <h3>Rent GPU</h3>
            <div className="small dim">{offer.gpu_name} · {offer.num_gpus}× · {offer.display_location}</div>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={15} /></button>
        </div>

        <div style={{ padding: "18px 20px", display: "flex", flexDirection: "column", gap: 18 }}>
          <div className="grid cols-4" style={{ gap: 10 }}>
            <div className="row gap-sm small"><Cpu size={14} />{offer.cpu_cores} vCPU</div>
            <div className="row gap-sm small"><MemoryStick size={14} />{offer.cpu_ram} GB</div>
            <div className="row gap-sm small"><HardDrive size={14} />{offer.disk_space} TB</div>
            <div className="row gap-sm small"><Network size={14} />{Math.round(offer.inet_down / 1000)} Gbps</div>
          </div>

          <div>
            <div className="small dim" style={{ marginBottom: 8 }}>Docker template</div>
            <div className="grid cols-2" style={{ gap: 10 }}>
              {templates.slice(0, 6).map((t) => (
                <div key={t.id}
                  className="row"
                  style={{
                    gap: 10, padding: "10px 12px", borderRadius: 8, cursor: "pointer",
                    border: `1px solid ${template?.id === t.id ? "var(--accent)" : "var(--border)"}`,
                    background: template?.id === t.id ? "rgba(9,105,218,0.06)" : "var(--bg-soft)",
                  }}
                  onClick={() => setTemplate(t)}>
                  <div className="flex-1">
                    <div className="row" style={{ gap: 6 }}>
                      <span style={{ fontWeight: 600 }}>{t.name}</span>
                      {t.verified && <Badge tone="green">verified</Badge>}
                    </div>
                    <div className="small dim mono" style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{t.image}</div>
                  </div>
                  {template?.id === t.id && <Check size={16} color="var(--accent)" />}
                </div>
              ))}
            </div>
          </div>

          <div className="field-row">
            <label className="field">
              Instance label
              <input className="input mono" value={label} onChange={(e) => setLabel(e.target.value)} />
            </label>
            <label className="field">
              Run type
              <select className="select" value={runtype} onChange={(e) => setRuntype(e.target.value)}>
                <option value="ssh">SSH + Jupyter</option>
                <option value="script">Startup script</option>
              </select>
            </label>
          </div>

          {runtype === "script" && (
            <label className="field">
              Startup script
              <textarea className="input mono" rows={5} value={script} onChange={(e) => setScript(e.target.value)} placeholder="#!/bin/bash&#10;echo hello" />
            </label>
          )}

          <label className="switch-field">
            <span className="switch"><input type="checkbox" checked={bid} onChange={(e) => setBid(e.target.checked)} /><span className="slider" /></span>
            Use bid pricing <span className="muted small">(cheaper, may be interrupted)</span>
          </label>

          <div className="row between" style={{ borderTop: "1px solid var(--border-soft)", paddingTop: 16 }}>
            <div>
              <div className="muted small">Total cost</div>
              <div className="stat-value" style={{ fontSize: 22, color: "var(--green)" }}>
                {fmtMoney(price)}<span className="muted small" style={{ fontSize: 12 }}>/hr</span>
              </div>
            </div>
            <div className="row" style={{ gap: 10 }}>
              <button className="btn" onClick={onClose}>Cancel</button>
              <button className="btn btn-green" onClick={rent} disabled={busy}>
                {busy ? "Launching…" : "Rent for " + fmtMoney(price) + "/hr"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
