const GPUS = [
  { name: "H100 SXM", vram: 81559, bench: 940 },
  { name: "H200 SXM", vram: 141248, bench: 1010 },
  { name: "A100 SXM4", vram: 81559, bench: 690 },
  { name: "A100 PCIe", vram: 81559, bench: 640 },
  { name: "A6000", vram: 49140, bench: 460 },
  { name: "RTX 4090", vram: 24564, bench: 420 },
  { name: "RTX 4080", vram: 16372, bench: 310 },
  { name: "RTX 3090", vram: 24564, bench: 300 },
  { name: "RTX 3080", vram: 12228, bench: 230 },
  { name: "L40S", vram: 49140, bench: 480 },
  { name: "A5000", vram: 24564, bench: 330 },
  { name: "RTX 4070", vram: 12228, bench: 210 },
];

const LOCATIONS = [
  { region: "NAM", country: "United States", loc: "US - California" },
  { region: "NAM", country: "United States", loc: "US - Texas" },
  { region: "NAM", country: "United States", loc: "US - New Jersey" },
  { region: "EUR", country: "Netherlands", loc: "EU - Netherlands" },
  { region: "EUR", country: "Germany", loc: "EU - Germany" },
  { region: "EUR", country: "Finland", loc: "EU - Finland" },
  { region: "ASI", country: "Singapore", loc: "AS - Singapore" },
  { region: "ASI", country: "Japan", loc: "AS - Japan" },
  { region: "ASI", country: "India", loc: "AS - India" },
  { region: "SAM", country: "Brazil", loc: "SA - Brazil" },
  { region: "OCE", country: "Australia", loc: "OC - Australia" },
];

const TAGS = ["dl-per-gpu-dollar", "verified", "datacenter", "interruptible"];

function rand(min, max) {
  return Math.random() * (max - min) + min;
}

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function buildOffers() {
  const offers = [];
  const priceBy = {
    "H100 SXM": { base: 1.85, bid: 1.3 },
    "H200 SXM": { base: 2.4, bid: 1.7 },
    "A100 SXM4": { base: 1.1, bid: 0.75 },
    "A100 PCIe": { base: 0.95, bid: 0.62 },
    "A6000": { base: 0.6, bid: 0.4 },
    "RTX 4090": { base: 0.34, bid: 0.22 },
    "RTX 4080": { base: 0.24, bid: 0.15 },
    "RTX 3090": { base: 0.19, bid: 0.12 },
    "RTX 3080": { base: 0.15, bid: 0.09 },
    "L40S": { base: 0.72, bid: 0.48 },
    "A5000": { base: 0.4, bid: 0.26 },
    "RTX 4070": { base: 0.17, bid: 0.1 },
  };

  let id = 9001000;
  for (const gpu of GPUS) {
    const count = gpu.name.startsWith("H100") || gpu.name.startsWith("H200") || gpu.name.startsWith("A100") ? 14 : 24;
    for (let i = 0; i < count; i++) {
      const p = priceBy[gpu.name];
      const numGpus = pick([1, 1, 1, 1, 2, 2, 4, 8]);
      const dc = Math.random() < 0.35;
      const verified = Math.random() < 0.78;
      const location = pick(LOCATIONS);
      const interruptible = Math.random() < 0.4;
      const dphBase = +(p.base * (0.92 + rand(0, 0.16))).toFixed(3);
      const dphBid = +(p.bid * (0.9 + rand(0, 0.18))).toFixed(3);
      const basePer = interruptible ? dphBid : dphBase;
      const dphTotal = +(basePer * numGpus * 0.97).toFixed(3);
      const cpuCores = Math.round(rand(8, 32));
      const cpuRam = pick([32, 64, 96, 128, 256, 512]);
      const disk = +rand(0.3, 8).toFixed(1);
      const reliability = +(rand(0.9, 1) * 0.99).toFixed(3);
      const score = +(reliability * 0.92 + (verified ? 0.05 : 0) + (dc ? 0.03 : 0)).toFixed(3);
      const inetDown = Math.round(rand(500, 25000));
      const inetUp = Math.round(rand(200, 10000));
      const rented = Math.random() < 0.72;

      offers.push({
        id: id++,
        host_id: 10000 + i + 1,
        machine_id: "c" + Math.random().toString(16).slice(2, 10),
        gpu_name: gpu.name,
        gpu_vram: gpu.vram,
        benchmark: gpu.bench,
        num_gpus: numGpus,
        gpus_free: rented ? pick([0, 1, numGpus]) : numGpus,
        cpu_cores: cpuCores,
        cpu_ram: cpuRam,
        disk_space: disk,
        dph_total: dphTotal,
        dph_base: +(dphBase * numGpus).toFixed(3),
        dph_bid: +(dphBid * numGpus).toFixed(3),
        inet_down: inetDown,
        inet_up: inetUp,
        reliability2: reliability,
        score,
        verified,
        datacenter: dc,
        interruptible,
        rented,
        rental_count: Math.round(rand(5, 400)),
        region: location.region,
        country: location.country,
        location: location.loc,
        display_location: location.loc,
        display_price: null,
        tags: [TAGS[0], ...(verified ? [TAGS[1]] : []), ...(dc ? [TAGS[2]] : []), ...(interruptible ? [TAGS[3]] : [])],
      });
    }
  }
  return offers;
}

