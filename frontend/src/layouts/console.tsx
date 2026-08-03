import { useEffect, useState } from "react";
import { Link, useLocation, useNavigate, Outlet } from "react-router";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { consoleMenuItems } from "@/config/console-menu";
import { useAuth } from "@/hooks/use-auth";

export function ConsoleLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const pathname = location.pathname;
  const { user, isAuthenticated, isLoading } = useAuth();
  const isAdmin = user?.role === "admin";

  // Guard: redirect unauthenticated users to /unauthorized,
  // and non-admins away from admin-only pages to /forbidden.
  useEffect(() => {
    if (isLoading) return;

    if (!isAuthenticated) {
      navigate("/unauthorized", { replace: true });
      return;
    }

    const currentItem = consoleMenuItems.find((item) => item.href === pathname);
    if (currentItem?.adminOnly && !isAdmin) {
      navigate("/forbidden", { replace: true });
    }
  }, [isAuthenticated, isLoading, isAdmin, pathname, navigate]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  const currentItem = consoleMenuItems.find((item) => item.href === pathname);
  if (currentItem?.adminOnly && !isAdmin) {
    return null;
  }

  const filteredMenuItems = consoleMenuItems.filter(
    (item) => !item.adminOnly || isAdmin,
  );

  return (
    <div className="flex-1 flex justify-center">
      <div className="relative container">
        {/* Sidebar */}
        <aside
          className={`absolute left-0 top-0 z-40 bottom-0 bg-white dark:bg-black border-r border-slate-200 dark:border-slate-700 transition-all duration-300 ${
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
                  to={item.href}
                  className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 cursor-pointer ${
                    isActive
                      ? "bg-blue-50 dark:bg-blue-700 text-blue-600 dark:text-white font-medium"
                      : "text-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700 hover:text-slate-900"
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
        <section
          className={`flex-1 transition-all duration-300 ${
            collapsed ? "ml-16" : "ml-64"
          }`}
        >
          <div className="p-8">
            <Outlet />
          </div>
        </section>
      </div>
    </div>
  );
}