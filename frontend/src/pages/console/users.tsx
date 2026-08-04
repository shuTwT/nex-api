import { useCallback, useEffect, useState, useTransition } from "react";
import { Avatar, Button, Card, Flex, Input, Modal, Select, Space, Statistic, Table, Tag, Typography } from "antd";
import { Calendar, Crown, Edit, Key, Plus, Search, Shield, Trash2, UserIcon, Users as UsersIcon } from "lucide-react";
import { api, responseData } from "@/lib/api";
import { UserForm } from "@/components/user-form";
import { DeleteUserDialog } from "@/components/delete-user-dialog";

interface User {
  id: string;
  email: string;
  username: string;
  role: string;
  credits: number;
  createdAt: Date;
  subscription?: { planName: string; endDate: Date } | null;
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
  const [searchInput, setSearchInput] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedRole, setAppliedRole] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [showUserForm, setShowUserForm] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [, startTransition] = useTransition();

  const loadUsers = useCallback(async () => {
    setIsLoading(true);
    const result = await api.users_route_get({ role: appliedRole, search: appliedSearch, page: currentPage, limit: pageSize });
    if (result.success) {
      const data = responseData<User[]>(result);
      if (data) setUsers(data);
      if (result.pagination) setPagination(result.pagination);
    }
    setIsLoading(false);
  }, [appliedRole, appliedSearch, currentPage, pageSize]);

  useEffect(() => { void loadUsers(); }, [loadUsers]);

  const loadStats = useCallback(async () => {
    const data = responseData<UserStats>(await api.users_stats_route_get());
    if (data) setStats(data);
  }, []);
  useEffect(() => { void loadStats(); }, [loadStats]);

  function refresh() { startTransition(() => { void loadUsers(); void loadStats(); }); }
  function handleAddUser() { setEditingUser(null); setShowUserForm(true); }
  function handlePageChange(page: number) { setCurrentPage(page); }
  function handlePageSizeChange(size: number) { setPageSize(size); setCurrentPage(1); }
  function handleQuery() { setAppliedSearch(searchInput); setAppliedRole(roleFilter); setCurrentPage(1); }
  function handleReset() { setSearchInput(""); setRoleFilter("all"); setAppliedSearch(""); setAppliedRole("all"); setCurrentPage(1); }

  const statsCards = [
    { title: "总用户数", value: stats?.totalUsers || 0, icon: <UsersIcon size={20} />, color: "#1677ff" },
    { title: "活跃用户", value: stats?.activeUsers || 0, icon: <Shield size={20} />, color: "#52c41a" },
    { title: "管理员", value: stats?.adminUsers || 0, icon: <Crown size={20} />, color: "#722ed1" },
    { title: "本月新增", value: stats?.newUsersThisMonth || 0, icon: <Calendar size={20} />, color: "#13c2c2" },
  ];

  return (
    <Flex vertical gap={24}>
      <Flex justify="space-between" align="center">
        <div><Typography.Title level={2} style={{ margin: 0 }}>用户管理</Typography.Title><Typography.Text type="secondary">管理系统用户和权限</Typography.Text></div>
        <Button type="primary" icon={<Plus size={16} />} onClick={handleAddUser}>添加用户</Button>
      </Flex>

      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => <Card key={stat.title}><Statistic title={stat.title} value={stat.value} prefix={<span style={{ color: stat.color }}>{stat.icon}</span>} /></Card>)}
      </div>

      <Card>
        <Space wrap>
          <Input prefix={<Search size={16} />} placeholder="搜索用户名或邮箱..." value={searchInput} onChange={(event) => setSearchInput(event.target.value)} style={{ width: 280 }} onPressEnter={handleQuery} />
          <Select value={roleFilter} onChange={setRoleFilter} style={{ width: 130 }} options={[{ value: "all", label: "全部" }, { value: "admin", label: "管理员" }, { value: "user", label: "普通用户" }]} />
          <Button type="primary" size="medium" onClick={handleQuery}>查询</Button>
          <Button size="medium" onClick={handleReset}>重置</Button>
        </Space>
      </Card>

      <Card title="用户列表">
        <Table<User>
          rowKey="id"
          loading={isLoading}
          dataSource={users}
          scroll={{ x: 900 }}
          locale={{ emptyText: <Flex vertical align="center" gap={12}><UserIcon size={40} color="#bfbfbf" /><Typography.Text type="secondary">没有找到用户</Typography.Text><Button type="primary" icon={<Plus size={16} />} onClick={handleAddUser}>添加用户</Button></Flex> }}
          pagination={{
            current: pagination?.page ?? currentPage,
            pageSize: pagination?.limit ?? pageSize,
            total: pagination?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, size) => size !== pageSize ? handlePageSizeChange(size) : handlePageChange(page),
          }}
          columns={[
            { title: "用户信息", key: "user", render: (_, user) => <Space><Avatar style={{ background: "linear-gradient(135deg, #1677ff, #13c2c2)" }}>{user.username.charAt(0).toUpperCase()}</Avatar><Flex vertical gap={0}><Typography.Text strong>{user.username}</Typography.Text><Typography.Text type="secondary">{user.email}</Typography.Text></Flex></Space> },
            { title: "角色", dataIndex: "role", render: (role) => <Tag color={role === "admin" ? "purple" : "default"}>{role === "admin" ? "管理员" : "用户"}</Tag> },
            { title: "订阅计划", key: "subscription", render: (_, user) => user.subscription?.planName || "免费版" },
            { title: "剩余积分", dataIndex: "credits", render: (credits) => Number(credits).toLocaleString() },
            { title: "注册时间", dataIndex: "createdAt", render: (createdAt) => new Date(createdAt).toLocaleDateString("zh-CN") },
            { title: "操作", key: "actions", render: (_, user) => <Space size="small"><Button type="text" shape="circle" title="编辑" icon={<Edit size={16} />} onClick={() => { setEditingUser(user); setShowUserForm(true); }} /><Button type="text" shape="circle" title="重置密码" icon={<Key size={16} />} /><Button type="text" danger shape="circle" title="删除" icon={<Trash2 size={16} />} onClick={() => setDeletingUser(user)} /></Space> },
          ]}
        />
      </Card>

      <Modal
        open={showUserForm}
        title={editingUser ? "编辑用户" : "添加用户"}
        onCancel={() => setShowUserForm(false)}
        destroyOnHidden
        width={560}
        okText={editingUser ? "保存" : "创建"}
        okButtonProps={{ htmlType: "submit", form: "user-form" }}
        modalRender={(dom) => <>{dom}</>}
      >
        <UserForm user={editingUser || undefined} onClose={() => setShowUserForm(false)} onSuccess={refresh} formId="user-form" />
      </Modal>

      {deletingUser && <DeleteUserDialog open onOpenChange={(open) => !open && setDeletingUser(null)} userId={deletingUser.id} userName={deletingUser.username} onSuccess={refresh} />}
    </Flex>
  );
}