function buildInstances() {
  const instances = [];
  const images = [
    { uuid: "ai-torch", name: "pytorch:latest", url: "nvidia/cuda:12.1.0-base-ubuntu22.04" },
    { uuid: "ai-tf", name: "tensorflow:latest", url: "tensorflow/tensorflow:latest-gpu" },
    { uuid: "ai-jupyter", name: "jupyter/pytorch", url: "jupyter/base-notebook" },
    { uuid: "ai-vllm", name: "vllm/vllm-openai", url: "vllm/vllm-openai:latest" },
    { uuid: "ai-ollama", name: "ollama/ollama", url: "ollama/ollama:latest" },
    { uuid: "ai-comfy", name: "comfyui:latest", url: "comfyanonymous/comfyui:latest" },
  ];
  const states = ["running", "running", "running", "pending", "stopped", "off"];

  let id = 24000000;
  for (let i = 0; i < 12; i++) {
    const gpu = pick(GPUS);
    const p = gpu.name.startsWith("RTX 40") || gpu.name.startsWith("RTX 30") ? 0.3 : 1.0;
    const state = pick(states);
    const startOffset = Math.floor(rand(1, 220)) * 3600;
    const duration = Math.floor(rand(0.2, 60) * 3600);
    const location = pick(LOCATIONS);
    const sshHost = `203.${Math.floor(rand(1, 255))}.${Math.floor(rand(1, 255))}.${Math.floor(rand(2, 250))}`;
    const image = pick(images);
    const bid = Math.random() < 0.3;

    instances.push({
      id: id++,
      label: `ml-${gpu.name.toLowerCase().replace(/\s+/g, "-")}-${i + 1}`,
      actual_status: state,
      cur_state: state,
      gpu_name: gpu.name,
      num_gpus: pick([1, 1, 1, 2]),
      gpu_util: state === "running" ? Math.round(rand(0, 100)) : 0,
      cpu_util: state === "running" ? Math.round(rand(0, 80)) : 0,
      mem_util: state === "running" ? Math.round(rand(0, 90)) : 0,
      disk_usage: state === "running" ? Math.round(rand(5, 65)) : 0,
      dph_total: +(p * (bid ? 0.7 : 1) * (0.85 + rand(0, 0.3))).toFixed(3),
      machine_id: "m" + Math.random().toString(16).slice(2, 10),
      ssh_host: sshHost,
      ssh_port: Math.floor(rand(20000, 45000)),
      public_ipaddr: sshHost,
      jupyter_token: Math.random().toString(36).slice(2, 14),
      image_uuid: image.uuid,
      image_name: image.name,
      image_url: image.url,
      start_date: new Date(Date.now() - startOffset * 1000).toISOString(),
      duration: duration,
      duration_total: duration + startOffset,
      inet_down: Math.round(rand(800, 20000)),
      inet_up: Math.round(rand(400, 9000)),
      region: location.region,
      location: location.loc,
      display_location: location.loc,
      bid: bid,
      external: false,
      cpu_cores: Math.round(rand(4, 24)),
      cpu_ram: pick([32, 64, 96, 128]),
      disk_space: +rand(0.2, 2).toFixed(1),
      onstart: { runtype: "ssh", startup_script: `#!/bin/bash\npip install -r requirements.txt\npython train.py --epochs 50` },
    });
  }
  return instances;
}

