"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Coins } from "lucide-react";
import { getPaymentSettings } from "@/app/actions/payment-settings";
import { createRechargePayment } from "@/app/actions/recharge";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface RechargeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RechargeDialog({ open, onOpenChange }: RechargeDialogProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [amount, setAmount] = useState("");
  const [paymentMethod, setPaymentMethod] = useState<string>("");
  const [settings, setSettings] = useState<{
    creditPrice: number;
    minRecharge: number;
    alipayEnabled: boolean;
    wechatEnabled: boolean;
  } | null>(null);

  useEffect(() => {
    if (open) {
      loadSettings();
    }
  }, [open]);

  async function loadSettings() {
    setLoading(true);
    const result = await getPaymentSettings();
    if (result.success && result.data) {
      setSettings(result.data);
      if (result.data.alipayEnabled) {
        setPaymentMethod("alipay");
      } else if (result.data.wechatEnabled) {
        setPaymentMethod("wechat");
      }
    } else {
      toast.error("获取支付设置失败");
      onOpenChange(false);
    }
    setLoading(false);
  }

  const credits = amount ? Math.floor(parseFloat(amount) / (settings?.creditPrice || 1)) : 0;
  const isValidAmount = amount && parseFloat(amount) >= (settings?.minRecharge || 10);

  async function handleRecharge() {
    if (!isValidAmount || !paymentMethod) {
      toast.error("请输入有效金额并选择支付方式");
      return;
    }

    setProcessing(true);
    try {
      const result = await createRechargePayment({
        amount: parseFloat(amount),
        credits,
        method: paymentMethod as any,
      });

      if (result.success && result.data) {
        toast.success("支付订单已创建");
        onOpenChange(false);
        router.push(`/payment?outTradeNo=${result.data.outTradeNo}`);
      } else {
        toast.error(result.error || "创建支付订单失败");
      }
    } catch (error) {
      console.error("充值失败:", error);
      toast.error("充值失败，请重试");
    } finally {
      setProcessing(false);
    }
  }

  const quickAmounts = [10, 50, 100, 500];

  if (loading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-cyan-500" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Coins className="h-5 w-5 text-cyan-500" />
            积分充值
          </DialogTitle>
          <DialogDescription>
            充值积分以使用 API 服务
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          <div className="space-y-2">
            <Label>快捷充值</Label>
            <div className="grid grid-cols-4 gap-2">
              {quickAmounts.map((amt) => (
                <Button
                  key={amt}
                  variant={amount === amt.toString() ? "default" : "outline"}
                  onClick={() => setAmount(amt.toString())}
                  className="w-full"
                >
                  ¥{amt}
                </Button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="amount">充值金额</Label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">¥</span>
              <Input
                id="amount"
                type="number"
                placeholder={`最低 ${settings?.minRecharge} 元`}
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="pl-8"
                min={settings?.minRecharge}
                step="0.01"
              />
            </div>
            {amount && !isValidAmount && (
              <p className="text-sm text-red-500">
                最低充值金额为 ¥{settings?.minRecharge}
              </p>
            )}
          </div>

          {isValidAmount && (
            <div className="bg-cyan-50 border border-cyan-200 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-cyan-700">可获得积分</span>
                <span className="text-2xl font-bold text-cyan-600">
                  {credits.toLocaleString()}
                </span>
              </div>
              <p className="text-xs text-cyan-600 mt-1">
                单价：¥{settings?.creditPrice}/积分
              </p>
            </div>
          )}

          <div className="space-y-2">
            <Label>支付方式</Label>
            <div className="grid grid-cols-2 gap-2">
              {settings?.wechatEnabled && (
                <Button
                  variant={paymentMethod === "wechat" ? "default" : "outline"}
                  onClick={() => setPaymentMethod("wechat")}
                  className="h-auto py-4 flex-col gap-1"
                >
                  <span className="text-2xl">💬</span>
                  <span className="font-medium">微信支付</span>
                  <span className="text-xs text-slate-500">推荐使用</span>
                </Button>
              )}
              {settings?.alipayEnabled && (
                <Button
                  variant={paymentMethod === "alipay" ? "default" : "outline"}
                  onClick={() => setPaymentMethod("alipay")}
                  className="h-auto py-4 flex-col gap-1"
                >
                  <span className="text-2xl">💳</span>
                  <span className="font-medium">支付宝</span>
                  <span className="text-xs text-slate-500">安全便捷</span>
                </Button>
              )}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleRecharge}
            disabled={!isValidAmount || !paymentMethod || processing}
          >
            {processing ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                处理中...
              </>
            ) : (
              `立即充值 ¥${amount || "0"}`
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
