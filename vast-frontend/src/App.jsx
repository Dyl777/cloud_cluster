import { Route, Routes, Navigate, useLocation } from "react-router-dom";
import { AuthProvider, useAuth } from "./auth";
import { StoreProvider } from "./store";
import { Topbar, Toast } from "./components/ui";
import Marketplace from "./pages/Marketplace";
import Instances from "./pages/Instances";
import InstanceDetail from "./pages/InstanceDetail";
import Hosts from "./pages/Hosts";
import Billing from "./pages/Billing";
import Settings from "./pages/Settings";
import AuthPage from "./pages/AuthPage";

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
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
      <Toast />
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <StoreProvider>
        <Shell />
      </StoreProvider>
    </AuthProvider>
  );
}
