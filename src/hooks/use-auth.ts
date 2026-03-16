"use client";

import { useAuthStore } from "@/stores/auth-store";
import type { User } from "@/types/auth";

export function useAuth() {
  const {
    user,
    token,
    isAuthenticated,
    isLoading,
    login,
    logout,
    updateUser,
    updateCredits,
    setLoading,
    initializeAuth,
  } = useAuthStore();

  const handleLogin = async (userData: User, authToken: string) => {
    login(userData, authToken);
  };

  const handleLogout = async () => {
    try {
      if (token) {
        await fetch("/api/auth/logout", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
        });
      }
    } catch (error) {
      console.error("Logout API call failed:", error);
    } finally {
      logout();
    }
  };

  const refreshUserInfo = async () => {
    if (!token) return;

    try {
      setLoading(true);
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
    } finally {
      setLoading(false);
    }
  };

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
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