function buildHosts() {
  const hosts = [];
  let id = 10001;
  for (const gpu of GPUS.slice(0, 8)) {
    const total = pick([2, 4, 8]);
    const rented = Math.round(rand(0, total));
    const dph = +(gpu.name.startsWith("H") ? 1.1 : gpu.name.startsWith("A") ? 0.55 : 0.22).toFixed(3);
    const months = Array.from({ length: 12 }, () => +(rand(0, total * dph * 720)).toFixed(0));
    hosts.push({
      id: id++,
      machine_id: "m" + Math.random().toString(16).slice(2, 10),
      gpu_name: gpu.name,
      num_gpus_total: total,
      num_gpus: rented,
      num_rentable_gpus: total - rented,
      dph: dph,
      dph_total: +(dph * rented).toFixed(3),
      total_storage: +(rand(0.5, 8) + (total === 8 ? 4 : 0)).toFixed(1),
      used_storage: +rand(0.1, 1.5).toFixed(1),
      ram: pick([128, 256, 512, 1024]),
      cpu_cores: total * pick([8, 16, 24]),
      reliability: +(rand(0.9, 1) * 0.99).toFixed(3),
      uptime: +(rand(92, 100)).toFixed(1),
      verified: Math.random() < 0.8,
      datacenter: Math.random() < 0.5,
      months: months,
      earn_rate: +(dph * rented).toFixed(3),
    });
  }
  return hosts;
}

function buildAccount() {
  return {
    user: "alex@example.com",
    name: "Alex Rivera",
    balance: 42.18,
    credits: 0,
    reserved: 0,
    pending: 0,
    total_spent: 1342.56,
    total_earned: 518.2,
    created: "2023-03-14",
    plan: "Pay-as-you-go",
  };
}

function buildTransactions() {
  const types = [
    { t: "rental", label: "Rent instance", dir: -1 },
    { t: "credit", label: "Credit top-up", dir: 1 },
    { t: "earnings", label: "Host earnings", dir: 1 },
    { t: "storage", label: "Storage fee", dir: -1 },
  ];
  const tx = [];
  const now = Date.now();
  for (let i = 0; i < 24; i++) {
    const spec = pick(types);
    const amount = +(rand(1, 60) * spec.dir).toFixed(2);
    tx.push({
      id: 9000 + i,
      time: new Date(now - i * (2.4 + rand(0, 4)) * 3600 * 1000).toISOString(),
      type: spec.t,
      label: spec.label,
      amount,
      status: pick(["completed", "completed", "completed", "pending"]),
    });
  }
  return tx;
}

function buildTemplates() {
  const tpl = [
    { name: "PyTorch", image: "pytorch/pytorch:latest", desc: "PyTorch with CUDA, Jupyter, and common data science libs", verified: true },
    { name: "vLLM", image: "vllm/vllm-openai:latest", desc: "OpenAI-compatible LLM serving for HF transformers", verified: true },
    { name: "ComfyUI", image: "comfyanonymous/comfyui:latest", desc: "Node-based Stable Diffusion / Flux pipeline GUI", verified: true },
    { name: "Ollama", image: "ollama/ollama:latest", desc: "Run LLMs like Llama, Mistral, Qwen with one command", verified: true },
    { name: "AUTOMATIC1111", image: "ghcr.io/ai-dock/stable-diffusion-webui", desc: "Stable Diffusion WebUI with extras", verified: false },
    { name: "RunPod SSH", image: "runpod/pytorch:2.2.0-py3.10-cuda12.1.1", desc: "SSH-ready pytorch base with CUDA toolchain", verified: true },
    { name: "Vast Studio", image: "vastai/studio:latest", desc: "Vast Studio IDE with VS Code Server", verified: true },
    { name: "Miniconda", image: "continuumio/miniconda3", desc: "Bare miniconda base image", verified: false },
  ];
  return tpl.map((t, i) => ({
    id: 500 + i,
    image_uuid: "ai-" + i,
    ...t,
    user: t.verified ? "vastai" : pick(["community", "kael-t", "momo-ml"]),
    downloads: Math.round(rand(100, 20000)),
  }));
}

