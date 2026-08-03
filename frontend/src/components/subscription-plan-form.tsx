import { useState, type FormEvent } from "react";
import { Card, Checkbox, Input, Select, Typography } from "antd";
import { api } from "@/lib/api";

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

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const formData = new FormData(e.currentTarget);
    formData.set("validityUnit", validityUnit);
    formData.set("creditResetCycle", creditResetCycle);
    const body = {
      title: String(formData.get("title") ?? ""),
      price: Number(formData.get("price") ?? 0),
      totalCredits: Number(formData.get("totalCredits") ?? 0),
      sortOrder: Number(formData.get("sortOrder") ?? 0),
      validityDuration: Number(formData.get("validityDuration") ?? 0),
      validityUnit,
      creditResetCycle,
      isActive: formData.has("isActive"),
    };

    try {
      const result = isEdit
        ? await api.subscription_plans_id_route_put({ id: plan?.id ?? "" }, body)
        : await api.subscription_plans_route_post(body);

      if (result.success) {
        onSuccess();
      } else {
        setError(result.error || "操作失败");
      }
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form id={formId} onSubmit={handleSubmit} className="space-y-4">
      <Card styles={{ body: { padding: 24 } }}>
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Typography.Text>套餐标题 *</Typography.Text>
              <Input
                id="title"
                name="title"
                defaultValue={plan?.title || ""}
                required
                placeholder="例如：基础版、专业版"
              />
            </div>

            <div className="space-y-2">
              <Typography.Text>价格 *</Typography.Text>
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
              <Typography.Text>总额度 *</Typography.Text>
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
              <Typography.Text>排序</Typography.Text>
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
              <Typography.Text>有效期时长 *</Typography.Text>
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
              <Typography.Text>有效期单位</Typography.Text>
              <Select value={validityUnit} onChange={setValidityUnit} disabled={isLoading} options={[{ value: "day", label: "天" }, { value: "week", label: "周" }, { value: "month", label: "月" }, { value: "year", label: "年" }]} />
            </div>
          </div>

          <div className="space-y-2">
            <Typography.Text>额度重置周期</Typography.Text>
            <Select value={creditResetCycle} onChange={setCreditResetCycle} disabled={isLoading} options={[{ value: "day", label: "每天" }, { value: "week", label: "每周" }, { value: "month", label: "每月" }, { value: "year", label: "每年" }]} />
          </div>

          <div className="flex items-center gap-2">
            <Checkbox
              id="isActive"
              name="isActive"
              defaultChecked={plan?.isActive ?? true}
            >启用</Checkbox>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}
        </div>
      </Card>
    </form>
  );
}
