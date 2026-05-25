"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Plus, Edit, Trash2, CreditCard, Calendar, DollarSign, ArrowUpDown } from "lucide-react";
import { api } from "@/lib/api-client";
import { SubscriptionPlanForm } from "@/components/subscription-plan-form";
import { DeleteSubscriptionPlanDialog } from "@/components/delete-subscription-plan-dialog";
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

export default function SubscriptionPlansPage() {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingPlan, setEditingPlan] = useState<SubscriptionPlan | null>(null);
  const [deletingPlan, setDeletingPlan] = useState<SubscriptionPlan | null>(null);

  async function loadPlans() {
    setIsLoading(true);
    const result = await api.get("/api/subscription-plans");
    if (result.success && result.data) {
      setPlans(result.data);
    }
    setIsLoading(false);
  }

  useEffect(() => {
    loadPlans();
  }, []);

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
    <div className="space-y-6">
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

      {plans.length === 0 ? (
        <Card>
          <CardContent className="text-center py-12">
            <CreditCard className="h-12 w-12 text-slate-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-slate-900 mb-2">暂无订阅计划</h3>
            <p className="text-slate-500 mb-4">创建第一个订阅计划开始使用</p>
            <Button onClick={handleAddPlan} className="gap-2 cursor-pointer">
              <Plus className="h-4 w-4" />
              添加套餐
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {plans.map((plan) => (
            <Card key={plan.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <CardTitle className="text-lg">{plan.title}</CardTitle>
                  <Badge variant="outline" className={plan.isActive ? "bg-green-50 text-green-700 border-green-200" : "bg-gray-50 text-gray-700 border-gray-200"}>
                    {plan.isActive ? "启用" : "禁用"}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-baseline gap-1">
                  <DollarSign className="h-5 w-5 text-slate-400" />
                  <span className="text-3xl font-bold text-slate-900">{plan.price}</span>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <CreditCard className="h-4 w-4 text-slate-400" />
                    <span className="text-slate-600">总额度: </span>
                    <span className="font-medium text-slate-900">{plan.totalCredits.toLocaleString()}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-slate-400" />
                    <span className="text-slate-600">有效期: </span>
                    <span className="font-medium text-slate-900">{plan.validityDuration} {validityUnitLabels[plan.validityUnit]}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                    <span className="text-slate-600">重置周期: </span>
                    <span className="font-medium text-slate-900">{creditResetCycleLabels[plan.creditResetCycle]}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <span className="h-4 w-4 text-slate-400" />
                    <span className="text-slate-600">排序: </span>
                    <span className="font-medium text-slate-900">{plan.sortOrder}</span>
                  </div>
                </div>

                <div className="flex items-center gap-2 pt-4 border-t border-slate-200">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0 cursor-pointer"
                    onClick={() => handleEditPlan(plan)}
                    title="编辑"
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0 cursor-pointer text-red-600 hover:text-red-700 hover:bg-red-50"
                    onClick={() => handleDeletePlan(plan)}
                    title="删除"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent className="max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{editingPlan ? "编辑订阅计划" : "添加订阅计划"}</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto">
            <SubscriptionPlanForm
              plan={editingPlan || undefined}
              onSuccess={handleFormSuccess}
              formId="subscription-plan-form"
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
            <Button type="submit" form="subscription-plan-form" className="cursor-pointer">
              {editingPlan ? "保存" : "创建"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
