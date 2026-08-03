import { useEffect, useState } from "react";
import { Button, InputNumber, Modal, Radio, Space, Spin, Statistic, Typography } from "antd";
import { Coins } from "lucide-react";
import { api, responseData } from "@/lib/api";
import { toast } from "sonner";
import { useNavigate } from "react-router";

interface RechargeDialogProps { open: boolean; onOpenChange: (open: boolean) => void; }
interface PaymentSettings { creditPrice: number; minRecharge: number; alipayEnabled: boolean; wechatEnabled: boolean; mockEnabled: boolean; }

export function RechargeDialog({ open, onOpenChange }: RechargeDialogProps) {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [amount, setAmount] = useState<number | null>(null);
  const [paymentMethod, setPaymentMethod] = useState("");
  const [settings, setSettings] = useState<PaymentSettings | null>(null);
  useEffect(() => { if (open) loadSettings(); }, [open]);
  async function loadSettings() { setLoading(true); const result = await api.payment_settings_route_get(); const data = responseData<PaymentSettings>(result); if (result.success && data) { setSettings(data); setPaymentMethod(data.mockEnabled ? "mock" : data.alipayEnabled ? "alipay" : data.wechatEnabled ? "wechat" : ""); } else { toast.error("获取支付设置失败"); onOpenChange(false); } setLoading(false); }
  const credits = amount ? Math.floor(amount / (settings?.creditPrice || 1)) : 0;
  const isValidAmount = amount !== null && amount >= (settings?.minRecharge || 10);
  async function handleRecharge() { if (!isValidAmount || !paymentMethod || amount === null) return toast.error("请输入有效金额并选择支付方式"); setProcessing(true); try { const result = await api.recharge_route_post({ amount, credits, method: paymentMethod }); const data = responseData<{ outTradeNo: string }>(result); if (result.success && data) { toast.success("支付订单已创建"); onOpenChange(false); navigate(`/payment?outTradeNo=${data.outTradeNo}`); } else toast.error(result.error || "创建支付订单失败"); } catch { toast.error("充值失败，请重试"); } finally { setProcessing(false); } }
  const methodOptions = [{ enabled: settings?.mockEnabled, value: "mock", label: "🧪 模拟支付（测试用）" }, { enabled: settings?.wechatEnabled, value: "wechat", label: "💬 微信支付（推荐使用）" }, { enabled: settings?.alipayEnabled, value: "alipay", label: "💳 支付宝（安全便捷）" }].filter((option) => option.enabled);
  return <Modal open={open} title={<span className="flex items-center gap-2"><Coins size={20} />积分充值</span>} okText={`立即充值 ¥${amount || 0}`} cancelText="取消" onOk={handleRecharge} onCancel={() => onOpenChange(false)} confirmLoading={processing} okButtonProps={{ disabled: !isValidAmount || !paymentMethod }} destroyOnHidden>
    {loading ? <div className="flex justify-center py-8"><Spin size="large" /></div> : <Space direction="vertical" size="large" className="w-full"><Typography.Text type="secondary">充值积分以使用 API 服务</Typography.Text><div><Typography.Text>快捷充值</Typography.Text><Space wrap className="mt-2">{[10, 50, 100, 500].map((quickAmount) => <Button key={quickAmount} type={amount === quickAmount ? "primary" : "default"} onClick={() => setAmount(quickAmount)}>¥{quickAmount}</Button>)}</Space></div><div><Typography.Text>充值金额</Typography.Text><InputNumber className="mt-2 w-full" min={settings?.minRecharge} step={0.01} prefix="¥" placeholder={`最低 ${settings?.minRecharge} 元`} value={amount} onChange={setAmount} />{amount !== null && !isValidAmount && <Typography.Text type="danger">最低充值金额为 ¥{settings?.minRecharge}</Typography.Text>}</div>{isValidAmount && <Statistic title="可获得积分" value={credits} suffix="积分" /> }<div><Typography.Text>支付方式</Typography.Text><Radio.Group className="mt-2" value={paymentMethod} onChange={(event) => setPaymentMethod(event.target.value)} options={methodOptions} /></div></Space>}
  </Modal>;
}
