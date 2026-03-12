"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { useAuth } from "@/lib/auth-context";
import { useRouter } from "next/navigation";
import { exchangeClient } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Loader2 } from "lucide-react";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [googleLoading, setGoogleLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();
  const popupRef = useRef<Window | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const handleLoginSuccess = useCallback(
    (token: string, admin: Record<string, string | undefined>) => {
      const adminUser = {
        id: admin.id || "",
        email: admin.email || "",
        display_name:
          admin.name ||
          admin.full_name ||
          admin.display_name ||
          admin.email ||
          "Admin",
        role: admin.role || "admin",
      };
      login(token, adminUser);
      router.push("/");
    },
    [login, router]
  );

  // Cleanup popup polling on unmount
  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await exchangeClient.post("/api/v1/admin/auth/login", {
        email,
        password,
      });

      const data = res.data?.data ?? res.data;

      if (data?.token) {
        handleLoginSuccess(data.token, data.admin || { email });
      } else {
        setError("Unexpected response from server.");
      }
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: {
          data?: { error?: { message?: string } | string; message?: string };
          status?: number;
        };
      };
      if (axiosErr.response?.status === 401) {
        setError("Invalid email or password.");
      } else if (typeof axiosErr.response?.data?.error === "string") {
        setError(axiosErr.response.data.error);
      } else if (
        typeof axiosErr.response?.data?.error === "object" &&
        axiosErr.response?.data?.error?.message
      ) {
        setError(axiosErr.response.data.error.message);
      } else if (axiosErr.response?.data?.message) {
        setError(axiosErr.response.data.message);
      } else {
        setError("Failed to connect to the server.");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleLogin = async () => {
    setError("");
    setGoogleLoading(true);

    try {
      const res = await exchangeClient.get(
        "/api/v1/admin/auth/oauth/google"
      );
      const data = res.data?.data ?? res.data;

      if (!data?.url) {
        setError("Failed to get Google sign-in URL.");
        setGoogleLoading(false);
        return;
      }

      // Open Google OAuth in a popup window
      const width = 500;
      const height = 600;
      const left = window.screenX + (window.outerWidth - width) / 2;
      const top = window.screenY + (window.outerHeight - height) / 2;

      popupRef.current = window.open(
        data.url,
        "google-oauth",
        `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no`
      );

      // Poll the popup for the OAuth callback redirect
      // The Exchange will redirect to its admin web URL with ?token=...
      if (pollRef.current) clearInterval(pollRef.current);
      pollRef.current = setInterval(() => {
        try {
          const popup = popupRef.current;
          if (!popup || popup.closed) {
            if (pollRef.current) clearInterval(pollRef.current);
            setGoogleLoading(false);
            return;
          }

          // Try reading the popup URL — will throw if cross-origin
          const popupUrl = popup.location.href;

          // Check if the popup was redirected to the callback with token
          if (popupUrl.includes("oauth-callback")) {
            const url = new URL(popupUrl);
            const token = url.searchParams.get("token");
            const adminEmail = url.searchParams.get("email");
            const adminId = url.searchParams.get("admin_id");

            popup.close();
            if (pollRef.current) clearInterval(pollRef.current);

            if (token) {
              handleLoginSuccess(token, {
                id: adminId || "",
                email: adminEmail || "",
              });
            } else {
              const errorMsg = url.searchParams.get("error");
              setError(errorMsg || "Google sign-in failed.");
              setGoogleLoading(false);
            }
          }
        } catch {
          // Cross-origin — popup is still on Google's domain, keep polling
        }
      }, 500);
    } catch {
      setError("Google sign-in is not available.");
      setGoogleLoading(false);
    }
  };

  const isLoading = loading || googleLoading;

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center pb-2">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
            <svg
              className="h-6 w-6 text-primary"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z"
              />
            </svg>
          </div>
          <CardTitle className="text-xl">XEX Play Admin</CardTitle>
          <CardDescription>
            Sign in with your XEX Exchange admin credentials
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Google Sign-In */}
          <Button
            type="button"
            variant="outline"
            className="w-full"
            disabled={isLoading}
            onClick={handleGoogleLogin}
          >
            {googleLoading ? (
              <Loader2 className="size-4 mr-2 animate-spin" />
            ) : (
              <svg className="size-4 mr-2" viewBox="0 0 24 24">
                <path
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
                  fill="#4285F4"
                />
                <path
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  fill="#34A853"
                />
                <path
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                  fill="#FBBC05"
                />
                <path
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                  fill="#EA4335"
                />
              </svg>
            )}
            {googleLoading ? "Signing in..." : "Continue with Google"}
          </Button>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-card px-2 text-muted-foreground">
                Or continue with email
              </span>
            </div>
          </div>

          {/* Email/Password Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@xex.exchange"
                autoComplete="email"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter your password"
                autoComplete="current-password"
                required
              />
            </div>

            {error && (
              <p className="text-sm text-destructive text-center">{error}</p>
            )}

            <Button type="submit" className="w-full" disabled={isLoading}>
              {loading && <Loader2 className="size-4 mr-2 animate-spin" />}
              {loading ? "Signing in..." : "Sign in"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
