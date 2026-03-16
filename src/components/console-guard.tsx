"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { consoleMenuItems } from "@/config/console-menu";

interface ConsoleGuardProps {
  children: React.ReactNode;
}

export function ConsoleGuard({ children }: ConsoleGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { user, isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (isLoading) return;

    if (!isAuthenticated) {
      router.push("/unauthorized");
      return;
    }

    const currentItem = consoleMenuItems.find(item => item.href === pathname);
    
    if (currentItem?.adminOnly && user?.role !== "admin") {
      router.push("/forbidden");
      return;
    }
  }, [isAuthenticated, isLoading, pathname, user, router]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  const currentItem = consoleMenuItems.find(item => item.href === pathname);
  if (currentItem?.adminOnly && user?.role !== "admin") {
    return null;
  }

  return <>{children}</>;
}
