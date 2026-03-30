import { 
  LayoutDashboard, 
  Crown, 
  Settings, 
  BarChart3, 
  FileText, 
  Users,
  Key,
  Sliders,
  User,
  CreditCard,
  LucideIcon
} from "lucide-react";

export interface MenuItem {
  name: string;
  icon: LucideIcon;
  href: string;
  adminOnly: boolean;
}

export const consoleMenuItems: MenuItem[] = [
  {
    name: "概览",
    icon: LayoutDashboard,
    href: "/console",
    adminOnly: false,
  },
  {
    name: "个人中心",
    icon: User,
    href: "/console/personal",
    adminOnly: false,
  },
  {
    name: "我的会员",
    icon: Crown,
    href: "/console/membership",
    adminOnly: false,
  },
  {
    name: "令牌管理",
    icon: Key,
    href: "/console/tokens",
    adminOnly: false,
  },
  {
    name: "接口管理",
    icon: Settings,
    href: "/console/api-management",
    adminOnly: true,
  },
  {
    name: "订阅计划",
    icon: CreditCard,
    href: "/console/subscription-plans",
    adminOnly: true,
  },
  {
    name: "用量统计",
    icon: BarChart3,
    href: "/console/usage",
    adminOnly: false,
  },
  {
    name: "审计日志",
    icon: FileText,
    href: "/console/audit-logs",
    adminOnly: true,
  },
  {
    name: "用户管理",
    icon: Users,
    href: "/console/users",
    adminOnly: true,
  },
  {
    name: "系统设置",
    icon: Sliders,
    href: "/console/settings",
    adminOnly: true,
  },
];
