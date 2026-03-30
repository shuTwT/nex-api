"use client";

import { useEffect, useState, useTransition } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { 
  Search, 
  Plus, 
  Edit,
  Trash2,
  Key,
  UserIcon,
  Shield,
  Crown,
  Calendar,
  Users as UsersIcon
} from "lucide-react";
import { getUsers, getUserStats } from "@/app/actions/users";
import { UserForm } from "@/components/user-form";
import { DeleteUserDialog } from "@/components/delete-user-dialog";
import { Pagination } from "@/components/pagination";

interface User {
  id: string;
  email: string;
  username: string;
  role: string;
  credits: number;
  createdAt: Date;
  subscription?: {
    planName: string;
    endDate: Date;
  } | null;
}

interface UserStats {
  totalUsers: number;
  activeUsers: number;
  adminUsers: number;
  newUsersThisMonth: number;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRole, setSelectedRole] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [showUserForm, setShowUserForm] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    loadUsers();
    loadStats();
  }, [searchQuery, selectedRole, currentPage, pageSize]);

  async function loadUsers() {
    setIsLoading(true);
    const result = await getUsers({
      role: selectedRole,
      search: searchQuery,
      page: currentPage,
      limit: pageSize,
    });
    
    if (result.success && result.data) {
      setUsers(result.data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }

  async function loadStats() {
    const result = await getUserStats();
    if (result.success && result.data) {
      setStats(result.data);
    }
  }

  function handleAddUser() {
    setEditingUser(null);
    setShowUserForm(true);
  }

  function handleEditUser(user: User) {
    setEditingUser(user);
    setShowUserForm(true);
  }

  function handleDeleteUser(user: User) {
    setDeletingUser(user);
  }

  function handleFormSuccess() {
    startTransition(() => {
      loadUsers();
      loadStats();
    });
  }

  function handleDeleteSuccess() {
    startTransition(() => {
      loadUsers();
      loadStats();
    });
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  function handleSearchChange(value: string) {
    setSearchQuery(value);
    setCurrentPage(1);
  }

  function handleRoleChange(role: string) {
    setSelectedRole(role);
    setCurrentPage(1);
  }

  const statsCards = [
    {
      title: "总用户数",
      value: stats?.totalUsers || 0,
      icon: UsersIcon,
      color: "blue",
    },
    {
      title: "活跃用户",
      value: stats?.activeUsers || 0,
      icon: Shield,
      color: "green",
    },
    {
      title: "管理员",
      value: stats?.adminUsers || 0,
      icon: Crown,
      color: "purple",
    },
    {
      title: "本月新增",
      value: stats?.newUsersThisMonth || 0,
      icon: Calendar,
      color: "cyan",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">用户管理</h1>
          <p className="text-slate-500 mt-1">管理系统用户和权限</p>
        </div>
        <Button 
          className="gap-2 cursor-pointer"
          onClick={handleAddUser}
        >
          <Plus className="h-4 w-4" />
          添加用户
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            blue: "bg-blue-50 text-blue-600",
            green: "bg-green-50 text-green-600",
            purple: "bg-purple-50 text-purple-600",
            cyan: "bg-cyan-50 text-cyan-600",
          };
          
          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow cursor-pointer">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-slate-500">{stat.title}</p>
                    <p className="text-2xl font-bold text-slate-900 mt-1">{stat.value}</p>
                  </div>
                  <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
                    <Icon className="h-5 w-5" />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索用户名或邮箱..."
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex gap-2">
              <Button
                variant={selectedRole === "all" ? "default" : "outline"}
                size="sm"
                onClick={() => handleRoleChange("all")}
                className="cursor-pointer"
              >
                全部
              </Button>
              <Button
                variant={selectedRole === "admin" ? "default" : "outline"}
                size="sm"
                onClick={() => handleRoleChange("admin")}
                className="cursor-pointer"
              >
                管理员
              </Button>
              <Button
                variant={selectedRole === "user" ? "default" : "outline"}
                size="sm"
                onClick={() => handleRoleChange("user")}
                className="cursor-pointer"
              >
                普通用户
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Users Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">用户列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-12">
              <UserIcon className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">没有找到用户</h3>
              <p className="text-slate-500 mb-4">尝试调整搜索条件或添加新用户</p>
              <Button onClick={handleAddUser} className="gap-2 cursor-pointer">
                <Plus className="h-4 w-4" />
                添加用户
              </Button>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">用户信息</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">角色</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">订阅计划</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">剩余积分</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">注册时间</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((user) => (
                      <tr key={user.id} className="border-b border-slate-100 hover:bg-slate-50 transition-colors">
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-3">
                            <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
                              <span className="text-white text-sm font-medium">
                                {user.username.charAt(0).toUpperCase()}
                              </span>
                            </div>
                            <div>
                              <p className="text-sm font-medium text-slate-900">{user.username}</p>
                              <p className="text-xs text-slate-500">{user.email}</p>
                            </div>
                          </div>
                        </td>
                        <td className="py-3 px-4">
                          <Badge 
                            variant="outline" 
                            className={
                              user.role === "admin" 
                                ? "bg-purple-50 text-purple-700 border-purple-200" 
                                : "bg-gray-50 text-gray-700 border-gray-200"
                            }
                          >
                            {user.role === "admin" ? "管理员" : "用户"}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {user.subscription?.planName || "免费版"}
                        </td>
                        <td className="py-3 px-4">
                          <span className="text-sm font-medium text-slate-900">
                            {user.credits.toLocaleString()}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {new Date(user.createdAt).toLocaleDateString("zh-CN")}
                        </td>
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-8 w-8 p-0 cursor-pointer"
                              onClick={() => handleEditUser(user)}
                              title="编辑"
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-8 w-8 p-0 cursor-pointer"
                              title="重置密码"
                            >
                              <Key className="h-4 w-4" />
                            </Button>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-8 w-8 p-0 cursor-pointer text-red-600 hover:text-red-700 hover:bg-red-50"
                              onClick={() => handleDeleteUser(user)}
                              title="删除"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Pagination */}
              {pagination && (
                <div className="mt-4">
                  <Pagination
                    currentPage={pagination.page}
                    totalPages={pagination.totalPages}
                    total={pagination.total}
                    pageSize={pagination.limit}
                    onPageChange={handlePageChange}
                    onPageSizeChange={handlePageSizeChange}
                  />
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* User Form Dialog */}
      <Dialog open={showUserForm} onOpenChange={setShowUserForm}>
        <DialogContent className="max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{editingUser ? "编辑用户" : "添加用户"}</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto">
            <UserForm
              user={editingUser || undefined}
              onClose={() => setShowUserForm(false)}
              onSuccess={handleFormSuccess}
              formId="user-form"
            />
          </div>
          <DialogFooter className="border-t pt-4 mt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowUserForm(false)}
              className="cursor-pointer"
            >
              取消
            </Button>
            <Button type="submit" form="user-form" className="cursor-pointer">
              {editingUser ? "保存" : "创建"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      {deletingUser && (
        <DeleteUserDialog
          open={!!deletingUser}
          onOpenChange={(open) => !open && setDeletingUser(null)}
          userId={deletingUser.id}
          userName={deletingUser.username}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </div>
  );
}
