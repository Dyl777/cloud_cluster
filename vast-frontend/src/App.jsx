import { Route, Routes, Navigate, Outlet, useLocation } from "react-router-dom";
import { AuthProvider, useAuth } from "./auth";
import { StoreProvider } from "./store";
import { ThemeProvider } from "./theme";
import { Topbar, Toast } from "./components/ui";
import Marketplace from "./pages/Marketplace";
import Instances from "./pages/Instances";
import InstanceDetail from "./pages/InstanceDetail";
import Hosts from "./pages/Hosts";
import Billing from "./pages/Billing";
import Settings from "./pages/Settings";
import AuthPage from "./pages/AuthPage";
import AdminOverview from "./pages/admin/AdminOverview";
import AdminUsers from "./pages/admin/AdminUsers";
import AdminOffers from "./pages/admin/AdminOffers";
import AdminHosts from "./pages/admin/AdminHosts";
import AdminInstances from "./pages/admin/AdminInstances";
import AdminPayments from "./pages/admin/AdminPayments";

function AdminGate() {
  const { isSuperadmin } = useAuth();
  if (!isSuperadmin) return <Navigate to="/" replace />;
  return <Outlet />;
}

function Shell() {
  const { user } = useAuth();
  const location = useLocation();

  if (!user) {
    return <AuthPage mode="login" />;
  }

  return (
    <div className="app">
      <Topbar />
      <main className="content">
        <Routes location={location}>
          <Route path="/" element={<Marketplace />} />
          <Route path="/instances" element={<Instances />} />
          <Route path="/instances/:id" element={<InstanceDetail />} />
          <Route path="/hosts" element={<Hosts />} />
          <Route path="/billing" element={<Billing />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/admin" element={<AdminGate />}>
            <Route index element={<AdminOverview />} />
            <Route path="users" element={<AdminUsers />} />
            <Route path="offers" element={<AdminOffers />} />
            <Route path="hosts" element={<AdminHosts />} />
            <Route path="instances" element={<AdminInstances />} />
            <Route path="payments" element={<AdminPayments />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
      <Toast />
    </div>
  );
}

export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <StoreProvider>
          <Shell />
        </StoreProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}
