"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import type { User } from "@/types/auth";

interface AuthProviderProps {
  children: React.ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const { login, logout, setLoading } = useAuthStore();

  useEffect(() => {
    const syncAuthState = async () => {
      setLoading(true);
      try {
        const response = await fetch("/api/auth/me", {
          credentials: "include",
        });

        if (response.ok) {
          const data = await response.json();
          if (data.success && data.data) {
            const user: User = {
              id: data.data.id,
              email: data.data.email,
              username: data.data.username,
              role: data.data.role,
              credits: data.data.credits ?? 0,
            };
            login(user, "session");
          } else {
            logout();
          }
        } else {
          logout();
        }
      } catch (error) {
        console.error("Failed to sync auth state:", error);
        logout();
      }
    };

    syncAuthState();
  }, [login, logout, setLoading]);

  return <>{children}</>;
}
