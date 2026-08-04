import { useCallback, useEffect, useState } from "react";
import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tooltip,
  Typography,
  type TableColumnsType,
} from "antd";
import { CalendarClock, Edit, Play, Plus, RefreshCw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { api, responseData } from "@/lib/api";

interface RuntimeInfo {
  scheduled: boolean;
  running: boolean;
  nextRun?: string;
}

interface ScheduledJob {
  id: string;
  name: string;
  taskKey: string;
  scheduleType: "cron" | "duration";
  expression: string;
  enabled: boolean;
  description?: string;
  lastRunAt?: string;
  lastStatus: "never" | "success" | "failed";
  lastError?: string;
  createdAt: string;
  updatedAt: string;
  runtime: RuntimeInfo;
}

interface TaskDescriptor {
  key: string;
  description: string;
}

interface ScheduledJobInput {
  name: string;
  taskKey: string;
  scheduleType: "cron" | "duration";
  expression: string;
  enabled: boolean;
  description?: string;
}

const scheduleTypeOptions = [
  { value: "duration", label: "固定间隔" },
  { value: "cron", label: "Cron 表达式" },
];

const statusLabels: Record<ScheduledJob["lastStatus"], string> = {
  never: "未执行",
  success: "成功",
  failed: "失败",
};

const statusColors: Record<ScheduledJob["lastStatus"], "default" | "success" | "error"> = {
  never: "default",
  success: "success",
  failed: "error",
};

function formatDate(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(date);
}

function inputFromJob(job: ScheduledJob): ScheduledJobInput {
  return {
    name: job.name,
    taskKey: job.taskKey,
    scheduleType: job.scheduleType,
    expression: job.expression,
    enabled: job.enabled,
    description: job.description,
  };
}

export default function ScheduledJobsPage() {
  const [jobs, setJobs] = useState<ScheduledJob[]>([]);
  const [tasks, setTasks] = useState<TaskDescriptor[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [actionJobID, setActionJobID] = useState<string | null>(null);
  const [editingJob, setEditingJob] = useState<ScheduledJob | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm<ScheduledJobInput>();
  const scheduleType = Form.useWatch("scheduleType", form) ?? "duration";

  const loadPage = useCallback(async (showSpinner = true) => {
    if (showSpinner) setLoading(true);
    const [jobsResult, tasksResult] = await Promise.all([
      api.scheduled_jobs_route_get(),
      api.scheduled_jobs_tasks_route_get(),
    ]);

    if (jobsResult.success) {
      setJobs(responseData<ScheduledJob[]>(jobsResult) ?? []);
    } else {
      toast.error(jobsResult.error || "加载定时任务失败");
    }

    if (tasksResult.success) {
      setTasks(responseData<TaskDescriptor[]>(tasksResult) ?? []);
    } else {
      toast.error(tasksResult.error || "加载可用任务失败");
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  async function handleRefresh() {
    setRefreshing(true);
    await loadPage(false);
    setRefreshing(false);
  }

  function openCreateModal() {
    setEditingJob(null);
    setModalOpen(true);
  }

  function openEditModal(job: ScheduledJob) {
    setEditingJob(job);
    setModalOpen(true);
  }

  function closeModal() {
    setModalOpen(false);
    setEditingJob(null);
    form.resetFields();
  }

  async function handleSubmit(values: ScheduledJobInput) {
    setSaving(true);
    const result = editingJob
      ? await api.scheduled_jobs_id_route_put({ id: editingJob.id }, values)
      : await api.scheduled_jobs_route_post(values);
    setSaving(false);

    if (!result.success) {
      toast.error(result.error || "保存定时任务失败");
      return;
    }
    toast.success(editingJob ? "定时任务已更新" : "定时任务已创建");
    closeModal();
    await loadPage(false);
  }

  async function handleEnabledChange(job: ScheduledJob, enabled: boolean) {
    setActionJobID(job.id);
    const result = await api.scheduled_jobs_id_route_put(
      { id: job.id },
      { ...inputFromJob(job), enabled },
    );
    setActionJobID(null);
    if (!result.success) {
      toast.error(result.error || "更新任务状态失败");
      return;
    }
    toast.success(enabled ? "任务已启用" : "任务已停用");
    await loadPage(false);
  }

  async function handleRun(job: ScheduledJob) {
    setActionJobID(job.id);
    const result = await api.scheduled_jobs_id_run_route_post({ id: job.id });
    setActionJobID(null);
    if (!result.success) {
      toast.error(result.error || "触发任务失败");
      return;
    }
    toast.success("任务已触发，请稍后刷新查看结果");
    await loadPage(false);
  }

  async function handleDelete(job: ScheduledJob) {
    setActionJobID(job.id);
    const result = await api.scheduled_jobs_id_route_delete({ id: job.id });
    setActionJobID(null);
    if (!result.success) {
      toast.error(result.error || "删除定时任务失败");
      return;
    }
    toast.success("定时任务已删除");
    await loadPage(false);
  }

  const columns: TableColumnsType<ScheduledJob> = [
    {
      title: "任务",
      dataIndex: "name",
      key: "name",
      render: (_name, job) => (
        <div>
          <Typography.Text strong>{job.name}</Typography.Text>
          {job.description && <Typography.Paragraph type="secondary" className="mb-0 mt-1">{job.description}</Typography.Paragraph>}
        </div>
      ),
    },
    {
      title: "任务标识",
      dataIndex: "taskKey",
      key: "taskKey",
      render: (taskKey) => <Typography.Text code>{taskKey}</Typography.Text>,
    },
    {
      title: "调度规则",
      key: "schedule",
      render: (_value, job) => (
        <Space direction="vertical" size={2}>
          <Badge color="blue" text={job.scheduleType === "cron" ? "Cron 表达式" : "固定间隔"} />
          <Typography.Text code>{job.expression}</Typography.Text>
        </Space>
      ),
    },
    {
      title: "运行状态",
      key: "runtime",
      render: (_value, job) => (
        <Space direction="vertical" size={2}>
          <Badge status={statusColors[job.lastStatus]} text={statusLabels[job.lastStatus]} />
          {job.runtime.running && <Badge status="processing" text="运行中" />}
          {job.runtime.nextRun && <Typography.Text type="secondary">下次：{formatDate(job.runtime.nextRun)}</Typography.Text>}
          {job.lastError && <Tooltip title={job.lastError}><Typography.Text type="danger" ellipsis>最近执行失败</Typography.Text></Tooltip>}
        </Space>
      ),
    },
    {
      title: "上次执行",
      dataIndex: "lastRunAt",
      key: "lastRunAt",
      render: (lastRunAt) => formatDate(lastRunAt),
    },
    {
      title: "启用",
      dataIndex: "enabled",
      key: "enabled",
      render: (enabled, job) => (
        <Switch
          checked={enabled}
          loading={actionJobID === job.id}
          onChange={(checked) => void handleEnabledChange(job, checked)}
        />
      ),
    },
    {
      title: "操作",
      key: "actions",
      fixed: "right",
      render: (_value, job) => (
        <Space size="small">
          <Tooltip title={job.enabled ? "立即执行" : "请先启用任务"}>
            <Button
              type="text"
              icon={<Play size={16} />}
              aria-label={`立即执行 ${job.name}`}
              disabled={!job.enabled || actionJobID === job.id}
              onClick={() => void handleRun(job)}
            />
          </Tooltip>
          <Button
            type="text"
            icon={<Edit size={16} />}
            aria-label={`编辑 ${job.name}`}
            disabled={actionJobID === job.id}
            onClick={() => openEditModal(job)}
          />
          <Popconfirm
            title="删除定时任务"
            description={`确定删除“${job.name}”吗？`}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
            onConfirm={() => handleDelete(job)}
          >
            <Button
              danger
              type="text"
              icon={<Trash2 size={16} />}
              aria-label={`删除 ${job.name}`}
              disabled={actionJobID === job.id}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <Typography.Title level={2} className="mb-1">定时任务管理</Typography.Title>
          <Typography.Text type="secondary">配置内置任务的执行计划，并查看近期运行状态。</Typography.Text>
        </div>
        <Space>
          <Button icon={<RefreshCw size={16} />} loading={refreshing} onClick={() => void handleRefresh()}>刷新</Button>
          <Button type="primary" icon={<Plus size={16} />} onClick={openCreateModal}>新建任务</Button>
        </Space>
      </div>

      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={jobs}
          loading={loading}
          pagination={false}
          scroll={{ x: 1080 }}
          locale={{ emptyText: "暂无定时任务" }}
        />
      </Card>

      <Modal
        open={modalOpen}
        title={editingJob ? "编辑定时任务" : "新建定时任务"}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnHidden
        onCancel={closeModal}
        onOk={() => void form.submit()}
      >
        <Form<ScheduledJobInput>
          key={editingJob?.id ?? "new"}
          form={form}
          layout="vertical"
          initialValues={editingJob ? inputFromJob(editingJob) : {
            name: "",
            taskKey: tasks[0]?.key,
            scheduleType: "duration",
            expression: "5m",
            enabled: true,
            description: "",
          }}
          preserve={false}
          onFinish={(values) => void handleSubmit(values)}
        >
          <Form.Item name="name" label="任务名称" rules={[{ required: true, message: "请输入任务名称" }]}>
            <Input placeholder="例如：统计数据同步" />
          </Form.Item>
          <Form.Item name="taskKey" label="内置任务" rules={[{ required: true, message: "请选择内置任务" }]}>
            <Select
              placeholder="选择任务"
              options={tasks.map((task) => ({ value: task.key, label: `${task.key} — ${task.description}` }))}
            />
          </Form.Item>
          <Form.Item name="scheduleType" label="调度方式" rules={[{ required: true }]}>
            <Select options={scheduleTypeOptions} />
          </Form.Item>
          <Form.Item
            name="expression"
            label={scheduleType === "cron" ? "Cron 表达式" : "执行间隔"}
            extra={scheduleType === "cron" ? "使用标准五段 cron，例如：*/5 * * * *" : "使用 Go duration，例如：5m、1h、24h"}
            rules={[{ required: true, message: "请输入调度规则" }]}
          >
            <Input placeholder={scheduleType === "cron" ? "*/5 * * * *" : "5m"} />
          </Form.Item>
          <Form.Item name="description" label="说明">
            <Input.TextArea rows={3} placeholder="说明该任务的用途（可选）" />
          </Form.Item>
          <Form.Item name="enabled" label="启用状态" valuePropName="checked">
            <Switch checkedChildren="启用" unCheckedChildren="停用" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
