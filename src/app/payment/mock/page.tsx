"use client";

import { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { MainLayout } from "@/components/main-layout";
import { api } from "@/lib/api-client";

export default function MockPaymentPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const outTradeNo = searchParams.get("outTradeNo");
  
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [payment, setPayment] = useState<any>(null);

  useEffect(() => {
    if (!outTradeNo) {
      router.push("/pricing");
      return;
    }

    loadPaymentInfo();
  }, [outTradeNo, router]);

  const loadPaymentInfo = async () => {
    try {
      const result = await api.get(`/api/payment/${outTradeNo}`);
      if (result.success && result.data) {
        setPayment(result.data);
      } else {
        router.push("/pricing");
      }
    } catch (err) {
      router.push("/pricing");
    } finally {
      setLoading(false);
    }
  };

  const handleMockPayment = async (success: boolean) => {
    setProcessing(true);
    try {
      const response = await fetch("/api/payment/callback/mock", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          outTradeNo,
          success,
        }),
      });

      const result = await response.json();

      if (result.success) {
        router.push(`/payment/result?outTradeNo=${outTradeNo}&status=${success ? "success" : "failed"}`);
      } else {
        alert("处理失败，请重试");
        setProcessing(false);
      }
    } catch (err) {
      console.error("模拟支付失败:", err);
      alert("处理失败，请重试");
      setProcessing(false);
    }
  };

  if (loading) {
    return (
      <MainLayout>
        <div className="container px-4 py-16 md:px-6 mx-auto">
          <div className="flex items-center justify-center min-h-[400px]">
            <Loader2 className="h-8 w-8 animate-spin text-cyan-500" />
          </div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="container px-4 py-16 md:px-6 mx-auto">
        <Card className="max-w-md mx-auto">
          <CardHeader className="text-center">
            <h2 className="text-2xl font-bold">模拟支付</h2>
            <p className="text-slate-600">测试环境 - 模拟支付流程</p>
          </CardHeader>

          <CardContent className="space-y-6">
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <p className="text-sm text-yellow-800">
                这是模拟支付环境，用于测试支付流程。实际生产环境中不会出现此页面。
              </p>
            </div>

            {payment && (
              <div className="bg-slate-50 rounded-lg p-4 space-y-3">
                <div className="flex justify-between">
                  <span className="text-slate-600">订单号</span>
                  <span className="font-mono text-sm">{payment.outTradeNo}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-600">支付金额</span>
                  <span className="font-bold text-lg">¥{payment.amount}</span>
                </div>
              </div>
            )}

            <div className="space-y-3">
              <Button
                onClick={() => handleMockPayment(true)}
                disabled={processing}
                className="w-full bg-green-600 hover:bg-green-700"
                size="lg"
              >
                {processing ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    处理中...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="mr-2 h-4 w-4" />
                    模拟支付成功
                  </>
                )}
              </Button>

              <Button
                onClick={() => handleMockPayment(false)}
                disabled={processing}
                variant="destructive"
                className="w-full"
                size="lg"
              >
                {processing ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    处理中...
                  </>
                ) : (
                  <>
                    <XCircle className="mr-2 h-4 w-4" />
                    模拟支付失败
                  </>
                )}
              </Button>
            </div>

            <div className="text-center">
              <Button
                variant="ghost"
                onClick={() => router.push(`/payment?outTradeNo=${outTradeNo}`)}
              >
                返回支付页面
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
}
