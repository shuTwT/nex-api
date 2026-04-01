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
  const { user, isAuthenticated } = useAuth();

  useEffect(() => {

    if (!isAuthenticated) {
      router.push("/unauthorized");
      return;
    }

    const currentItem = consoleMenuItems.find(item => item.href === pathname);
    
    if (currentItem?.adminOnly && user?.role !== "admin") {
      router.push("/forbidden");
      return;
    }
  }, [isAuthenticated, pathname, user, router]);

  if (!isAuthenticated) {
    return null;
  }

  const currentItem = consoleMenuItems.find(item => item.href === pathname);
  if (currentItem?.adminOnly && user?.role !== "admin") {
    return null;
  }

  return <>{children}</>;
}
