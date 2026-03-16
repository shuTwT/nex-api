"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { 
  ChevronLeft,
  ChevronRight
} from "lucide-react";
import { consoleMenuItems } from "@/config/console-menu";
import { useAuthStore } from "@/stores/auth-store";

interface ConsoleLayoutProps {
  children: React.ReactNode;
}

export function ConsoleLayout({ children }: ConsoleLayoutProps) {
  const [collapsed, setCollapsed] = useState(false);
  const pathname = usePathname();
  const { user } = useAuthStore();
  const isAdmin = user?.role === "admin";

  const filteredMenuItems = consoleMenuItems.filter(item => !item.adminOnly || isAdmin);

  return (
    <div className="flex-1 flex">
      {/* Sidebar */}
      <aside
        className={`fixed left-0 top-14 z-40 h-[calc(100vh-3.5rem)] bg-white border-r border-slate-200 transition-all duration-300 ${
          collapsed ? "w-16" : "w-64"
        }`}
      >
        {/* Navigation */}
        <nav className="flex-1 px-3 py-4 space-y-1">
          {filteredMenuItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 cursor-pointer ${
                  isActive
                    ? "bg-blue-50 text-blue-600 font-medium"
                    : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                }`}
              >
                <item.icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span className="text-sm">{item.name}</span>}
              </Link>
            );
          })}
        </nav>

        {/* Collapse Button */}
        <div className="absolute top-4 -right-3">
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="h-6 w-6 bg-white border border-slate-200 rounded-full flex items-center justify-center shadow-sm hover:bg-slate-50 transition-colors cursor-pointer"
          >
            {collapsed ? (
              <ChevronRight className="h-3 w-3 text-slate-600" />
            ) : (
              <ChevronLeft className="h-3 w-3 text-slate-600" />
            )}
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main
        className={`flex-1 transition-all duration-300 ${
          collapsed ? "ml-16" : "ml-64"
        }`}
      >
        <div className="p-8">
          {children}
        </div>
      </main>
    </div>
  );
}
