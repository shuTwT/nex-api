"use client";

import { useEffect, useState, useCallback, useTransition } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
  Image as ImageIcon,
  Megaphone,
  ExternalLink,
  Power,
  Eye
} from "lucide-react";
import { api } from "@/lib/api-client";
import { AdvertisementForm } from "@/components/advertisement-form";
import { DeleteAdvertisementDialog } from "@/components/delete-advertisement-dialog";
import { Pagination } from "@/components/pagination";
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

interface AdvertisementStats {
  totalAds: number;
  activeAds: number;
  inactiveAds: number;
  positionStats: Array<{
    position: string;
    _count: number;
  }>;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export default function AdvertisementsPage() {
  const [advertisements, setAdvertisements] = useState<Advertisement[]>([]);
  const [stats, setStats] = useState<AdvertisementStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [positionFilter, setPositionFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedPosition, setAppliedPosition] = useState<string>("all");
  const [appliedStatus, setAppliedStatus] = useState<string>("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingAd, setEditingAd] = useState<Advertisement | null>(null);
  const [deletingAd, setDeletingAd] = useState<Advertisement | null>(null);
  const [isPending, startTransition] = useTransition();

  const loadAdvertisements = useCallback(async () => {
    setIsLoading(true);
    const params: Record<string, string | number | boolean | undefined> = {
      page: currentPage,
      limit: pageSize,
    };

    if (appliedSearch) {
      params.search = appliedSearch;
    }

    if (appliedPosition !== "all") {
      params.position = appliedPosition;
    }

    if (appliedStatus !== "all") {
      params.isActive = appliedStatus === "active";
    }

    const result = await api.paginated("/api/advertisements", params);

    if (result.success && result.data) {
      setAdvertisements(result.data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedPosition, appliedStatus]);

  useEffect(() => {
    loadAdvertisements();
  }, [loadAdvertisements]);

  async function loadStats() {
    const result = await api.get("/api/advertisements/stats");
    if (result.success && result.data) {
      setStats(result.data);
    }
  }

  useEffect(() => {
    loadStats();
  }, []);

  function handleAddAdvertisement() {
    setEditingAd(null);
    setShowForm(true);
  }

  function handleEditAdvertisement(ad: Advertisement) {
    setEditingAd(ad);
    setShowForm(true);
  }

  function handleDeleteAdvertisement(ad: Advertisement) {
    setDeletingAd(ad);
  }

  function handleFormSuccess() {
    setShowForm(false);
    setEditingAd(null);
    startTransition(() => {
      loadAdvertisements();
      loadStats();
    });
  }

  function handleDeleteSuccess() {
    startTransition(() => {
      loadAdvertisements();
      loadStats();
    });
  }

  async function handleToggleStatus(ad: Advertisement) {
    const result = await api.put(`/api/advertisements/${ad.id}/toggle`);
    if (result.success) {
      startTransition(() => {
        loadAdvertisements();
        loadStats();
      });
    }
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedPosition(positionFilter);
    setAppliedStatus(statusFilter);
    setCurrentPage(1);
  }

  function handleReset() {
    setSearchInput("");
    setPositionFilter("all");
    setStatusFilter("all");
    setAppliedSearch("");
    setAppliedPosition("all");
    setAppliedStatus("all");
    setCurrentPage(1);
  }

  function getPositionLabel(position: string): string {
    return AdPositionLabels[position as AdPosition] || position;
  }

  const statsCards = [
    {
      title: "总广告数",
      value: stats?.totalAds || 0,
      icon: Megaphone,
      color: "blue",
    },
    {
      title: "已启用",
      value: stats?.activeAds || 0,
      icon: Power,
      color: "green",
    },
    {
      title: "已禁用",
      value: stats?.inactiveAds || 0,
      icon: Power,
      color: "gray",
    },
    {
      title: "广告位类型",
      value: stats?.positionStats.length || 0,
      icon: Eye,
      color: "purple",
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">广告位管理</h1>
          <p className="text-slate-500 mt-1">管理网站广告位和广告内容</p>
        </div>
        <Button 
          className="gap-2 cursor-pointer"
          onClick={handleAddAdvertisement}
        >
          <Plus className="h-4 w-4" />
          添加广告
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            blue: "bg-blue-50 text-blue-600",
            green: "bg-green-50 text-green-600",
            gray: "bg-gray-50 text-gray-600",
            purple: "bg-purple-50 text-purple-600",
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
          <div className="flex flex-wrap items-end gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索广告标题..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={positionFilter} onValueChange={setPositionFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="广告位" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部广告位</SelectItem>
                  {AdPositionOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[120px]">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="active">已启用</SelectItem>
                  <SelectItem value="inactive">已禁用</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button
              size="sm"
              onClick={handleQuery}
              className="cursor-pointer"
            >
              查询
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleReset}
              className="cursor-pointer"
            >
              重置
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">广告列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
            </div>
          ) : advertisements.length === 0 ? (
            <div className="text-center py-12">
              <ImageIcon className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">没有找到广告</h3>
              <p className="text-slate-500 mb-4">尝试调整搜索条件或添加新广告</p>
              <Button onClick={handleAddAdvertisement} className="gap-2 cursor-pointer">
                <Plus className="h-4 w-4" />
                添加广告
              </Button>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">广告信息</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">广告位</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">图片尺寸</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">状态</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">创建时间</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {advertisements.map((ad) => (
                      <tr key={ad.id} className="border-b border-slate-100 hover:bg-slate-50 transition-colors">
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-3">
                            <div className="h-16 w-24 rounded bg-slate-100 overflow-hidden flex items-center justify-center">
                              {ad.image ? (
                                <img 
                                  src={ad.image} 
                                  alt={ad.title}
                                  className="w-full h-full object-cover"
                                />
                              ) : (
                                <ImageIcon className="h-6 w-6 text-slate-400" />
                              )}
                            </div>
                            <div>
                              <p className="text-sm font-medium text-slate-900">{ad.title}</p>
                              <a 
                                href={ad.link} 
                                target="_blank" 
                                rel="noopener noreferrer"
                                className="text-xs text-blue-600 hover:underline flex items-center gap-1 mt-1"
                              >
                                <ExternalLink className="h-3 w-3" />
                                查看链接
                              </a>
                            </div>
                          </div>
                        </td>
                        <td className="py-3 px-4">
                          <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                            {getPositionLabel(ad.position)}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {ad.imageWidth} × {ad.imageHeight}
                        </td>
                        <td className="py-3 px-4">
                          <Badge 
                            variant="outline" 
                            className={
                              ad.isActive
                                ? "bg-green-50 text-green-700 border-green-200" 
                                : "bg-gray-50 text-gray-700 border-gray-200"
                            }
                          >
                            {ad.isActive ? "已启用" : "已禁用"}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {new Date(ad.createdAt).toLocaleDateString("zh-CN")}
                        </td>
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-8 w-8 p-0 cursor-pointer"
                              onClick={() => handleEditAdvertisement(ad)}
                              title="编辑"
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-8 w-8 p-0 cursor-pointer"
                              onClick={() => handleToggleStatus(ad)}
                              title={ad.isActive ? "禁用" : "启用"}
                            >
                              <Power className={`h-4 w-4 ${ad.isActive ? "text-green-600" : "text-gray-400"}`} />
                            </Button>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="h-8 w-8 p-0 cursor-pointer text-red-600 hover:text-red-700 hover:bg-red-50"
                              onClick={() => handleDeleteAdvertisement(ad)}
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

              <div className="mt-4">
                <Pagination
                  currentPage={pagination?.page ?? 1}
                  totalPages={pagination?.totalPages ?? 1}
                  total={pagination?.total ?? 0}
                  pageSize={pagination?.limit ?? pageSize}
                  onPageChange={handlePageChange}
                  onPageSizeChange={handlePageSizeChange}
                />
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent className="max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{editingAd ? "编辑广告" : "添加广告"}</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto">
            <AdvertisementForm
              advertisement={editingAd || undefined}
              onSuccess={handleFormSuccess}
              formId="advertisement-form"
            />
          </div>
          <DialogFooter className="border-t pt-4 mt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowForm(false)}
              className="cursor-pointer"
            >
              取消
            </Button>
            <Button type="submit" form="advertisement-form" className="cursor-pointer">
              {editingAd ? "保存" : "创建"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {deletingAd && (
        <DeleteAdvertisementDialog
          open={!!deletingAd}
          onOpenChange={(open) => !open && setDeletingAd(null)}
          advertisementId={deletingAd.id}
          advertisementTitle={deletingAd.title}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </div>
  );
}
