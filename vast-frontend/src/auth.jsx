import { createContext, useContext, useEffect, useState } from "react";

const AuthContext = createContext(null);
const AUTH_KEY = "vast-auth-v1";

export const SUPERADMIN_EMAIL = "admin@vast.ai";

function deriveName(email) {
  return email
    .split("@")[0]
    .replace(/[._]/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

function deriveRole(email) {
  return email.trim().toLowerCase() === SUPERADMIN_EMAIL ? "superadmin" : "user";
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => {
    try {
      const raw = JSON.parse(localStorage.getItem(AUTH_KEY));
      if (raw) raw.role = raw.role || deriveRole(raw.email || "");
      return raw || null;
    } catch {
      return null;
    }
  });

  useEffect(() => {
    try {
      if (user) localStorage.setItem(AUTH_KEY, JSON.stringify(user));
      else localStorage.removeItem(AUTH_KEY);
    } catch {
      /* ignore */
    }
  }, [user]);

  function login(email, apiKey) {
    const role = deriveRole(email);
    setUser({
      email,
      apiKey,
      role,
      name: email === SUPERADMIN_EMAIL ? "Platform Superadmin" : deriveName(email),
      since: "2023-03-14",
    });
  }

  function loginAs(email, name, role) {
    setUser({
      email,
      apiKey: "abcdef0123456789abcdef0123456789",
      role: role || deriveRole(email),
      name: name || deriveName(email),
      since: "2023-03-14",
    });
  }

  function logout() {
    setUser(null);
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        login,
        loginAs,
        logout,
        isSuperadmin: user?.role === "superadmin",
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
