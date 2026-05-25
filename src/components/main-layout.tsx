"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Zap, LogOut } from "lucide-react";
import { AuthModal } from "@/components/auth-modal";
import { useAuth } from "@/hooks/use-auth";
import { ThemeToggle } from "@/components/theme-toggle";

interface MainLayoutProps {
  children: React.ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const pathname = usePathname();
  const { user, isAuthenticated, logout } = useAuth();

  const isActive = (path: string) => {
    if (path === "/console") {
      return pathname.startsWith("/console");
    }
    return pathname === path;
  };

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-white/95 backdrop-blur supports-[backdrop-filter]:bg-white/60">
        <div className="container relative flex h-14 items-center px-4 md:px-6 mx-auto">
          <Link href="/" className="absolute left-4 flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-blue-600 flex items-center justify-center">
              <Zap className="h-5 w-5 text-white" />
            </div>
            <span className="text-lg font-semibold">NexApi 聚合管理系统</span>
          </Link>
          <nav className="flex items-center gap-6 mx-auto">
            <Link 
              href="/" 
              className={`text-sm font-medium transition-colors ${
                isActive("/") ? "text-blue-600" : "hover:text-blue-600"
              }`}
            >
              首页
            </Link>
            <Link 
              href="/api-market" 
              className={`text-sm font-medium transition-colors ${
                isActive("/api-market") ? "text-blue-600" : "hover:text-blue-600"
              }`}
            >
              API 市场
            </Link>
            <Link 
              href="/pricing" 
              className={`text-sm font-medium transition-colors ${
                isActive("/pricing") ? "text-blue-600" : "hover:text-blue-600"
              }`}
            >
              定价
            </Link>
            <Link 
              href="/console" 
              className={`text-sm font-medium transition-colors ${
                isActive("/console") ? "text-blue-600" : "hover:text-blue-600"
              }`}
            >
              控制台
            </Link>
          </nav>
          <div className="absolute right-4 flex items-center gap-4">
            <ThemeToggle />
            {isAuthenticated && user ? (
              <>
                <div className="flex items-center gap-2">
                  <div className="h-8 w-8 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
                    <span className="text-white text-sm font-medium">
                      {user.username.charAt(0).toUpperCase()}
                    </span>
                  </div>
                  <span className="text-sm font-medium text-slate-700">{user.username}</span>
                </div>
                <Button 
                  variant="ghost" 
                  size="sm"
                  onClick={handleLogout}
                  className="cursor-pointer"
                >
                  <LogOut className="h-4 w-4 mr-1" />
                  退出
                </Button>
              </>
            ) : (
              <Button 
                variant="ghost" 
                size="sm"
                onClick={() => setAuthModalOpen(true)}
                className="cursor-pointer"
              >
                登录
              </Button>
            )}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1">
        {children}
      </main>

      {/* Footer */}
      <footer className="border-t bg-white dark:bg-black">
        <div className="container px-4 py-8 md:px-6 mx-auto">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <div className="h-6 w-6 rounded bg-blue-600 flex items-center justify-center">
                <Zap className="h-4 w-4 text-white" />
              </div>
              <span className="text-sm font-semibold">NexApi 聚合管理系统</span>
            </div>
            <p className="text-sm text-gray-500">
              © 2026 NexApi. All rights reserved.
            </p>
          </div>
        </div>
      </footer>

      {/* Auth Modal */}
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} />
    </div>
  );
}
