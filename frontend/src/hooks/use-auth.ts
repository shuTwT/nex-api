import { useAuthContext } from "@/providers/auth";

export function useAuth() {
  const { user, isAuthenticated, isLoading, login, logout, refreshUser } =
    useAuthContext();

  return {
    user,
    isAuthenticated,
    isLoading,
    isAdmin: user?.role === "admin",
    credits: user?.credits ?? 0,
    login,
    logout,
    refreshUser,
  };
}