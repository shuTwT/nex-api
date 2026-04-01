import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import type { AuthStore, User } from "@/types/auth";

const STORAGE_KEY = "one-api-auth";

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: true,

      login: (user: User, token: string) => {
        set({
          user,
          token,
          isAuthenticated: true,
          isLoading: false,
        });
      },

      logout: () => {
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          isLoading: false,
        });
      },

      updateUser: (userData: Partial<User>) => {
        const currentUser = get().user;
        if (currentUser) {
          set({
            user: { ...currentUser, ...userData },
          });
        }
      },

      updateCredits: (credits: number) => {
        const currentUser = get().user;
        if (currentUser) {
          set({
            user: { ...currentUser, credits },
          });
        }
      },

      initializeAuth: () => {
        const state = get();
        if (state.token && state.user) {
          set({
            isAuthenticated: true,
            isLoading: false,
          });
        } else {
          set({
            isAuthenticated: false,
            isLoading: false,
          });
        }
      },
      fetchUserInfo:async()=>{
        try{
          const response = await fetch("/api/auth/me",{
            credentials: "include",
          })

          if(response.ok) {
            const data = await response.json();
            if(data.success && data.data) {
              const user: User = {
                id: data.data.id,
                email: data.data.email,
                username: data.data.username,
                role: data.data.role,
                credits: data.data.credits ?? 0,
              };
              set({
                user,
                isAuthenticated: true,
                isLoading: false,
              });
            }
          }
        }catch(error){
          console.error("Failed to fetch user info:", error);
        }
        
      }
    }),
    {
      name: STORAGE_KEY,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.isLoading = false;
        }
      },
    }
  )
);

export const selectUser = (state: AuthStore) => state.user;
export const selectIsAuthenticated = (state: AuthStore) => state.isAuthenticated;
export const selectToken = (state: AuthStore) => state.token;
export const selectIsLoading = (state: AuthStore) => state.isLoading;
export const selectUserCredits = (state: AuthStore) => state.user?.credits ?? 0;
export const selectUserRole = (state: AuthStore) => state.user?.role ?? "user";
export const selectIsAdmin = (state: AuthStore) => state.user?.role === "admin";
