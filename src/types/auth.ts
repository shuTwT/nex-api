export interface User {
  id: string;
  email: string;
  username: string;
  role: "user" | "admin";
  credits: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface AuthActions {
  login: (user: User, token: string) => void;
  logout: () => void;
  updateUser: (user: Partial<User>) => void;
  updateCredits: (credits: number) => void;
  setLoading: (loading: boolean) => void;
  initializeAuth: () => void;
  fetchUserInfo: () => Promise<void>;
}

export type AuthStore = AuthState & AuthActions;
