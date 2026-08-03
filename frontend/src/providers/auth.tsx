import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import type { User } from "@/types/auth";

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const CSRF_COOKIE_NAME = "nex_csrf";
const CSRF_HEADER_NAME = "X-CSRF-Token";

function getCSRFToken(): string {
  const match = document.cookie.match(
    new RegExp(`(?:^|; )${CSRF_COOKIE_NAME}=([^;]*)`),
  );
  return match ? decodeURIComponent(match[1]) : "";
}

async function ensureCSRFToken(): Promise<string> {
  const existing = getCSRFToken();
  if (existing) return existing;
  try {
    const res = await fetch("/api/auth/csrf", { credentials: "include" });
    const data = await res.json();
    return data?.data?.token ?? "";
  } catch {
    return "";
  }
}

function parseUser(data: unknown): User | null {
  if (!data || typeof data !== "object") return null;
  const raw = data as Record<string, unknown>;
  return {
    id: String(raw.id ?? ""),
    email: String(raw.email ?? ""),
    username: String(raw.username ?? raw.name ?? ""),
    role: (raw.role as "user" | "admin") ?? "user",
    credits: typeof raw.credits === "number" ? raw.credits : 0,
  };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    try {
      const res = await fetch("/api/auth/me", { credentials: "include" });
      if (res.ok) {
        const json = await res.json();
        if (json?.success) {
          setUser(parseUser(json.data));
          return;
        }
      }
      setUser(null);
    } catch {
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshUser();
  }, [refreshUser]);

  const login = useCallback(
    async (email: string, password: string): Promise<boolean> => {
      const csrfToken = await ensureCSRFToken();
      try {
        const res = await fetch("/api/auth/login", {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            [CSRF_HEADER_NAME]: csrfToken,
          },
          body: JSON.stringify({ email, password }),
        });
        if (res.ok) {
          await refreshUser();
          return true;
        }
        return false;
      } catch {
        return false;
      }
    },
    [refreshUser],
  );

  const logout = useCallback(async () => {
    const csrfToken = getCSRFToken();
    try {
      await fetch("/api/auth/logout", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          [CSRF_HEADER_NAME]: csrfToken,
        },
      });
    } catch {
      // best-effort
    } finally {
      setUser(null);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: user !== null,
        isLoading,
        login,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }
  return ctx;
}