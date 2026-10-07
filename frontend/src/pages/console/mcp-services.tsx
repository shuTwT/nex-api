import { useEffect, useState, useCallback, useTransition } from "react";
import {
  Alert,
  Button,
  Card,
  Input,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
  type TableColumnsType,
} from "antd";
import {
  Search,
  Plus,
  Edit,
  Trash2,
  Plug,
  Activity,
  Pause,
  Database,
} from "lucide-react";
import { toast } from "sonner";
import { api, responseData } from "@/lib/api";
import { McpFormDialog, type McpServiceData } from "@/components/mcp-form-dialog";

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

interface McpStats {
  totalServices: number;
  activeServices: number;
  inactiveServices: number;
  totalCalls: number;
}

interface Category {
  id: string;
  name: string;
  apiCount: number;
}

const TYPE_LABELS: Record<string, string> = {
  stdio: "stdio",
  sse: "SSE",
  streamableHttp: "Streamable HTTP",
};

const TYPE_TAG_COLORS: Record<string, string> = {
  stdio: "default",
  sse: "blue",
  streamableHttp: "purple",
};

export default function McpServicesPage() {
  const [services, setServices] = useState<McpServiceData[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [stats, setStats] = useState<McpStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedType, setAppliedType] = useState("all");
  const [appliedCategory, setAppliedCategory] = useState("all");
  const [appliedStatus, setAppliedStatus] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [categoriesLoading, setCategoriesLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingService, setEditingService] = useState<McpServiceData | null>(null);

  const loadServices = useCallback(async () => {
    setIsLoading(true);
    setLoadError(null);
    const query: Record<string, string | number | boolean> = {
      type: appliedType,
      category: appliedCategory,
      search: appliedSearch,
      status: appliedStatus,
      page: currentPage,
      limit: pageSize,
    };
    const result = await api.mcp_services_route_get(query);

    if (result.success) {
      const data = responseData<McpServiceData[]>(result);
      if (data) setServices(data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    } else {
      setServices([]);
      setPagination(null);
      setLoadError(result.error || "MCP 服务列表加载失败");
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedType, appliedCategory, appliedStatus]);

  useEffect(() => {
    loadServices();
  }, [loadServices]);

  async function loadStats() {
    const result = await api.mcp_services_stats_route_get();
    const data = responseData<McpStats>(result);
    if (data) setStats(data);
  }

  async function loadCategories() {
    setCategoriesLoading(true);
    const result = await api.categories_route_get();
    const data = responseData<Category[]>(result);
    if (data) setCategories(data);
    setCategoriesLoading(false);
  }

  useEffect(() => {
    loadCategories();
    loadStats();
  }, []);

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedType(typeFilter);
    setAppliedCategory(categoryFilter);
    setAppliedStatus(statusFilter);
    setCurrentPage(1);
  }

  function handleReset() {
    setSearchInput("");
    setTypeFilter("all");
    setCategoryFilter("all");
    setStatusFilter("all");
    setAppliedSearch("");
    setAppliedType("all");
    setAppliedCategory("all");
    setAppliedStatus("all");
    setCurrentPage(1);
  }

  function handleToggleStatus(service: McpServiceData) {
    setTogglingId(service.id);
    startTransition(async () => {
      const result = await api.mcp_services_id_toggle_route_put({ id: service.id });
      if (result.success) {
        loadServices();
        loadStats();
      } else {
        toast.error(result.error || "切换状态失败");
      }
      setTogglingId(null);
    });
  }

  function handleDelete(service: McpServiceData) {
    startTransition(async () => {
      const result = await api.mcp_services_id_route_delete({ id: service.id });
      if (result.success) {
        toast.success("MCP 服务已删除");
        loadServices();
        loadStats();
      } else {
        toast.error(result.error || "删除失败");
      }
    });
  }

  function handleAdd() {
    setEditingService(null);
    setShowForm(true);
  }

  function handleEdit(service: McpServiceData) {
    setEditingService(service);
    setShowForm(true);
  }

  async function handleFormSuccess() {
    await loadServices();
    await loadStats();
  }

  const statsCards = [
    {
      title: "服务总数",
      value: stats?.totalServices || 0,
      icon: Plug,
      color: "blue",
    },
    {
      title: "活跃服务",
      value: stats?.activeServices || 0,
      icon: Activity,
      color: "green",
    },
    {
      title: "停用服务",
      value: stats?.inactiveServices || 0,
      icon: Pause,
      color: "orange",
    },
    {
      title: "总调用次数",
      value: (stats?.totalCalls || 0).toLocaleString(),
      icon: Database,
      color: "purple",
    },
  ];

  const categoryOptions = [
    { value: "all", label: "全部分类" },
    ...categories.map((category) => ({ value: category.id, label: category.name })),
  ];

  const columns: TableColumnsType<McpServiceData> = [
    {
      title: "服务信息",
      dataIndex: "name",
      key: "name",
      width:120,
      render: (_value, svc) => (
        <div className="max-w-[280px]">
          <Typography.Text strong ellipsis style={{ maxWidth: "100%" }} className="block">
            {svc.name}
          </Typography.Text>
          {svc.description && (
            <Tooltip title={svc.description}>
              <Typography.Text type="secondary" ellipsis style={{ maxWidth: "100%" }} className="mt-1 block">
                {svc.description}
              </Typography.Text>
            </Tooltip>
          )}
        </div>
      ),
    },
    {
      title: "标识",
      dataIndex: "identifier",
      key: "identifier",
      width:120,
      render: (identifier) => <Typography.Text code>{identifier}</Typography.Text>,
    },
    {
      title: "类型",
      dataIndex: "type",
      key: "type",
      width:220,
      render: (type) => (
        <Tag color={TYPE_TAG_COLORS[type] || "default"}>{TYPE_LABELS[type] || type}</Tag>
      ),
    },
    {
      title: "分类",
      key: "category",
      width:90,
      render: (_value, svc) => svc.category?.name || "未分类",
    },
    {
      title: "端点",
      key: "endpoint",
      width:280,
      render: (_value, svc) => {
        const target = svc.type === "stdio" ? svc.command : svc.endpoint;
        if (!target) {
          return <Typography.Text type="secondary">-</Typography.Text>;
        }
        return (
          <Tooltip title={target}>
            <Typography.Text type="secondary" code ellipsis style={{ maxWidth: 220 }}>
              {target}
            </Typography.Text>
          </Tooltip>
        );
      },
    },
    {
      title: "定价",
      dataIndex: "pricing",
      key: "pricing",
      width:120
    },
    {
      title: "状态",
      dataIndex: "isActive",
      key: "isActive",
      width:120,
      render: (isActive, svc) => (
        <Switch
          checked={isActive}
          loading={togglingId === svc.id}
          onChange={() => handleToggleStatus(svc)}
        />
      ),
    },
    {
      title: "调用次数",
      dataIndex: "callCount",
      key: "callCount",
      width:120,
      render: (callCount) => callCount.toLocaleString(),
    },
    {
      title: "操作",
      key: "actions",
      fixed: "right",
      width:120,
      render: (_value, svc) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<Edit size={16} />}
            aria-label={`编辑 ${svc.name}`}
            onClick={() => handleEdit(svc)}
          />
          <Popconfirm
            title="删除 MCP 服务"
            description="确定要删除这个 MCP 服务吗？此操作无法撤销。"
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
            onConfirm={() => handleDelete(svc)}
          >
            <Button
              danger
              type="text"
              size="small"
              icon={<Trash2 size={16} />}
              aria-label={`删除 ${svc.name}`}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  void isPending;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">MCP 服务管理</h1>
          <p className="text-slate-500 mt-1">管理 MCP（Model Context Protocol）服务</p>
        </div>
        <Button className="gap-2 cursor-pointer" onClick={handleAdd}>
          <Plus className="h-4 w-4" />
          添加服务
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses: Record<string, string> = {
            blue: "bg-blue-50 text-blue-600",
            green: "bg-green-50 text-green-600",
            orange: "bg-orange-50 text-orange-600",
            purple: "bg-purple-50 text-purple-600",
          };

          return (
            <Card
              key={stat.title}
              className="hover:shadow-md transition-shadow cursor-pointer"
            >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-slate-500">{stat.title}</p>
                    <p className="text-2xl font-bold text-slate-900 mt-1">
                      {stat.value}
                    </p>
                  </div>
                  <div
                    className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color]}`}
                  >
                    <Icon className="h-5 w-5" />
                  </div>
                </div>
            </Card>
          );
        })}
      </div>

      <Card>

          <div className="flex flex-wrap items-end gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索服务名称或标识..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={typeFilter} onChange={setTypeFilter} className="w-[160px]" options={[{ value: "all", label: "全部类型" }, { value: "stdio", label: "stdio" }, { value: "sse", label: "SSE" }, { value: "streamableHttp", label: "Streamable HTTP" }]} />
            <Select value={categoryFilter} onChange={setCategoryFilter} className="w-[140px]" options={categoryOptions} />
            <Select value={statusFilter} onChange={setStatusFilter} className="w-[130px]" options={[{ value: "all", label: "全部" }, { value: "active", label: "已启用" }, { value: "inactive", label: "已停用" }]} />
            <Button size="medium" onClick={handleQuery} className="cursor-pointer">
              查询
            </Button>
            <Button
              type="default"
              size="medium"
              onClick={handleReset}
              className="cursor-pointer"
            >
              重置
            </Button>
          </div>
      </Card>

      <Card>
          <Typography.Title level={5}>服务列表</Typography.Title>
          {loadError && (
            <Alert
              type="error"
              showIcon
              className="mb-4"
              message="服务列表加载失败"
              description={loadError}
              action={<Button size="small" onClick={() => void loadServices()}>重试</Button>}
            />
          )}
          <Table
            rowKey="id"
            columns={columns}
            dataSource={services}
            loading={isLoading}
            scroll={{ x: 1100 }}
            locale={{
              emptyText: loadError ? (
                "加载失败，请重试"
              ) : (
                <div className="py-8 text-center">
                  <Plug className="mx-auto mb-4 h-12 w-12 text-slate-300" />
                  <h3 className="mb-2 text-lg font-medium text-slate-900">
                    没有找到 MCP 服务
                  </h3>
                  <p className="mb-4 text-slate-500">尝试调整搜索条件或添加新服务</p>
                  <Button className="gap-2 cursor-pointer" onClick={handleAdd}>
                    <Plus className="h-4 w-4" />
                    添加服务
                  </Button>
                </div>
              ),
            }}
            pagination={{
              current: pagination?.page ?? currentPage,
              pageSize: pagination?.limit ?? pageSize,
              total: pagination?.total ?? 0,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
              onChange: (page, size) => {
                if (size !== pageSize) {
                  handlePageSizeChange(size);
                } else if (page !== currentPage) {
                  handlePageChange(page);
                }
              },
            }}
          />

      </Card>

      <McpFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        service={editingService}
        categories={categories.map((category) => ({ id: category.id, name: category.name }))}
        categoriesLoading={categoriesLoading}
        onSuccess={handleFormSuccess}
      />
    </div>
  );
}
