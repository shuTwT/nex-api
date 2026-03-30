"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { createSubscriptionPlan, updateSubscriptionPlan } from "@/app/actions/subscription-plans";

interface SubscriptionPlan {
  id?: string;
  title: string;
  price: number;
  totalCredits: number;
  sortOrder: number;
  validityDuration: number;
  validityUnit: string;
  creditResetCycle: string;
  isActive: boolean;
}

interface SubscriptionPlanFormProps {
  plan?: SubscriptionPlan;
  onSuccess: () => void;
  formId: string;
}

export function SubscriptionPlanForm({ plan, onSuccess, formId }: SubscriptionPlanFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validityUnit, setValidityUnit] = useState(plan?.validityUnit || "day");
  const [creditResetCycle, setCreditResetCycle] = useState(plan?.creditResetCycle || "month");

  const isEdit = !!plan;

  async function handleSubmit(formData: FormData) {
    setIsLoading(true);
    setError(null);

    formData.set("validityUnit", validityUnit);
    formData.set("creditResetCycle", creditResetCycle);

    try {
      const result = isEdit
        ? await updateSubscriptionPlan(formData)
        : await createSubscriptionPlan(formData);

      if (result.success) {
        onSuccess();
      } else {
        setError(result.error || "操作失败");
      }
    } catch (err) {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form id={formId} action={handleSubmit} className="space-y-4">
      {plan?.id && (
        <input type="hidden" name="id" value={plan.id} />
      )}

      <Card>
        <CardContent className="p-6 space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="title">套餐标题 *</Label>
              <Input
                id="title"
                name="title"
                defaultValue={plan?.title || ""}
                required
                placeholder="例如：基础版、专业版"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="price">价格 *</Label>
              <Input
                id="price"
                name="price"
                type="number"
                step="0.01"
                min="0"
                defaultValue={plan?.price || 0}
                required
                placeholder="例如：99.99"
              />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="totalCredits">总额度 *</Label>
              <Input
                id="totalCredits"
                name="totalCredits"
                type="number"
                min="0"
                defaultValue={plan?.totalCredits || 0}
                required
                placeholder="例如：1000"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="sortOrder">排序</Label>
              <Input
                id="sortOrder"
                name="sortOrder"
                type="number"
                min="0"
                defaultValue={plan?.sortOrder || 0}
                placeholder="数字越小越靠前"
              />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="validityDuration">有效期时长 *</Label>
              <Input
                id="validityDuration"
                name="validityDuration"
                type="number"
                min="1"
                defaultValue={plan?.validityDuration || 30}
                required
                placeholder="例如：30"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="validityUnit">有效期单位</Label>
              <Select value={validityUnit} onValueChange={setValidityUnit} disabled={isLoading}>
                <SelectTrigger id="validityUnit">
                  <SelectValue placeholder="选择单位" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="day">天</SelectItem>
                  <SelectItem value="week">周</SelectItem>
                  <SelectItem value="month">月</SelectItem>
                  <SelectItem value="year">年</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="creditResetCycle">额度重置周期</Label>
            <Select value={creditResetCycle} onValueChange={setCreditResetCycle} disabled={isLoading}>
              <SelectTrigger id="creditResetCycle">
                <SelectValue placeholder="选择周期" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="day">每天</SelectItem>
                <SelectItem value="week">每周</SelectItem>
                <SelectItem value="month">每月</SelectItem>
                <SelectItem value="year">每年</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="isActive"
              name="isActive"
              defaultChecked={plan?.isActive ?? true}
              className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
            />
            <Label htmlFor="isActive">启用</Label>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}
        </CardContent>
      </Card>
    </form>
  );
}
