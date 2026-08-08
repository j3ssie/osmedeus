"use client";

import * as React from "react";
import { useRouter, usePathname } from "next/navigation";
import { login as apiLogin } from "@/lib/api/auth";

export interface User {
  id: string;
  username: string;
  email: string;
  name?: string;
}

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = React.createContext<AuthContextValue | undefined>(undefined);

const PUBLIC_PATHS = ["/login"];

function getCookieValue(name: string): string | null {
  if (typeof document === "undefined") return null;
  const parts = document.cookie.split("; ").map((p) => p.split("="));
  const found = parts.find(([k]) => k === name);
  if (!found) return null;
  const [, v] = found;
  return typeof v === "string" && v.length > 0 ? v : null;
}

function decodeCookieValue(raw: string): string {
  try {
    return decodeURIComponent(raw);
  } catch {
    return raw;
  }
}

function looksLikeJwt(token: string): boolean {
  const parts = token.split(".");
  return parts.length === 3 && parts.every((p) => p.length > 0);
}

function writeCookie(name: string, value: string) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=604800; SameSite=Lax`;
}

function clearCookie(name: string) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=; Path=/; Max-Age=0; SameSite=Lax`;
}

function readAuthFromCookie(): { token?: string; user?: User } {
  const raw = getCookieValue("osmedeus_session");
  if (!raw) return {};
  const decoded = decodeCookieValue(raw).trim();
  if (!decoded) return {};

  if (looksLikeJwt(decoded)) {
    return { token: decoded };
  }

  try {
    const parsed = JSON.parse(decoded) as unknown;
    if (parsed && typeof parsed === "object") {
      const maybeToken = (parsed as Record<string, unknown>)["token"];
      const maybeUser = (parsed as Record<string, unknown>)["user"];
      if (typeof maybeToken === "string" && looksLikeJwt(maybeToken)) {
        return { token: maybeToken };
      }
      if (
        maybeUser &&
        typeof maybeUser === "object" &&
        typeof (maybeUser as Record<string, unknown>)["username"] === "string"
      ) {
        return { user: maybeUser as User };
      }
      if (typeof (parsed as Record<string, unknown>)["username"] === "string") {
        return { user: parsed as User };
      }
    }
  } catch {
    return {};
  }

  return {};
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<User | null>(null);
  const [token, setToken] = React.useState<string | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const router = useRouter();
  const pathname = usePathname();
  const DISABLE_AUTH =
    typeof process !== "undefined" &&
    process.env.NEXT_PUBLIC_DISABLE_LOGIN === "true";
  const isAuthenticated = DISABLE_AUTH || !!token;

  // Check for existing session on mount
  React.useEffect(() => {
    const checkSession = () => {
      try {
        clearCookie("osmedeus_cookie");
        clearCookie("osmedeus_token");
        const fromStorageToken = localStorage.getItem("osmedeus_token");
        const fromStorageUser = localStorage.getItem("osmedeus_session");
        const forcedLoggedOut = localStorage.getItem("osmedeus_force_logged_out") === "true";
        const fromCookie = forcedLoggedOut ? {} : readAuthFromCookie();

        const resolvedToken = fromStorageToken || fromCookie.token || null;
        if (resolvedToken) {
          if (!fromStorageToken) {
            localStorage.setItem("osmedeus_token", resolvedToken);
          }
          setToken(resolvedToken);
        }

        if (fromStorageUser) {
          const parsed = JSON.parse(fromStorageUser) as User;
          setUser(parsed);
          return;
        }

        if (fromCookie.user) {
          localStorage.setItem("osmedeus_session", JSON.stringify(fromCookie.user));
          setUser(fromCookie.user);
          return;
        }

        if (resolvedToken) {
          setUser({
            id: "user",
            username: "osmedeus",
            email: "osmedeus@osmedeus.org",
            name: "Osmedeus",
          });
          return;
        }
      } catch {
        localStorage.removeItem("osmedeus_session");
        localStorage.removeItem("osmedeus_token");
        clearCookie("osmedeus_session");
      } finally {
        setIsLoading(false);
      }
    };

    checkSession();
  }, []);

  // Redirect logic
  React.useEffect(() => {
    if (isLoading) return;
    if (DISABLE_AUTH) return;

    const isPublicPath = PUBLIC_PATHS.includes(pathname);

    if (!isAuthenticated && !isPublicPath) {
      router.replace("/login");
    } else if (isAuthenticated && isPublicPath) {
      router.replace("/");
    }
  }, [isAuthenticated, isLoading, pathname, router, DISABLE_AUTH]);

  const login = React.useCallback(
    async (username: string, password: string) => {
      const nextToken = await apiLogin(username, password);
      localStorage.removeItem("osmedeus_force_logged_out");
      localStorage.setItem("osmedeus_token", nextToken);
      writeCookie("osmedeus_session", nextToken);
      setToken(nextToken);
      const userData: User = {
        id: `user-${Date.now()}`,
        username,
        email: `${username}@osmedeus.io`,
        name: username.charAt(0).toUpperCase() + username.slice(1),
      };
      localStorage.setItem("osmedeus_session", JSON.stringify(userData));
      setUser(userData);
      router.replace("/");
    },
    [router]
  );

  const logout = React.useCallback(() => {
    localStorage.removeItem("osmedeus_token");
    localStorage.removeItem("osmedeus_session");
    clearCookie("osmedeus_cookie");
    clearCookie("osmedeus_token");
    clearCookie("osmedeus_session");
    setToken(null);
    setUser(null);
    router.replace("/login");
  }, [router]);

  React.useEffect(() => {
    if (!isLoading && DISABLE_AUTH && !user) {
      setUser({
        id: "guest",
        username: "guest",
        email: "guest@osmedeus.io",
        name: "Guest",
      });
    }
  }, [isLoading, DISABLE_AUTH, user]);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated,
        isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = React.useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
