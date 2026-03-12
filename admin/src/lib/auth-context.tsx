"use client";

import React, { createContext, useContext, useState, useCallback, useEffect } from "react";
import { exchangeClient } from "./api-client";

interface AdminUser {
  id: string;
  email: string;
  display_name: string;
  role: string;
}

interface AuthContextType {
  token: string | null;
  user: AdminUser | null;
  isAuthenticated: boolean;
  login: (token: string, user?: AdminUser) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

function getStoredToken() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("admin_token");
}

function getStoredUser(): AdminUser | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem("admin_user");
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => getStoredToken());
  const [user, setUser] = useState<AdminUser | null>(() => getStoredUser());

  // Validate session on mount by calling Exchange /admin/auth/me
  useEffect(() => {
    const storedToken = getStoredToken();
    if (!storedToken) return;

    exchangeClient
      .get("/api/v1/admin/auth/me", {
        headers: { Authorization: `Bearer ${storedToken}` },
      })
      .then((res) => {
        const data = res.data?.data ?? res.data;
        if (data) {
          const adminUser: AdminUser = {
            id: data.id || data.admin_id || "",
            email: data.email || "",
            display_name: data.name || data.display_name || data.email || "Admin",
            role: data.role || "admin",
          };
          localStorage.setItem("admin_user", JSON.stringify(adminUser));
          setUser(adminUser);
        }
      })
      .catch(() => {
        // Session invalid/expired — clear and redirect
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_user");
        setToken(null);
        setUser(null);
      });
  }, []);

  const login = useCallback((newToken: string, newUser?: AdminUser) => {
    localStorage.setItem("admin_token", newToken);
    setToken(newToken);
    if (newUser) {
      localStorage.setItem("admin_user", JSON.stringify(newUser));
      setUser(newUser);
    }
  }, []);

  const logout = useCallback(() => {
    const storedToken = localStorage.getItem("admin_token");
    if (storedToken) {
      exchangeClient
        .post(
          "/api/v1/admin/auth/logout",
          {},
          { headers: { Authorization: `Bearer ${storedToken}` } }
        )
        .catch(() => {});
    }
    localStorage.removeItem("admin_token");
    localStorage.removeItem("admin_user");
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{ token, user, isAuthenticated: !!token, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
