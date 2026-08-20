import { useEffect, useState } from "react";
import { Check, CreditCard, Smartphone, X } from "lucide-react";
import { useAuth } from "../auth";
import { useStore } from "../store";
import { fmtMoney } from "../data/mock";
import { Badge } from "./ui";
import { buildUSSD, confirmTopup, dialUSSD, startTopup } from "../api/payments";

const STEPS = { pick: "pick", dial: "dial", confirm: "confirm" };

export default function TopUpModal({ amount, onClose }) {
  const { user } = useAuth();
  const { paymentMethods, completeTopUp } = useStore();
  const [methodId, setMethodId] = useState(paymentMethods[0]?.id || "");
  const [step, setStep] = useState(STEPS.pick);
  const [busy, setBusy] = useState(false);
  const [topupId, setTopupId] = useState(null);
  const [ussd, setUssd] = useState("");
  const [carrierMsg, setCarrierMsg] = useState("");
  const [routePath, setRoutePath] = useState("");
  const [routeReason, setRouteReason] = useState("");
  const [systemDest, setSystemDest] = useState(null);
  const [error, setError] = useState("");

  const method = paymentMethods.find((m) => m.id === methodId);
  const isMobile = method?.kind === "mobile_money";
  const amountUnits = Math.round(amount);

  useEffect(() => {
    const onKey = (e) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function beginTopUp() {
    if (!method) return;
    setBusy(true);
    setError("");
    try {
      const subunits = Math.round(amount * 1_000_000);
      const res = await startTopup({
        user_id: user.email,
        payment_method_id: method.id,
        method: method.kind,
        subunits,
        currency: "USD",
      });
      setTopupId(res.topup_id);
      const intent = res.intent || {};
      const code = intent.ussd_code || buildUSSD(method.carrier, method.phone, amountUnits);
      setUssd(code || "");
      setRoutePath(intent.route_path || "");
      setRouteReason(intent.route_reason || intent.raw || "");
      setSystemDest(intent.system_destination || null);

      if (isMobile && code) {
        setStep(STEPS.dial);
        if (intent.route_path === "direct") {
          try {
            const dial = await dialUSSD({
              carrier: method.carrier,
              phone: method.phone,
              amount: amountUnits,
              ussd: code,
            });
            setCarrierMsg(dial.message);
          } catch {
            setCarrierMsg(
              `Enter this code on your ${method.carrier.toUpperCase()} line: ${code}. ` +
              "Run simbridge locally if using a laptop SIM modem.",
            );
          }
        } else if (intent.route_path === "mobilegateway") {
          setCarrierMsg(
            intent.raw || `Dispatched to gateway node ${intent.node_id || "online"}. Approve on the device SIM.`,
          );
        } else if (intent.route_path === "mobilevm") {
          setCarrierMsg(intent.raw || "Cross-rail transfer running via mobileVM.");
        } else {
          setCarrierMsg(intent.raw || "Complete payment on your device.");
        }
        setStep(STEPS.confirm);
      } else {
        setStep(STEPS.confirm);
        setCarrierMsg(`Complete payment via ${method.label}, then confirm below.`);
      }
    } catch (e) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  async function finishTopUp() {
    if (!topupId) return;
    setBusy(true);
    setError("");
    try {
      await confirmTopup(topupId);
      completeTopUp(amount, method?.label || "Credit top-up");
      onClose();
    } catch (e) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  function methodIcon(m) {
    return m.kind === "mobile_money" ? Smartphone : CreditCard;
  }

  return (
    <div
      style={{
        position: "fixed", inset: 0, background: "var(--scrim)", backdropFilter: "blur(3px)",
        display: "grid", placeItems: "center", zIndex: 60, padding: 20,
      }}
      onClick={onClose}
    >
      <div
        className="card"
        style={{ width: "min(520px, 100%)", padding: 0 }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="row between" style={{ padding: "16px 18px", borderBottom: "1px solid var(--border-soft)" }}>
          <div>
            <h3>Add {fmtMoney(amount)}</h3>
            <div className="small dim">Choose a saved payment method</div>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose}><X size={15} /></button>
        </div>

        <div style={{ padding: "16px 18px", display: "flex", flexDirection: "column", gap: 14 }}>
          {step === STEPS.pick && (
            <>
              {paymentMethods.length === 0 ? (
                <div className="small dim">
                  No payment methods saved. Add one in Settings → Payment methods.
                </div>
              ) : (
                paymentMethods.map((m) => {
                  const Icon = methodIcon(m);
                  const active = m.id === methodId;
                  return (
                    <button
                      key={m.id}
                      type="button"
                      className="card"
                      style={{
                        textAlign: "left", cursor: "pointer", padding: "12px 14px",
                        borderColor: active ? "var(--accent)" : undefined,
                        background: active ? "var(--blue-bg)" : undefined,
                      }}
                      onClick={() => setMethodId(m.id)}
                    >
                      <div className="row between">
                        <div className="row" style={{ gap: 10 }}>
                          <Icon size={16} color="var(--accent)" />
                          <div>
                            <div style={{ fontWeight: 600 }}>{m.label}</div>
                            <div className="small dim">
                              {m.kind === "mobile_money"
                                ? `${m.carrier?.toUpperCase()} · ${m.phone}`
                                : `${m.kind} · ${m.provider}`}
                            </div>
                          </div>
                        </div>
                        {active && <Check size={16} color="var(--accent)" />}
                      </div>
                    </button>
                  );
                })
              )}
              {error && <div className="small" style={{ color: "var(--red)" }}>{error}</div>}
              <button
                className="btn btn-green"
                style={{ width: "100%", justifyContent: "center" }}
                disabled={!method || busy}
                onClick={beginTopUp}
              >
                Continue with {fmtMoney(amount)}
              </button>
            </>
          )}

          {(step === STEPS.dial || step === STEPS.confirm) && (
            <>
              {routePath && (
                <div className="small dim">
                  Route: <code>{routePath}</code>
                  {routeReason && <> — {routeReason}</>}
                </div>
              )}
              {systemDest?.number && (
                <div className="small dim">
                  Platform collection: <code>{systemDest.number}</code>
                  {systemDest.label && <> ({systemDest.label})</>}
                </div>
              )}
              {isMobile && ussd && (
                <div className="card" style={{ background: "var(--bg-soft)", padding: "12px 14px" }}>
                  <div className="small dim">USSD code</div>
                  <code style={{ fontSize: 15, fontWeight: 700 }}>{ussd}</code>
                </div>
              )}
              {carrierMsg && (
                <div className="card" style={{ borderColor: "var(--amber)", padding: "12px 14px" }}>
                  <div style={{ marginBottom: 8 }}><Badge tone="amber">Carrier message</Badge></div>
                  <div className="small">{carrierMsg}</div>
                </div>
              )}
              <div className="small dim">
                {isMobile
                  ? "Approve the payment on your phone when prompted, then confirm below."
                  : "Complete the transfer, then confirm below."}
              </div>
              {error && <div className="small" style={{ color: "var(--red)" }}>{error}</div>}
              <button
                className="btn btn-green"
                style={{ width: "100%", justifyContent: "center" }}
                disabled={busy}
                onClick={finishTopUp}
              >
                I completed the payment
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
