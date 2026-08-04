import { useCallback, useEffect, useState, useTransition } from "react";
import { Button, Card, Flex, Image, Input, Modal, Select, Space, Statistic, Table, Tag, Typography } from "antd";
import { Edit, ExternalLink, Eye, Image as ImageIcon, Megaphone, Plus, Power, Search, Trash2 } from "lucide-react";
import { api, responseData } from "@/lib/api";
import { AdvertisementForm } from "@/components/advertisement-form";
import { DeleteAdvertisementDialog } from "@/components/delete-advertisement-dialog";
import { AdPositionLabels, AdPosition, AdPositionOptions } from "@/types/ad-position";

interface Advertisement {
  id: string;
  image: string;
  imageWidth: number;
  imageHeight: number;
  link: string;
  title: string;
  position: string;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}
interface AdvertisementStats { totalAds: number; activeAds: number; inactiveAds: number; positionStats: Array<{ position: string; _count: number }>; }
interface PaginationInfo { page: number; limit: number; total: number; totalPages: number; }

export default function AdvertisementsPage() {
  const [advertisements, setAdvertisements] = useState<Advertisement[]>([]);
  const [stats, setStats] = useState<AdvertisementStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [positionFilter, setPositionFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedPosition, setAppliedPosition] = useState("all");
  const [appliedStatus, setAppliedStatus] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingAd, setEditingAd] = useState<Advertisement | null>(null);
  const [deletingAd, setDeletingAd] = useState<Advertisement | null>(null);
  const [, startTransition] = useTransition();

  const loadAdvertisements = useCallback(async () => {
    setIsLoading(true);
    const query: Record<string, string | number | boolean> = { page: currentPage, limit: pageSize };
    if (appliedSearch) query.search = appliedSearch;
    if (appliedPosition !== "all") query.position = appliedPosition;
    if (appliedStatus !== "all") query.isActive = appliedStatus === "active";
    const result = await api.advertisements_route_get(query);
    if (result.success) {
      const data = responseData<Advertisement[]>(result);
      if (data) setAdvertisements(data);
      if (result.pagination) setPagination(result.pagination);
    }
    setIsLoading(false);
  }, [appliedPosition, appliedSearch, appliedStatus, currentPage, pageSize]);
  useEffect(() => { void loadAdvertisements(); }, [loadAdvertisements]);

  const loadStats = useCallback(async () => {
    const data = responseData<AdvertisementStats>(await api.advertisements_stats_route_get());
    if (data) setStats(data);
  }, []);
  useEffect(() => { void loadStats(); }, [loadStats]);

  function refresh() { startTransition(() => { void loadAdvertisements(); void loadStats(); }); }
  function handleAddAdvertisement() { setEditingAd(null); setShowForm(true); }
  function handleQuery() { setAppliedSearch(searchInput); setAppliedPosition(positionFilter); setAppliedStatus(statusFilter); setCurrentPage(1); }
  function handleReset() { setSearchInput(""); setPositionFilter("all"); setStatusFilter("all"); setAppliedSearch(""); setAppliedPosition("all"); setAppliedStatus("all"); setCurrentPage(1); }
  function getPositionLabel(position: string) { return AdPositionLabels[position as AdPosition] || position; }
  async function handleToggleStatus(ad: Advertisement) { const result = await api.advertisements_id_toggle_route_put({ id: ad.id }); if (result.success) refresh(); }

  const statsCards = [
    { title: "总广告数", value: stats?.totalAds || 0, icon: <Megaphone size={20} />, color: "#1677ff" },
    { title: "已启用", value: stats?.activeAds || 0, icon: <Power size={20} />, color: "#52c41a" },
    { title: "已禁用", value: stats?.inactiveAds || 0, icon: <Power size={20} />, color: "#8c8c8c" },
    { title: "广告位类型", value: stats?.positionStats.length || 0, icon: <Eye size={20} />, color: "#722ed1" },
  ];

  return (
    <Flex vertical gap={24}>
      <Flex justify="space-between" align="center">
        <div><Typography.Title level={2} style={{ margin: 0 }}>广告位管理</Typography.Title><Typography.Text type="secondary">管理网站广告位和广告内容</Typography.Text></div>
        <Button type="primary" icon={<Plus size={16} />} onClick={handleAddAdvertisement}>添加广告</Button>
      </Flex>
      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => <Card key={stat.title}><Statistic title={stat.title} value={stat.value} prefix={<span style={{ color: stat.color }}>{stat.icon}</span>} /></Card>)}
      </div>
      <Card>
        <Space wrap>
          <Input prefix={<Search size={16} />} placeholder="搜索广告标题..." value={searchInput} onChange={(event) => setSearchInput(event.target.value)} style={{ width: 260 }} onPressEnter={handleQuery} />
          <Select value={positionFilter} onChange={setPositionFilter} style={{ width: 160 }} options={[{ value: "all", label: "全部广告位" }, ...AdPositionOptions.map((option) => ({ value: option.value, label: option.label }))]} />
          <Select value={statusFilter} onChange={setStatusFilter} style={{ width: 120 }} options={[{ value: "all", label: "全部" }, { value: "active", label: "已启用" }, { value: "inactive", label: "已禁用" }]} />
          <Button type="primary" size="medium" onClick={handleQuery}>查询</Button><Button size="medium" onClick={handleReset}>重置</Button>
        </Space>
      </Card>
      <Card title="广告列表">
        <Table<Advertisement>
          rowKey="id" loading={isLoading} dataSource={advertisements} scroll={{ x: 1000 }}
          locale={{ emptyText: <Flex vertical align="center" gap={12}><ImageIcon size={40} color="#bfbfbf" /><Typography.Text type="secondary">没有找到广告</Typography.Text><Button type="primary" icon={<Plus size={16} />} onClick={handleAddAdvertisement}>添加广告</Button></Flex> }}
          pagination={{ current: pagination?.page ?? currentPage, pageSize: pagination?.limit ?? pageSize, total: pagination?.total ?? 0, showSizeChanger: true, showTotal: (total) => `共 ${total} 条`, onChange: (page, size) => { if (size !== pageSize) { setPageSize(size); setCurrentPage(1); } else setCurrentPage(page); } }}
          columns={[
            { title: "广告信息", key: "ad", render: (_, ad) => <Space><Image width={96} height={64} preview={false} src={ad.image} fallback="" style={{ objectFit: "cover", borderRadius: 4 }} /><Flex vertical gap={2}><Typography.Text strong>{ad.title}</Typography.Text><Typography.Link href={ad.link} target="_blank"><ExternalLink size={12} style={{ display: "inline", marginRight: 4 }} />查看链接</Typography.Link></Flex></Space> },
            { title: "广告位", dataIndex: "position", render: (position) => <Tag color="blue">{getPositionLabel(position)}</Tag> },
            { title: "图片尺寸", key: "size", render: (_, ad) => `${ad.imageWidth} × ${ad.imageHeight}` },
            { title: "状态", dataIndex: "isActive", render: (active) => <Tag color={active ? "success" : "default"}>{active ? "已启用" : "已禁用"}</Tag> },
            { title: "创建时间", dataIndex: "createdAt", render: (createdAt) => new Date(createdAt).toLocaleDateString("zh-CN") },
            { title: "操作", key: "actions", render: (_, ad) => <Space size="small"><Button type="text" shape="circle" title="编辑" icon={<Edit size={16} />} onClick={() => { setEditingAd(ad); setShowForm(true); }} /><Button type="text" shape="circle" title={ad.isActive ? "禁用" : "启用"} icon={<Power size={16} color={ad.isActive ? "#52c41a" : "#8c8c8c"} />} onClick={() => void handleToggleStatus(ad)} /><Button type="text" danger shape="circle" title="删除" icon={<Trash2 size={16} />} onClick={() => setDeletingAd(ad)} /></Space> },
          ]}
        />
      </Card>
      <Modal open={showForm} title={editingAd ? "编辑广告" : "添加广告"} onCancel={() => setShowForm(false)} destroyOnHidden width={640} okText={editingAd ? "保存" : "创建"} okButtonProps={{ htmlType: "submit", form: "advertisement-form" }}>
        <AdvertisementForm advertisement={editingAd || undefined} onSuccess={() => { setShowForm(false); setEditingAd(null); refresh(); }} formId="advertisement-form" />
      </Modal>
      {deletingAd && <DeleteAdvertisementDialog open onOpenChange={(open) => !open && setDeletingAd(null)} advertisementId={deletingAd.id} advertisementTitle={deletingAd.title} onSuccess={refresh} />}
    </Flex>
  );
}
