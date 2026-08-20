import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { api } from "./data/mock";

const StoreContext = createContext(null);

const STORE_KEY = "vast-store-v1";

const DEFAULT_PAYMENT_METHODS = [
  { id: "pm-mtn", kind: "mobile_money", label: "MTN Mobile Money", carrier: "mtn", phone: "677123456" },
  { id: "pm-orange", kind: "mobile_money", label: "Orange Money", carrier: "orange", phone: "699987654" },
  { id: "pm-bank", kind: "bank", label: "Bank wire", provider: "wire", account_ref: "••••4821" },
];

function loadState() {
  try {
    const raw = localStorage.getItem(STORE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {
    /* ignore */
  }
  return null;
}

export function StoreProvider({ children }) {
  const persisted = useMemo(loadState, []);
  const [instances, setInstances] = useState(persisted?.instances ?? api.instances);
  const [offers, setOffers] = useState(persisted?.offers ?? api.offers);
  const [hosts, setHosts] = useState(persisted?.hosts ?? api.hosts);
  const [account, setAccount] = useState(persisted?.account ?? api.account);
  const [transactions, setTransactions] = useState(persisted?.transactions ?? api.transactions);
  const [paymentMethods, setPaymentMethods] = useState(
    persisted?.paymentMethods ?? DEFAULT_PAYMENT_METHODS,
  );
  const [toast, setToast] = useState(null);

  useEffect(() => {
    try {
      localStorage.setItem(
        STORE_KEY,
        JSON.stringify({ instances, offers, hosts, account, transactions, paymentMethods }),
      );
    } catch {
      /* ignore */
    }
  }, [instances, offers, hosts, account, transactions, paymentMethods]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 3500);
    return () => clearTimeout(t);
  }, [toast]);

  function notify(msg) {
    setToast(msg);
  }

  function rentOffer(offer, opts = {}) {
    const id = 24000000 + Math.floor(Math.random() * 9000000);
    const instance = {
      id,
      label: opts.label || `${offer.gpu_name.toLowerCase().replace(/\s+/g, "-")}-rental`,
      actual_status: "pending",
      cur_state: "pending",
      gpu_name: offer.gpu_name,
      num_gpus: offer.num_gpus,
      gpu_util: 0,
      cpu_util: 0,
      mem_util: 0,
      disk_usage: 0,
      dph_total: offer.dph_total,
      machine_id: offer.machine_id,
      ssh_host: offer.ssh_host || `203.${Math.floor(Math.random() * 200) + 10}.10.${Math.floor(Math.random() * 200) + 2}`,
      ssh_port: Math.floor(Math.random() * 20000) + 20000,
      public_ipaddr: null,
      jupyter_token: Math.random().toString(36).slice(2, 14),
      image_uuid: opts.template?.image_uuid || "ai-torch",
      image_name: opts.template?.image || "pytorch/pytorch:latest",
      image_url: opts.template?.image || "pytorch/pytorch:latest",
      start_date: new Date().toISOString(),
      duration: 0,
      duration_total: 0,
      inet_down: offer.inet_down,
      inet_up: offer.inet_up,
      region: offer.region,
      location: offer.location,
      display_location: offer.display_location,
      bid: opts.bid || false,
      external: false,
      cpu_cores: offer.cpu_cores,
      cpu_ram: offer.cpu_ram,
      disk_space: offer.disk_space,
      onstart: opts.onstart || { runtype: "ssh" },
    };
    setInstances((prev) => [instance, ...prev]);
    setOffers((prev) => prev.map((o) => (o.id === offer.id ? { ...o, rented: true, gpus_free: Math.max(0, (o.gpus_free || 1) - 1) } : o)));
    notify(`Instance #${id} launched (pending).`);
    return id;
  }

  function setInstanceStatus(id, status) {
    setInstances((prev) =>
      prev.map((inst) => (inst.id === id ? { ...inst, actual_status: status, cur_state: status } : inst)),
    );
  }

  function destroyInstance(id) {
    const inst = instances.find((i) => i.id === id);
    setInstances((prev) => prev.filter((i) => i.id !== id));
    setHosts((prev) => prev.map((h) => (h.machine_id === inst?.machine_id ? { ...h, num_gpus: Math.max(0, h.num_gpus - 1), num_rentable_gpus: Math.min(h.num_gpus_total, h.num_rentable_gpus + 1) } : h)));
    notify(`Instance #${id} destroyed.`);
  }

  function addTransaction(tx) {
    setTransactions((prev) => [tx, ...prev]);
  }

  function completeTopUp(amount, label = "Credit top-up") {
    setAccount((prev) => ({ ...prev, balance: +(prev.balance + amount).toFixed(2) }));
    addTransaction({
      id: Math.floor(Math.random() * 9000) + 100000,
      time: new Date().toISOString(),
      type: "credit",
      label,
      amount,
      status: "completed",
    });
    notify(`Added $${amount.toFixed(2)} to balance.`);
  }

  function addPaymentMethod(method) {
    setPaymentMethods((prev) => [...prev, method]);
    notify("Payment method saved.");
  }

  function removePaymentMethod(id) {
    setPaymentMethods((prev) => prev.filter((m) => m.id !== id));
    notify("Payment method removed.");
  }

  function addHost(host) {
    setHosts((prev) => [...prev, host]);
    notify("Host machine registered for rental.");
  }

  const value = {
    instances,
    offers,
    hosts,
    account,
    transactions,
    paymentMethods,
    templates: api.templates,
    rentOffer,
    setInstanceStatus,
    destroyInstance,
    completeTopUp,
    addPaymentMethod,
    removePaymentMethod,
    addTransaction,
    addHost,
    notify,
    toast,
  };

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore() {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error("useStore must be used within StoreProvider");
  return ctx;
}
