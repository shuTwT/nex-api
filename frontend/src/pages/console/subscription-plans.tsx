import { useEffect, useState, useCallback } from "react";
import { Badge, Button, Card, Input, Modal, Select, Typography } from "antd";
import {
  Plus,
  Edit,
  Trash2,
  CreditCard,
  Calendar,
  DollarSign,
  ArrowUpDown,
  Search,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { SubscriptionPlanForm } from "@/components/subscription-plan-form";
import { DeleteSubscriptionPlanDialog } from "@/components/delete-subscription-plan-dialog";
import { Pagination } from "@/components/pagination";
import { toast } from "sonner";

interface SubscriptionPlan {
  id: string;
  title: string;
  price: number;
  totalCredits: number;
  sortOrder: number;
  validityDuration: number;
  validityUnit: string;
  creditResetCycle: string;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export default function SubscriptionPlansPage() {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedStatus, setAppliedStatus] = useState<string>("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [showForm, setShowForm] = useState(false);
  const [editingPlan, setEditingPlan] = useState<SubscriptionPlan | null>(null);
  const [deletingPlan, setDeletingPlan] = useState<SubscriptionPlan | null>(
    null,
  );

  const loadPlans = useCallback(async () => {
    setIsLoading(true);
    const query: Record<string, string | number | boolean> = {
      page: currentPage,
      limit: pageSize,
      search: appliedSearch,
    };
    if (appliedStatus !== "all") {
      query.isActive = appliedStatus;
    }
    const result = await api.subscription_plans_route_get(query);
    if (result.success) {
      const data = responseData<SubscriptionPlan[]>(result);
      if (data) setPlans(data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedStatus]);

  useEffect(() => {
    loadPlans();
  }, [loadPlans]);

  function handleAddPlan() {
    setEditingPlan(null);
    setShowForm(true);
  }

  function handleEditPlan(plan: SubscriptionPlan) {
    setEditingPlan(plan);
    setShowForm(true);
  }

  function handleDeletePlan(plan: SubscriptionPlan) {
    setDeletingPlan(plan);
  }

  function handleFormSuccess() {
    toast.success(editingPlan ? "订阅计划更新成功" : "订阅计划创建成功");
    setShowForm(false);
    setEditingPlan(null);
    loadPlans();
  }

  function handleDeleteSuccess() {
    loadPlans();
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedStatus(statusFilter);
    setCurrentPage(1);
  }

  function handleReset() {
    setSearchInput("");
    setStatusFilter("all");
    setAppliedSearch("");
    setAppliedStatus("all");
    setCurrentPage(1);
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  const validityUnitLabels: Record<string, string> = {
    day: "天",
    week: "周",
    month: "月",
    year: "年",
  };

  const creditResetCycleLabels: Record<string, string> = {
    day: "每天",
    week: "每周",
    month: "每月",
    year: "每年",
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">订阅计划管理</h1>
          <p className="text-slate-500 mt-1">管理订阅套餐和价格</p>
        </div>
        <Button className="gap-2 cursor-pointer" onClick={handleAddPlan}>
          <Plus className="h-4 w-4" />
          添加套餐
        </Button>
      </div>

      <div>
        <Card>
          <div className="p-4">
            <div className="flex flex-wrap items-end gap-3">
              <div className="relative flex-1 min-w-[200px]">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
                <Input
                  placeholder="搜索计划名称..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  className="pl-10"
                />
              </div>
              <Select
                value={statusFilter}
                onChange={setStatusFilter}
                className="w-[120px]"
                options={[
                  { value: "all", label: "全部" },
                  { value: "true", label: "启用" },
                  { value: "false", label: "禁用" },
                ]}
              />
              <Button
                size="medium"
                onClick={handleQuery}
                className="cursor-pointer"
              >
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
          </div>
        </Card>
      </div>
      {plans.length === 0 ? (
        <Card>
          <div className="py-12 text-center">
            <CreditCard className="h-12 w-12 text-slate-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-slate-900 mb-2">
              暂无订阅计划
            </h3>
            <p className="text-slate-500 mb-4">创建第一个订阅计划开始使用</p>
            <Button onClick={handleAddPlan} className="gap-2 cursor-pointer">
              <Plus className="h-4 w-4" />
              添加套餐
            </Button>
          </div>
        </Card>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {plans.map((plan) => (
              <Card key={plan.id} className="hover:shadow-md transition-shadow">
                <div className="p-4">
                  <div className="flex items-start justify-between">
                    <Typography.Title level={5}>{plan.title}</Typography.Title>
                    <Badge
                      className={
                        plan.isActive
                          ? "bg-green-50 text-green-700 border-green-200"
                          : "bg-gray-50 text-gray-700 border-gray-200"
                      }
                    >
                      {plan.isActive ? "启用" : "禁用"}
                    </Badge>
                  </div>
                </div>
                <div className="flex flex-col gap-4 px-4 pb-4">
                  <div className="flex items-baseline gap-1">
                    <DollarSign className="h-5 w-5 text-slate-400" />
                    <span className="text-3xl font-bold text-slate-900">
                      {plan.price}
                    </span>
                  </div>

                  <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-2 text-sm">
                      <CreditCard className="h-4 w-4 text-slate-400" />
                      <span className="text-slate-600">总额度: </span>
                      <span className="font-medium text-slate-900">
                        {plan.totalCredits.toLocaleString()}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <Calendar className="h-4 w-4 text-slate-400" />
                      <span className="text-slate-600">有效期: </span>
                      <span className="font-medium text-slate-900">
                        {plan.validityDuration}{" "}
                        {validityUnitLabels[plan.validityUnit]}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <ArrowUpDown className="h-4 w-4 text-slate-400" />
                      <span className="text-slate-600">重置周期: </span>
                      <span className="font-medium text-slate-900">
                        {creditResetCycleLabels[plan.creditResetCycle]}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <span className="h-4 w-4 text-slate-400" />
                      <span className="text-slate-600">排序: </span>
                      <span className="font-medium text-slate-900">
                        {plan.sortOrder}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 pt-4 border-t border-slate-200">
                    <Button
                      type="text"
                      size="small"
                      className="h-8 w-8 p-0 cursor-pointer"
                      onClick={() => handleEditPlan(plan)}
                      title="编辑"
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button
                      type="text"
                      size="small"
                      className="h-8 w-8 p-0 cursor-pointer text-red-600 hover:text-red-700 hover:bg-red-50"
                      onClick={() => handleDeletePlan(plan)}
                      title="删除"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          <Pagination
            currentPage={pagination?.page ?? 1}
            totalPages={pagination?.totalPages ?? 1}
            total={pagination?.total ?? 0}
            pageSize={pagination?.limit ?? pageSize}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
          />
        </>
      )}

      <Modal
        open={showForm}
        title={editingPlan ? "编辑订阅计划" : "添加订阅计划"}
        onCancel={() => setShowForm(false)}
        destroyOnHidden
        footer={[
          <Button key="cancel" onClick={() => setShowForm(false)}>
            取消
          </Button>,
          <Button
            key="submit"
            type="primary"
            htmlType="submit"
            form="subscription-plan-form"
          >
            {editingPlan ? "保存" : "创建"}
          </Button>,
        ]}
      >
        <div className="flex-1 overflow-y-auto">
          <SubscriptionPlanForm
            plan={editingPlan || undefined}
            onSuccess={handleFormSuccess}
            formId="subscription-plan-form"
          />
        </div>
      </Modal>

      {deletingPlan && (
        <DeleteSubscriptionPlanDialog
          open={!!deletingPlan}
          onOpenChange={(open) => !open && setDeletingPlan(null)}
          planId={deletingPlan.id}
          planTitle={deletingPlan.title}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </div>
  );
}
