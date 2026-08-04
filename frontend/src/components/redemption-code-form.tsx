import { useEffect, useState } from "react";
import { Alert, Form, Input, InputNumber, Select } from "antd";
import { api, responseData } from "@/lib/api";

interface RedemptionCodeFormProps {
  onSuccess: () => void;
  formId: string;
}
interface RedemptionCodeValues {
  type: "subscription" | "quota";
  count: number;
  planId?: string;
  credits?: number;
  expiresAt?: string;
}

export function RedemptionCodeForm({
  onSuccess,
  formId,
}: RedemptionCodeFormProps) {
  const [form] = Form.useForm<RedemptionCodeValues>();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [plans, setPlans] = useState<{ id: string; title: string }[]>([]);
  const [plansLoading, setPlansLoading] = useState(true);
  const type = Form.useWatch("type", form) ?? "subscription";

  useEffect(() => {
    async function loadPlans() {
      setPlansLoading(true);
      const result = await api.redemption_codes_plans_route_get();
      const data = responseData<{ id: string; title: string }[]>(result);
      if (data) setPlans(data);
      setPlansLoading(false);
    }
    void loadPlans();
  }, []);

  async function handleFinish(values: RedemptionCodeValues) {
    setError(null);
    setIsLoading(true);
    const body: Record<string, string | number> = {
      type: values.type,
      count: values.count,
    };
    if (values.type === "subscription") body.planId = values.planId ?? "";
    if (values.type === "quota") body.credits = values.credits ?? 0;
    if (values.expiresAt)
      body.expiresAt = new Date(values.expiresAt).toISOString();
    try {
      const result = await api.redemption_codes_route_post(body);
      if (result.success) onSuccess();
      else setError(result.error || "操作失败");
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Form<RedemptionCodeValues>
      id={formId}
      form={form}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      initialValues={{ type: "subscription", count: 1 }}
    >
      <div className="grid gap-4 md:grid-cols-2">
        <Form.Item
          name="type"
          label="兑换码类型"
          rules={[{ required: true, message: "请选择兑换码类型" }]}
        >
          <Select
            options={[
              { value: "subscription", label: "订阅" },
              { value: "quota", label: "额度" },
            ]}
          />
        </Form.Item>
        <Form.Item
          name="count"
          label="生成数量"
          rules={[{ required: true, message: "请输入生成数量" }]}
        >
          <InputNumber
            min={1}
            max={1000}
            className="w-full"
            placeholder="1-1000"
          />
        </Form.Item>
        {type === "subscription" && (
          <Form.Item
            name="planId"
            label="订阅计划"
            className="md:col-span-2"
            rules={[{ required: true, message: "请选择订阅计划" }]}
            extra={
              plans.length === 0 && !plansLoading
                ? "暂无可用的订阅计划，请先创建订阅计划"
                : undefined
            }
          >
            <Select
              disabled={plansLoading}
              placeholder={plansLoading ? "加载中..." : "选择订阅计划"}
              options={plans.map((plan) => ({
                value: plan.id,
                label: plan.title,
              }))}
            />
          </Form.Item>
        )}
        {type === "quota" && (
          <Form.Item
            name="credits"
            label="额度数量"
            className="md:col-span-2"
            rules={[{ required: true, message: "请输入额度数量" }]}
          >
            <InputNumber
              min={1}
              className="w-full"
              placeholder="请输入额度数量"
            />
          </Form.Item>
        )}
        <Form.Item
          name="expiresAt"
          label="过期时间"
          extra="留空则为永久有效"
          className="md:col-span-2"
        >
          <Input
            type="datetime-local"
            min={new Date().toISOString().slice(0, 16)}
          />
        </Form.Item>
      </div>
      {error && <Alert type="error" message={error} showIcon />}
    </Form>
  );
}
