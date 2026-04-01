"use client";

import { useAuthStore } from "@/stores/auth-store";
import type { User } from "@/types/auth";
import { signOut } from "next-auth/react";

export function useAuth() {
  const {
    user,
    token,
    isAuthenticated,
    login,
    logout,
    updateUser,
    updateCredits,
    initializeAuth,
  } = useAuthStore();

  const handleLogin = async (userData: User, authToken: string) => {
    login(userData, authToken);
  };

  const handleLogout = async () => {
    try {
      await signOut()
    } catch (error) {
      console.error("Logout API call failed:", error);
    } finally {
      logout();
    }
  };

  const refreshUserInfo = async () => {
    if (!token) return;

    try {
      const response = await fetch("/api/auth/me", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        if (data.success && data.data) {
          updateUser(data.data);
        }
      }
    } catch (error) {
      console.error("Failed to refresh user info:", error);
    } 
  };

  return {
    user,
    token,
    isAuthenticated,
    isAdmin: user?.role === "admin",
    credits: user?.credits ?? 0,
    login: handleLogin,
    logout: handleLogout,
    updateUser,
    updateCredits,
    refreshUserInfo,
    initializeAuth,
  };
}
