"use client";

import type { Session } from "next-auth";
import { useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import type { User } from "@/types/auth";

interface AuthProviderProps {
  session?: Session | null;
  children: React.ReactNode;
}

interface SessionUser {
  id: string;
  email?: string | null;
  name?: string | null;
  role?: string;
}

export function AuthProvider({ session, children }: AuthProviderProps) {
  useEffect(() => {
    if (!session?.user) return;

    const state = useAuthStore.getState();
    if (state.isAuthenticated) return;

    const sessionUser = session.user as SessionUser;
    const user: User = {
      id: sessionUser.id,
      email: sessionUser.email || "",
      username: sessionUser.name || sessionUser.email || "",
      role: (sessionUser.role as "user" | "admin") || "user",
      credits: 0,
    };
    useAuthStore.setState({
      user,
      token: "",
      isAuthenticated: true,
      isLoading: false,
    });
  }, [session]);

  return <>{children}</>;
}
