import { useMemo, useState } from "react";
import { Cpu, HardDrive, MemoryStick, Network, RefreshCw, Search, Server } from "lucide-react";
import { useStore } from "../store";
import { fmtMoney } from "../data/mock";
import { Badge, PageHead } from "../components/ui";
import RentModal from "../components/RentModal";

const SORTS = [
  { key: "score", label: "Best score", dir: -1 },
  { key: "dph_total", label: "Price low → high", dir: 1 },
  { key: "benchmark", label: "GPU benchmark", dir: -1 },
  { key: "reliability2", label: "Reliability", dir: -1 },
  { key: "gpu_name", label: "GPU name", dir: 1 },
];

const GPU_OPTIONS = ["All GPUs", "H100 SXM", "H200 SXM", "A100 SXM4", "A100 PCIe", "RTX 4090", "RTX 4080", "RTX 3090", "RTX 3080", "RTX 4070", "A6000", "L40S", "A5000"];

export default function Marketplace() {
  const { offers, templates } = useStore();
  const [renting, setRenting] = useState(null);
  const [filters, setFilters] = useState({
    q: "",
    gpu: "All GPUs",
    maxPrice: "",
    minVram: "",
    minDisk: "",
    region: "All regions",
    verified: true,
    datacenter: false,
    interruptible: false,
    sort: "score",
  });

  const regions = useMemo(() => ["All regions", ...new Set(offers.map((o) => o.region))], [offers]);

  const results = useMemo(() => {
    let list = offers.filter((o) => o.rented === false || filters.q || true);
    list = list.filter((o) => {
      if (filters.q && !`${o.gpu_name} ${o.location} ${o.machine_id}`.toLowerCase().includes(filters.q.toLowerCase())) return false;
      if (filters.gpu !== "All GPUs" && o.gpu_name !== filters.gpu) return false;
      if (filters.maxPrice && o.dph_total > Number(filters.maxPrice)) return false;
      if (filters.minVram && o.gpu_vram < Number(filters.minVram)) return false;
      if (filters.minDisk && o.disk_space < Number(filters.minDisk)) return false;
      if (filters.region !== "All regions" && o.region !== filters.region) return false;
      if (filters.verified && !o.verified) return false;
      if (filters.datacenter && !o.datacenter) return false;
      if (filters.interruptible && !o.interruptible) return false;
      return true;
    });
    const sort = SORTS.find((s) => s.key === filters.sort);
    list = [...list].sort((a, b) => {
      let r = 0;
      if (sort.key === "gpu_name") r = a.gpu_name.localeCompare(b.gpu_name);
      else r = a[sort.key] - b[sort.key];
      return r * sort.dir;
    });
    return list.slice(0, 60);
  }, [offers, filters]);

  function set(name, value) {
    setFilters((f) => ({ ...f, [name]: value }));
  }

  return (
    <div>
      <PageHead title="GPU Marketplace" sub={`${results.length} offers · cheapest verified compute across 40+ datacenters`}>
        <button className="btn btn-ghost" onClick={() => setFilters({ q: "", gpu: "All GPUs", maxPrice: "", minVram: "", minDisk: "", region: "All regions", verified: true, datacenter: false, interruptible: false, sort: "score" })}>
          <RefreshCw size={14} /> Reset filters
        </button>
      </PageHead>

      <div className="card" style={{ marginBottom: 16 }}>
        <div className="filters">
          <label className="field grow">
            Search
            <div className="row" style={{ margin: 0 }}>
              <Search size={15} color="var(--text-mute)" style={{ position: "absolute", marginLeft: 11 }} />
              <input className="input" style={{ paddingLeft: 34 }} placeholder="GPU, location, machine id…" value={filters.q} onChange={(e) => set("q", e.target.value)} />
            </div>
          </label>
          <label className="field">
            GPU
            <select className="select" value={filters.gpu} onChange={(e) => set("gpu", e.target.value)}>
              {GPU_OPTIONS.map((g) => <option key={g}>{g}</option>)}
            </select>
          </label>
          <label className="field">
            Max price ($/hr)
            <input className="input" type="number" min="0" step="0.01" placeholder="0.50" value={filters.maxPrice} onChange={(e) => set("maxPrice", e.target.value)} />
          </label>
          <label className="field">
            Min VRAM (MB)
            <input className="input" type="number" min="0" step="512" placeholder="24564" value={filters.minVram} onChange={(e) => set("minVram", e.target.value)} />
          </label>
          <label className="field">
            Min disk (TB)
            <input className="input" type="number" min="0" step="0.1" placeholder="1" value={filters.minDisk} onChange={(e) => set("minDisk", e.target.value)} />
          </label>
          <label className="field">
            Region
            <select className="select" value={filters.region} onChange={(e) => set("region", e.target.value)}>
              {regions.map((r) => <option key={r}>{r}</option>)}
            </select>
          </label>
        </div>
        <div className="row wrap mt-sm" style={{ gap: 16 }}>
          <label className="switch-field">
            <span className="switch"><input type="checkbox" checked={filters.verified} onChange={(e) => set("verified", e.target.checked)} /><span className="slider" /></span>
            Verified only
          </label>
          <label className="switch-field">
            <span className="switch"><input type="checkbox" checked={filters.datacenter} onChange={(e) => set("datacenter", e.target.checked)} /><span className="slider" /></span>
            Datacenter GPUs
          </label>
          <label className="switch-field">
            <span className="switch"><input type="checkbox" checked={filters.interruptible} onChange={(e) => set("interruptible", e.target.checked)} /><span className="slider" /></span>
            Interruptible (cheaper)
          </label>
        </div>
      </div>

      <div className="row between" style={{ marginBottom: 12 }}>
        <span className="dim small">{results.length} matching offers</span>
        <label className="row" style={{ margin: 0 }}>
          <span className="muted small">Sort by</span>
          <select className="select" style={{ width: 180 }} value={filters.sort} onChange={(e) => set("sort", e.target.value)}>
            {SORTS.map((s) => <option key={s.key} value={s.key}>{s.label}</option>)}
          </select>
        </label>
      </div>

      <div className="offer-grid">
        {results.map((o) => (
          <div className="card offer-card" key={o.id}>
            <div className="offer-top">
              <div>
                <div className="offer-gpu">
                  <Cpu size={15} style={{ verticalAlign: -2 }} /> {o.gpu_name}
                </div>
                <div className="small dim">{o.num_gpus} GPU{o.num_gpus > 1 ? "s" : ""} · {o.display_location}</div>
              </div>
              <div className="offer-price">
                {fmtMoney(o.dph_total)}<small>/hr</small>
              </div>
            </div>

            <div className="offer-meta">
              <div className="meta"><Cpu size={14} />{o.cpu_cores} vCPU</div>
              <div className="meta"><MemoryStick size={14} />{o.cpu_ram} GB RAM</div>
              <div className="meta"><HardDrive size={14} />{o.disk_space} TB NVMe</div>
              <div className="meta"><Network size={14} />{Math.round(o.inet_down / 1000)} Gbps ↓</div>
            </div>

            <div className="row wrap" style={{ gap: 6 }}>
              <Badge tone="purple">bench {o.benchmark}</Badge>
              <Badge tone={o.verified ? "green" : "neutral"}>{o.verified ? "verified" : "unverified"}</Badge>
              {o.datacenter && <Badge tone="blue">datacenter</Badge>}
              {o.interruptible && <Badge tone="amber">interruptible</Badge>}
              <Badge tone="neutral">rel {Math.round(o.reliability2 * 100)}%</Badge>
              <Badge tone="neutral">score {o.score.toFixed(3)}</Badge>
            </div>

            <div className="offer-stats">
              <span>{o.rental_count} rentals</span>
              <span>id {o.id}</span>
            </div>

            <div className="offer-footer">
              <button className="btn btn-primary" onClick={() => setRenting(o)}>Rent</button>
              <button className="btn" onClick={() => setRenting(o)}>Details</button>
            </div>
          </div>
        ))}
      </div>

      {results.length === 0 && (
        <div className="card empty">
          <Server size={40} />
          <div>No offers match your filters. Try widening the price or VRAM range.</div>
        </div>
      )}

      {renting && (
        <RentModal offer={renting} templates={templates} onClose={() => setRenting(null)} />
      )}
    </div>
  );
}