const USER_FIRST = ["maya", "liam", "sophia", "noah", "zara", "kenji", "ade", "jelena", "omar", "priya", "yuval", "mia", "diego", "hannah", "kofi", "ines", "raj", "lucia", "tomas", "elsa"];
const USER_LAST = ["chen", "okafor", "novak", "sato", "bekele", "rossi", "khan", "garcia", "hill", "iwasa", "lopez", "müller", "obrien", "petrov", "tanaka", "weber", "yilmaz", "aldridge", "kasongo", "moreau"];

function buildUsers() {
  const users = [];
  const now = Date.now();
  const roles = ["user", "user", "user", "user", "user", "user", "admin", "user", "user", "host"];
  for (let i = 0; i < 24; i++) {
    const first = pick(USER_FIRST);
    const last = pick(USER_LAST);
    const email = `${first}.${last}@${pick(["gmail.com", "outlook.com", "hey.com", "proton.me", "bose.dev"])}`;
    const age = Math.floor(rand(3, 700));
    users.push({
      id: 3000 + i + 1,
      email,
      name: `${first.charAt(0).toUpperCase() + first.slice(1)} ${last.charAt(0).toUpperCase() + last.slice(1)}`,
      role: i === 0 ? "superadmin" : pick(roles),
      status: Math.random() < 0.88 ? "active" : Math.random() < 0.5 ? "suspended" : "pending",
      balance: +rand(0, 480).toFixed(2),
      spent: +rand(0, 9200).toFixed(2),
      earned: Math.random() < 0.4 ? +rand(0, 3400).toFixed(2) : 0,
      gpu_hours: Math.round(rand(0, 8200)),
      plan: pick(["Pay-as-you-go", "Pay-as-you-go", "monthly"]),
      created: new Date(now - age * 86400000).toISOString(),
      last_seen: new Date(now - rand(0.2, 40) * 3600000).toISOString(),
      api_key: Math.random().toString(16).slice(2, 34),
    });
  }
  return users.sort((a, b) => new Date(b.created) - new Date(a.created));
}

function buildActivity() {
  const out = [];
  const now = new Date();
  for (let i = 27; i >= 0; i--) {
    const d = new Date(now.getTime() - i * 86400000);
    out.push({
      date: d.toISOString().slice(0, 10),
      signups: Math.round(rand(2, 34)),
      revenue: +rand(180, 2200).toFixed(2),
      gpu_hours: Math.round(rand(900, 14000)),
    });
  }
  return out;
}

export const api = {
  offers: buildOffers(),
  instances: buildInstances(),
  hosts: buildHosts(),
  account: buildAccount(),
  transactions: buildTransactions(),
  templates: buildTemplates(),
  users: buildUsers(),
  activity: buildActivity(),
};

export function fmtMoney(n, digits = 2) {
  return "$" + n.toFixed(digits);
}

export function fmtBytes(mb) {
  if (mb >= 1024) return (mb / 1024).toFixed(1) + " GB";
  return Math.round(mb) + " MB";
}

export function fmtUptime(sec) {
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  if (d > 0) return `${d}d ${h}h`;
  const m = Math.floor((sec % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

export function timeAgo(iso) {
  const diff = Date.now() - new Date(iso).getTime();
  const s = Math.floor(diff / 1000);
  if (s < 60) return "just now";
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.floor(h / 24);
  return `${d}d ago`;
}

export function fmtDate(iso) {
  return new Date(iso).toLocaleString(undefined, {
    month: "short", day: "numeric", hour: "2-digit", minute: "2-digit",
  });
}
