import { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { api, responseData } from "@/lib/api";

interface PaymentInfo {
  outTradeNo: string;
  amount: number;
  transactionId?: string;
  status: string;
}

export default function PaymentResultPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const outTradeNo = searchParams.get("outTradeNo");
  const status = searchParams.get("status");

  const [loading, setLoading] = useState(true);
  const [payment, setPayment] = useState<PaymentInfo | null>(null);

  useEffect(() => {
    if (!outTradeNo) {
      navigate("/pricing", { replace: true });
      return;
    }

    const loadPaymentInfo = async () => {
      try {
        const result = await api.payment_outTradeNo_route_get({ outTradeNo });
        const data = responseData<PaymentInfo>(result);
        if (result.success && data) {
          setPayment(data);
        }
      } catch {
      } finally {
        setLoading(false);
      }
    };

    loadPaymentInfo();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [outTradeNo]);

  if (loading) {
    return (
      <div className="container px-4 py-16 md:px-6 mx-auto">
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin text-cyan-500" />
        </div>
      </div>
    );
  }

  const isSuccess = status === "success" || payment?.status === "paid";

  return (
    <div className="container px-4 py-16 md:px-6 mx-auto">
      <Card className="max-w-md mx-auto">
        <CardContent className="p-8 text-center space-y-6">
          <div className="flex justify-center">
            {isSuccess ? (
              <CheckCircle2 className="h-20 w-20 text-green-500" />
            ) : (
              <XCircle className="h-20 w-20 text-red-500" />
            )}
          </div>

          <div className="space-y-2">
            <h2 className="text-2xl font-bold">
              {isSuccess ? "支付成功！" : "支付失败"}
            </h2>
            <p className="text-slate-600">
              {isSuccess
                ? "感谢您的购买，您的订阅已激活"
                : "支付过程中出现问题，请重试"}
            </p>
          </div>

          {payment && (
            <div className="bg-slate-50 rounded-lg p-4 space-y-2 text-left">
              <div className="flex justify-between text-sm">
                <span className="text-slate-600">订单号</span>
                <span className="font-mono">{payment.outTradeNo}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-600">支付金额</span>
                <span className="font-bold">¥{payment.amount}</span>
              </div>
              {payment.transactionId && (
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">交易号</span>
                  <span className="font-mono text-xs">{payment.transactionId}</span>
                </div>
              )}
            </div>
          )}

          <div className="flex flex-col gap-3">
            <Button
              onClick={() => navigate("/console/membership")}
              className="w-full"
              size="lg"
            >
              查看我的订阅
            </Button>
            <Button
              variant="outline"
              onClick={() => navigate("/pricing")}
              className="w-full"
            >
              返回定价页面
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
