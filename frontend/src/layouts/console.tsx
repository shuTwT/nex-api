import { useEffect, useMemo, useState } from "react";
import { Link, Outlet, useLocation, useNavigate } from "react-router";
import { Button, Layout, Menu } from "antd";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { consoleMenuItems } from "@/config/console-menu";
import { useAuth } from "@/hooks/use-auth";
import { ConsolePageLoading } from "@/components/console-page-loading";

const { Content, Sider } = Layout;

export function ConsoleLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const pathname = location.pathname;
  const { user, isAuthenticated, isLoading } = useAuth();
  const isAdmin = user?.role === "admin";

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

  const menuItems = useMemo(
    () => consoleMenuItems
      .filter((item) => !item.adminOnly || isAdmin)
      .map((item) => ({
        key: item.href,
        icon: <item.icon size={18} />,
        label: <Link to={item.href}>{item.name}</Link>,
      })),
    [isAdmin],
  );

  if (isLoading) {
    return <ConsolePageLoading fullHeight />;
  }

  if (!isAuthenticated) return null;

  const currentItem = consoleMenuItems.find((item) => item.href === pathname);
  if (currentItem?.adminOnly && !isAdmin) return null;

  return (
    <Layout className="min-h-[calc(100dvh-3.5rem)]">
      <Sider
        breakpoint="lg"
        collapsedWidth={64}
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="light"
        trigger={null}
        width={256}
      >
        <div className="flex h-full min-h-0 flex-col">
          <div className="flex h-12 shrink-0 items-center justify-end px-2">
            <Button
              type="text"
              size="large"
              icon={collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
              aria-label={collapsed ? "展开侧边栏" : "收起侧边栏"}
              onClick={() => setCollapsed((current) => !current)}
            >
              {!collapsed && "收起侧边栏"}
            </Button>
          </div>
          <Menu
            mode="inline"
            selectedKeys={[pathname]}
            items={menuItems}
            className="min-h-0 flex-1 overflow-y-auto border-e-0"
          />
        </div>
      </Sider>
      <Layout>
        <Content className="p-4 md:p-8">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
