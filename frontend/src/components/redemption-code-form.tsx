import { useState, useEffect, type FormEvent } from "react";
import { Input, Select, Typography } from "antd";
import { api, responseData } from "@/lib/api";

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
      const result = await api.redemption_codes_plans_route_get();
      const data = responseData<{ id: string; title: string }[]>(result);
      if (data) setPlans(data);
      setPlansLoading(false);
    }
    loadPlans();
  }, []);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const body: Record<string, string | number> = { type, count };
    if (type === "subscription") body.planId = planId;
    if (type === "quota") body.credits = credits;
    if (expiresAt) body.expiresAt = new Date(expiresAt).toISOString();

    try {
      const result = await api.redemption_codes_route_post(body);

      if (result.success) {
        onSuccess();
      } else {
        setError(result.error || "操作失败");
      }
    } catch {
      setError("操作失败，请重试");
    }
  }

  return (
    <form id={formId} onSubmit={handleSubmit} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Typography.Text>兑换码类型 *</Typography.Text>
          <Select value={type} onChange={setType} options={[{ value: "subscription", label: "订阅" }, { value: "quota", label: "额度" }]} />
        </div>

        <div className="space-y-2">
          <Typography.Text>生成数量 *</Typography.Text>
          <Input
            id="count"
            type="number"
            min={1}
            max={1000}
            value={count}
            onChange={(e) => setCount(parseInt(e.target.value, 10) || 1)}
            placeholder="1-1000"
          />
        </div>

        {type === "subscription" && (
          <div className="space-y-2 md:col-span-2">
            <Typography.Text>订阅计划 *</Typography.Text>
            <Select
              value={planId}
              onChange={setPlanId}
              disabled={plansLoading}
              placeholder={plansLoading ? "加载中..." : "选择订阅计划"}
              options={plans.map(plan => ({ value: plan.id, label: plan.title }))}
            />
            {plans.length === 0 && !plansLoading && (
              <p className="text-sm text-amber-600">
                暂无可用的订阅计划，请先创建订阅计划
              </p>
            )}
          </div>
        )}

        {type === "quota" && (
          <div className="space-y-2 md:col-span-2">
            <Typography.Text>额度数量 *</Typography.Text>
            <Input
              id="credits"
              type="number"
              min={1}
              value={credits || ""}
              onChange={(e) => setCredits(parseInt(e.target.value, 10) || 0)}
              placeholder="请输入额度数量"
            />
          </div>
        )}

        <div className="space-y-2 md:col-span-2">
          <Typography.Text>过期时间</Typography.Text>
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
