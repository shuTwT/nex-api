"use client";

import type { Session } from "next-auth";

interface AuthProviderProps {
  session?: Session|null;
   children: React.ReactNode;
}

export function AuthProvider({ session, children }: AuthProviderProps) {


  return <>{children}</>;
}
