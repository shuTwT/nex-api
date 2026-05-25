"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  createRedemptionCodes,
  getSubscriptionPlansForSelect,
} from "@/app/actions/redemption-codes";

interface RedemptionCodeFormProps {
  onSuccess: () => void;
  formId: string;
}

export function RedemptionCodeForm({ onSuccess, formId }: RedemptionCodeFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [type, setType] = useState<string>("subscription");
  const [count, setCount] = useState<number>(1);
  const [planId, setPlanId] = useState<string>("");
  const [credits, setCredits] = useState<number>(0);
  const [expiresAt, setExpiresAt] = useState<string>("");
  const [plans, setPlans] = useState<{ id: string; title: string }[]>([]);
  const [plansLoading, setPlansLoading] = useState(true);

  useEffect(() => {
    async function loadPlans() {
      setPlansLoading(true);
      const result = await getSubscriptionPlansForSelect();
      if (result.success && result.data) {
        setPlans(result.data);
      }
      setPlansLoading(false);
    }
    loadPlans();
  }, []);

  async function handleSubmit(formData: FormData) {
    setError(null);

    formData.set("type", type);
    formData.set("count", count.toString());
    if (type === "subscription") {
      formData.set("planId", planId);
    }
    if (type === "quota") {
      formData.set("credits", credits.toString());
    }
    if (expiresAt) {
      formData.set("expiresAt", new Date(expiresAt).toISOString());
    }

    try {
      const result = await createRedemptionCodes(formData);

      if (result.success) {
        onSuccess();
      } else {
        setError(result.error || "操作失败");
      }
    } catch (err) {
      setError("操作失败，请重试");
    }
  }

  return (
    <form id={formId} action={handleSubmit} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Label>兑换码类型 *</Label>
          <Select value={type} onValueChange={setType}>
            <SelectTrigger>
              <SelectValue placeholder="选择类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="subscription">订阅</SelectItem>
              <SelectItem value="quota">额度</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="count">生成数量 *</Label>
          <Input
            id="count"
            type="number"
            min={1}
            max={1000}
            value={count}
            onChange={(e) => setCount(parseInt(e.target.value) || 1)}
            placeholder="1-1000"
          />
        </div>

        {type === "subscription" && (
          <div className="space-y-2 md:col-span-2">
            <Label>订阅计划 *</Label>
            <Select
              value={planId}
              onValueChange={setPlanId}
              disabled={plansLoading}
            >
              <SelectTrigger>
                <SelectValue
                  placeholder={plansLoading ? "加载中..." : "选择订阅计划"}
                />
              </SelectTrigger>
              <SelectContent>
                {plans.map((plan) => (
                  <SelectItem key={plan.id} value={plan.id}>
                    {plan.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {plans.length === 0 && !plansLoading && (
              <p className="text-sm text-amber-600">
                暂无可用的订阅计划，请先创建订阅计划
              </p>
            )}
          </div>
        )}

        {type === "quota" && (
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="credits">额度数量 *</Label>
            <Input
              id="credits"
              type="number"
              min={1}
              value={credits || ""}
              onChange={(e) => setCredits(parseInt(e.target.value) || 0)}
              placeholder="请输入额度数量"
            />
          </div>
        )}

        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="expiresAt">过期时间</Label>
          <Input
            id="expiresAt"
            type="datetime-local"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            min={new Date().toISOString().slice(0, 16)}
          />
          <p className="text-xs text-slate-400">留空则为永久有效</p>
        </div>
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}
    </form>
  );
}
