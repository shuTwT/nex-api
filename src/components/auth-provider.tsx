"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import type { User } from "@/types/auth";
import type { Session } from "next-auth";

interface AuthProviderProps {
  session?: Session|null;
   children: React.ReactNode;
}

export function AuthProvider({ session, children }: AuthProviderProps) {
  const { login, logout, setLoading, fetchUserInfo } = useAuthStore();

  useEffect(() => {
    const syncAuthState = async () => {
      setLoading(true);
      await fetchUserInfo()
    };

    syncAuthState();
  }, [login, logout, setLoading]);

  return <>{children}</>;
}
