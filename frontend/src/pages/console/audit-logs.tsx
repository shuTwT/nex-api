import { useCallback, useEffect, useState } from "react";
import {
  Button,
  Card,
  Flex,
  Input,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import {
  AlertCircle,
  Clock,
  Download,
  FileText,
  Info,
  RefreshCw,
  Search,
  Trash2,
  User,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { DeleteAuditLogDialog } from "@/components/delete-audit-log-dialog";

interface AuditLog {
  id: string;
  userId: string | null;
  user: { id: string; name: string | null; email: string | null } | null;
  action: string;
  resource: string;
  details: string | null;
  ipAddress: string | null;
  userAgent: string | null;
  level: string;
  status: string;
  metadata: string | null;
  createdAt: string;
}
interface AuditLogStats {
  totalLogs: number;
  infoLogs: number;
  warningLogs: number;
  errorLogs: number;
  successLogs: number;
  failedLogs: number;
}
interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

function toRFC3339(value: string) {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toISOString();
}

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [stats, setStats] = useState<AuditLogStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [levelFilter, setLevelFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [startDateInput, setStartDateInput] = useState("");
  const [endDateInput, setEndDateInput] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedLevel, setAppliedLevel] = useState("all");
  const [appliedStatus, setAppliedStatus] = useState("all");
  const [appliedStartDate, setAppliedStartDate] = useState("");
  const [appliedEndDate, setAppliedEndDate] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [deletingAuditLog, setDeletingAuditLog] = useState<AuditLog | null>(
    null,
  );

  const loadLogs = useCallback(async () => {
    setIsLoading(true);
    const result = await api.audit_logs_route_get({
      search: appliedSearch,
      level: appliedLevel,
      status: appliedStatus,
      startDate: toRFC3339(appliedStartDate),
      endDate: toRFC3339(appliedEndDate),
      page: currentPage,
      limit: pageSize,
    });
    if (result.success) {
      const data = responseData<AuditLog[]>(result);
      if (data) setLogs(data);
      if (result.pagination) setPagination(result.pagination);
    }
    setIsLoading(false);
  }, [
    appliedEndDate,
    appliedLevel,
    appliedSearch,
    appliedStartDate,
    appliedStatus,
    currentPage,
    pageSize,
  ]);
  useEffect(() => {
    void loadLogs();
  }, [loadLogs]);
  const loadStats = useCallback(async () => {
    const data = responseData<AuditLogStats>(
      await api.audit_logs_stats_route_get(),
    );
    if (data) setStats(data);
  }, []);
  useEffect(() => {
    void loadStats();
  }, [loadStats]);

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedLevel(levelFilter);
    setAppliedStatus(statusFilter);
    setAppliedStartDate(startDateInput);
    setAppliedEndDate(endDateInput);
    setCurrentPage(1);
  }
  function handleReset() {
    setSearchInput("");
    setLevelFilter("all");
    setStatusFilter("all");
    setStartDateInput("");
    setEndDateInput("");
    setAppliedSearch("");
    setAppliedLevel("all");
    setAppliedStatus("all");
    setAppliedStartDate("");
    setAppliedEndDate("");
    setCurrentPage(1);
  }
  function refresh() {
    void loadLogs();
    void loadStats();
  }
  async function handleExport() {
    const result = await api.audit_logs_export_route_get({
      level: levelFilter,
      status: statusFilter,
      startDate: toRFC3339(startDateInput),
      endDate: toRFC3339(endDateInput),
    });
    if (!result.success) {
      alert(result.error || "导出失败");
      return;
    }
    const data = responseData<string>(result);
    if (data === null) return;
    const url = URL.createObjectURL(
      new Blob([data], { type: "text/csv;charset=utf-8;" }),
    );
    const link = document.createElement("a");
    link.href = url;
    link.download = `audit-logs-${new Date().toISOString()}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  }

  const statsCards = [
    {
      title: "总日志",
      value: stats?.totalLogs || 0,
      icon: <FileText size={20} />,
      color: "#1677ff",
    },
    {
      title: "信息",
      value: stats?.infoLogs || 0,
      icon: <Info size={20} />,
      color: "#1677ff",
    },
    {
      title: "警告",
      value: stats?.warningLogs || 0,
      icon: <AlertCircle size={20} />,
      color: "#fa8c16",
    },
    {
      title: "错误",
      value: stats?.errorLogs || 0,
      icon: <AlertCircle size={20} />,
      color: "#ff4d4f",
    },
  ];

  return (
    <Flex vertical gap={24}>
      <Flex justify="space-between" align="center">
        <div>
          <Typography.Title level={2} style={{ margin: 0 }}>
            审计日志
          </Typography.Title>
          <Typography.Text type="secondary">
            查看系统操作记录和安全日志
          </Typography.Text>
        </div>
        <Space>
          <Button
            icon={<RefreshCw size={16} />}
            onClick={() => void loadLogs()}
          >
            刷新
          </Button>
          <Button
            icon={<Download size={16} />}
            onClick={() => void handleExport()}
          >
            导出
          </Button>
        </Space>
      </Flex>
      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => (
          <Card key={stat.title}>
            <Statistic
              title={stat.title}
              value={stat.value}
              prefix={<span style={{ color: stat.color }}>{stat.icon}</span>}
            />
          </Card>
        ))}
      </div>
      <Card>
        <Space wrap>
          <Input
            prefix={<Search size={16} />}
            placeholder="搜索操作、资源或详情..."
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            style={{ width: 250 }}
            onPressEnter={handleQuery}
          />
          <Select
            value={levelFilter}
            onChange={setLevelFilter}
            style={{ width: 120 }}
            options={[
              { value: "all", label: "全部级别" },
              { value: "info", label: "信息" },
              { value: "warning", label: "警告" },
              { value: "error", label: "错误" },
            ]}
          />
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 110 }}
            options={[
              { value: "all", label: "全部" },
              { value: "success", label: "成功" },
              { value: "error", label: "失败" },
            ]}
          />
          <Input
            type="datetime-local"
            value={startDateInput}
            onChange={(event) => setStartDateInput(event.target.value)}
            style={{ width: 210 }}
            aria-label="开始时间"
          />
          <Input
            type="datetime-local"
            value={endDateInput}
            onChange={(event) => setEndDateInput(event.target.value)}
            style={{ width: 210 }}
            aria-label="结束时间"
          />
          <Button type="primary" size="medium" onClick={handleQuery}>
            查询
          </Button>
          <Button size="medium" onClick={handleReset}>重置</Button>
        </Space>
      </Card>
      <Card title="日志列表">
        <Table<AuditLog>
          rowKey="id"
          loading={isLoading}
          dataSource={logs}
          scroll={{ x: 1350 }}
          locale={{
            emptyText: (
              <Flex vertical align="center" gap={12}>
                <FileText size={40} color="#bfbfbf" />
                <Typography.Text type="secondary">
                  没有找到审计日志
                </Typography.Text>
              </Flex>
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
                setPageSize(size);
                setCurrentPage(1);
              } else setCurrentPage(page);
            },
          }}
          columns={[
            {
              title: "时间",
              dataIndex: "createdAt",
              width: 180,
              render: (createdAt) => (
                <Space size={4}>
                  <Clock size={16} color="#8c8c8c" />
                  {new Date(createdAt).toLocaleString("zh-CN")}
                </Space>
              ),
            },
            {
              title: "用户",
              key: "user",
              width: 170,
              render: (_, log) => (
                <Space size={4}>
                  <User size={16} color="#8c8c8c" />
                  {log.user?.email || log.user?.name || "系统"}
                </Space>
              ),
            },
            { title: "操作", dataIndex: "action", width: 150 },
            { title: "资源", dataIndex: "resource", width: 140 },
            {
              title: "详情",
              dataIndex: "details",
              width: 220,
              ellipsis: true,
              render: (details) => (
                <Tooltip title={details || "-"}>{details || "-"}</Tooltip>
              ),
            },
            {
              title: "IP 地址",
              dataIndex: "ipAddress",
              width: 130,
              render: (ipAddress) => (
                <Typography.Text code>{ipAddress || "-"}</Typography.Text>
              ),
            },
            {
              title: "级别",
              dataIndex: "level",
              width: 90,
              render: (level) => (
                <Tag
                  color={
                    level === "info"
                      ? "blue"
                      : level === "warning"
                        ? "orange"
                        : "error"
                  }
                >
                  {level === "info"
                    ? "信息"
                    : level === "warning"
                      ? "警告"
                      : "错误"}
                </Tag>
              ),
            },
            {
              title: "状态",
              dataIndex: "status",
              width: 90,
              render: (status) => (
                <Tag color={status === "success" ? "success" : "error"}>
                  {status === "success" ? "成功" : "失败"}
                </Tag>
              ),
            },
            {
              title: "操作",
              key: "actions",
              fixed: "right",
              width: 70,
              render: (_, log) => (
                <Button
                  type="text"
                  danger
                  shape="circle"
                  title="删除"
                  icon={<Trash2 size={16} />}
                  onClick={() => setDeletingAuditLog(log)}
                />
              ),
            },
          ]}
        />
      </Card>
      {deletingAuditLog && (
        <DeleteAuditLogDialog
          open
          onOpenChange={(open) => !open && setDeletingAuditLog(null)}
          auditLogId={deletingAuditLog.id}
          auditLogAction={deletingAuditLog.action}
          onSuccess={refresh}
        />
      )}
    </Flex>
  );
}
