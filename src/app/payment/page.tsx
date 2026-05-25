"use client";

import { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, XCircle, Clock, QrCode, ExternalLink, Loader2 } from "lucide-react";
import { MainLayout } from "@/components/main-layout";
import { api } from "@/lib/api-client";

interface PaymentState {
  outTradeNo: string;
  amount: number;
  transactionId?: string;
  status: string;
  method?: string;
  qrcodeUrl?: string;
  payUrl?: string;
  expiredAt?: string | number;
  paidAt?: string | number;
  createdAt?: string | number;
  plan?: { title?: string };
}

export default function PaymentPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const outTradeNo = searchParams.get("outTradeNo");
  
  const [loading, setLoading] = useState(true);
  const [payment, setPayment] = useState<PaymentState | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!outTradeNo) {
      setError("缺少订单号");
      setLoading(false);
      return;
    }

    loadPaymentInfo();
  }, [outTradeNo]);

  useEffect(() => {
    if (payment && payment.status === "pending") {
      const interval = setInterval(async () => {
        const result = await api.get(`/api/payment/${outTradeNo}/status`);
        if (result.success && result.data) {
          if (result.data.status !== "pending") {
            setPayment(result.data);
            if (result.data.status === "paid") {
              setTimeout(() => {
                router.push(`/payment/result?outTradeNo=${outTradeNo}&status=success`);
              }, 2000);
            }
          }
        }
      }, 3000);

      return () => clearInterval(interval);
    }
  }, [payment, outTradeNo, router]);

  const loadPaymentInfo = async () => {
    try {
      const result = await api.get(`/api/payment/${outTradeNo}`);
      if (result.success && result.data) {
        setPayment(result.data);
      } else {
        setError(result.error || "加载支付信息失败");
      }
    } catch (err) {
      setError("加载支付信息失败");
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = () => {
    if (!payment) return null;

    switch (payment.status) {
      case "pending":
        return <Badge variant="secondary">待支付</Badge>;
      case "paid":
        return <Badge variant="default" className="bg-green-500">已支付</Badge>;
      case "failed":
        return <Badge variant="destructive">支付失败</Badge>;
      case "cancelled":
        return <Badge variant="outline">已取消</Badge>;
      case "expired":
        return <Badge variant="outline">已过期</Badge>;
      default:
        return null;
    }
  };

  const getStatusIcon = () => {
    if (!payment) return null;

    switch (payment.status) {
      case "pending":
        return <Clock className="h-16 w-16 text-yellow-500" />;
      case "paid":
        return <CheckCircle2 className="h-16 w-16 text-green-500" />;
      case "failed":
      case "cancelled":
      case "expired":
        return <XCircle className="h-16 w-16 text-red-500" />;
      default:
        return null;
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

  if (error) {
    return (
      <MainLayout>
        <div className="container px-4 py-16 md:px-6 mx-auto">
          <Card className="max-w-md mx-auto">
            <CardContent className="p-8 text-center">
              <XCircle className="h-16 w-16 text-red-500 mx-auto mb-4" />
              <h2 className="text-xl font-bold mb-2">出错了</h2>
              <p className="text-slate-600 mb-4">{error}</p>
              <Button onClick={() => router.push("/pricing")}>返回定价页面</Button>
            </CardContent>
          </Card>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="container px-4 py-16 md:px-6 mx-auto">
        <Card className="max-w-lg mx-auto">
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              {getStatusIcon()}
            </div>
            <div className="flex items-center justify-center gap-2 mb-2">
              <h2 className="text-2xl font-bold">支付订单</h2>
              {getStatusBadge()}
            </div>
            <p className="text-slate-600">订单号：{payment?.outTradeNo}</p>
          </CardHeader>

          <CardContent className="space-y-6">
            <div className="bg-slate-50 rounded-lg p-4 space-y-3">
              <div className="flex justify-between">
                <span className="text-slate-600">订阅计划</span>
                <span className="font-semibold">{payment?.plan?.title || "自定义计划"}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-600">支付金额</span>
                <span className="font-bold text-lg">¥{payment?.amount}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-600">支付方式</span>
                <span className="font-semibold">
                  {payment?.method === "wechat" && "微信支付"}
                  {payment?.method === "alipay" && "支付宝"}
                  {payment?.method === "mock" && "模拟支付"}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-600">创建时间</span>
                <span className="text-sm text-slate-500">
                  {new Date(payment?.createdAt ?? Date.now()).toLocaleString("zh-CN")}
                </span>
              </div>
            </div>

            {payment?.status === "pending" && (
              <>
                {payment.method === "wechat" && payment.qrcodeUrl && (
                  <div className="text-center space-y-4">
                    <p className="text-sm text-slate-600">请使用微信扫描下方二维码完成支付</p>
                    <div className="bg-white p-4 rounded-lg inline-block border">
                      <QrCode className="h-48 w-48 text-slate-900" />
                    </div>
                    <p className="text-xs text-slate-500">
                      二维码有效期至 {new Date(payment.expiredAt!).toLocaleString("zh-CN")}
                    </p>
                  </div>
                )}

                {payment.method === "alipay" && payment.payUrl && (
                  <div className="text-center space-y-4">
                    <p className="text-sm text-slate-600">点击下方按钮跳转到支付宝完成支付</p>
                    <Button
                      onClick={() => window.open(payment.payUrl, "_blank")}
                      className="w-full"
                      size="lg"
                    >
                      <ExternalLink className="mr-2 h-4 w-4" />
                      前往支付宝支付
                    </Button>
                  </div>
                )}

                {payment.method === "mock" && payment.payUrl && (
                  <div className="text-center space-y-4">
                    <p className="text-sm text-slate-600">模拟支付模式</p>
                    <Button
                      onClick={() => router.push(payment.payUrl!)}
                      className="w-full"
                      size="lg"
                    >
                      前往模拟支付
                    </Button>
                  </div>
                )}

                <div className="flex items-center justify-center gap-2 text-sm text-slate-500">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>正在等待支付...</span>
                </div>
              </>
            )}

            {payment?.status === "paid" && (
              <div className="text-center space-y-4">
                <p className="text-green-600">支付成功！正在跳转...</p>
              </div>
            )}

            {(payment?.status === "failed" || payment?.status === "cancelled" || payment?.status === "expired") && (
              <div className="text-center space-y-4">
                <p className="text-red-600">
                  {payment?.status === "failed" && "支付失败，请重试"}
                  {payment?.status === "cancelled" && "订单已取消"}
                  {payment?.status === "expired" && "订单已过期"}
                </p>
                <Button onClick={() => router.push("/pricing")}>重新选择计划</Button>
              </div>
            )}
          </CardContent>

          <CardFooter className="flex justify-center gap-4">
            <Button variant="outline" onClick={() => router.push("/pricing")}>
              返回定价页面
            </Button>
            {payment?.status === "pending" && (
              <Button variant="ghost" onClick={() => router.push("/console/membership")}>
                查看我的订阅
              </Button>
            )}
          </CardFooter>
        </Card>
      </div>
    </MainLayout>
  );
}
