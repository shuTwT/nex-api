"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { api } from "@/lib/api-client";
import {
  CreditCard,
  Activity,
  Coins,
  Plus,
  Wallet,
  Ticket,
  Gift,
  AlertTriangle,
} from "lucide-react";
import { RechargeDialog } from "@/components/recharge-dialog";
import { toast } from "sonner";

interface UserProfile {
  id: string;
  name: string | null;
  email: string;
  image: string | null;
  username: string;
  role: string;
  credits: number;
  createdAt: string;
  totalCreditsSpent: number;
  totalRequests: number;
}

export default function PersonalPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [rechargeDialogOpen, setRechargeDialogOpen] = useState(false);
  const [redeemInput, setRedeemInput] = useState("");
  const [isRedeeming, setIsRedeeming] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    type: string;
    planName: string | null;
    credits: number | null;
  }>({ open: false, type: "", planName: null, credits: null });

  useEffect(() => {
    loadProfile();
  }, []);

  async function loadProfile() {
    setIsLoading(true);
    const result = await api.get("/api/personal/profile");
    if (result.success && result.data) {
      setProfile({
        ...result.data,
        createdAt:
          result.data.createdAt instanceof Date
            ? result.data.createdAt.toISOString()
            : result.data.createdAt,
      });
    }
    setIsLoading(false);
  }

  async function handleRedeem() {
    if (!redeemInput.trim()) return;
    setIsRedeeming(true);
    const result = await api.post("/api/personal/redeem/lookup", { code: redeemInput });
    if (result.success && result.data) {
      setConfirmDialog({
        open: true,
        type: result.data.type,
        planName: result.data.planName,
        credits: result.data.credits,
      });
    } else {
      toast.error(result.error || "查询失败");
    }
    setIsRedeeming(false);
  }

  async function handleConfirmRedeem() {
    setIsRedeeming(true);
    const result = await api.post("/api/personal/redeem", { code: redeemInput });
    if (result.success) {
      toast.success(result.message || "兑换成功");
      setRedeemInput("");
      setConfirmDialog({ open: false, type: "", planName: null, credits: null });
      loadProfile();
    } else {
      toast.error(result.error || "兑换失败");
    }
    setIsRedeeming(false);
  }

  const displayName = profile?.name || profile?.username;
  const initial = displayName ? displayName.charAt(0).toUpperCase() : "U";

  const balance = profile?.credits.toLocaleString() || "0";

  const statsCards = [
    {
      title: "历史消耗",
      value: profile?.totalCreditsSpent.toLocaleString() || "0",
      icon: Activity,
      color: "green" as const,
    },
    {
      title: "请求次数",
      value: profile?.totalRequests.toLocaleString() || "0",
      icon: CreditCard,
      color: "purple" as const,
    },
  ];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">个人中心</h1>
        <p className="text-slate-500 mt-1">查看和管理您的个人信息</p>
      </div>

      <RechargeDialog
        open={rechargeDialogOpen}
        onOpenChange={setRechargeDialogOpen}
      />

      <div className="flex flex-col gap-6 ">
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col">
              <div className="relative flex">
                <div className="relative">
                  {profile?.image ? (
                    <img
                      src={profile.image}
                      alt={displayName || "Avatar"}
                      className="h-24 w-24 rounded-full object-cover"
                    />
                  ) : (
                    <div className="h-24 w-24 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
                      <span className="text-white text-3xl font-bold">
                        {initial}
                      </span>
                    </div>
                  )}
                </div>
                <div className="relative flex flex-col ml-4 justify-around">
                  <h2 className="text-xl font-bold text-slate-900">
                    {displayName}
                  </h2>
                  <div className="flex gap-2">
                    <Badge variant="outline">{profile?.email}</Badge>
                    <Badge
                      variant="outline"
                      className={
                        profile?.role === "admin"
                          ? "bg-purple-50 text-purple-700 border-purple-200"
                          : "bg-gray-50 text-gray-700 border-gray-200"
                      }
                    >
                      {profile?.role === "admin" ? "管理员" : "普通用户"}
                    </Badge>
                    <Badge variant="outline">
                      ID: {profile?.id}
                    </Badge>
                    <Badge variant="outline">
                      注册时间: {profile?.createdAt
                        ? new Date(profile.createdAt).toLocaleDateString(
                            "zh-CN",
                          )
                        : "-"}
                    </Badge>
                  </div>
                </div>
              </div>
              <div className="relative flex"></div>
            </div>
          </CardContent>
        </Card>

        <div className="flex flex-col md:flex-row gap-6">
          <Card className="md:w-80 md:shrink-0">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Wallet className="h-5 w-5" />
                钱包
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-lg bg-blue-50 p-4">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center">
                    <Coins className="h-5 w-5 text-blue-600" />
                  </div>
                  <div>
                    <p className="text-sm text-slate-500">当前余额</p>
                    <p className="text-2xl font-bold text-slate-900">{balance}</p>
                  </div>
                </div>
                <Button
                  size="sm"
                  className="gap-1 cursor-pointer"
                  onClick={() => setRechargeDialogOpen(true)}
                >
                  <Plus className="h-4 w-4" />
                  充值
                </Button>
              </div>

              <div className="flex gap-3">
                <Input
                  placeholder="请输入兑换码"
                  value={redeemInput}
                  onChange={(e) => setRedeemInput(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleRedeem()}
                  className="flex-1 uppercase"
                  disabled={isRedeeming}
                />
                <Button
                  onClick={handleRedeem}
                  disabled={isRedeeming || !redeemInput.trim()}
                  className="cursor-pointer gap-1"
                >
                  <Ticket className="h-4 w-4" />
                  {isRedeeming ? "兑换中..." : "兑换"}
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card className="md:min-w-0 md:flex-1">
            <CardHeader>
              <CardTitle className="text-lg">统计信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                {statsCards.map((stat) => {
                  const Icon = stat.icon;
                  const colorClasses = {
                    green: "bg-green-50 text-green-600",
                    purple: "bg-purple-50 text-purple-600",
                  };

                  return (
                    <Card
                      key={stat.title}
                      className="hover:shadow-md transition-shadow"
                    >
                      <CardContent className="p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="text-sm text-slate-500">
                              {stat.title}
                            </p>
                            <p className="text-2xl font-bold text-slate-900 mt-1">
                              {stat.value}
                            </p>
                          </div>
                          <div
                            className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color]}`}
                          >
                            <Icon className="h-5 w-5" />
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <Dialog
        open={confirmDialog.open}
        onOpenChange={(open) => {
          if (!open) setConfirmDialog((prev) => ({ ...prev, open: false }));
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-amber-500" />
              确认兑换
            </DialogTitle>
            <DialogDescription>
              您即将使用兑换码，请确认以下信息：
            </DialogDescription>
          </DialogHeader>

          <div className="rounded-lg bg-slate-50 p-4 space-y-2">
            {confirmDialog.type === "subscription" ? (
              <>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-slate-500">类型：</span>
                  <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                    <Gift className="h-3 w-3 mr-1" />
                    订阅兑换码
                  </Badge>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-slate-500">订阅计划：</span>
                  <span className="font-medium text-slate-900">
                    {confirmDialog.planName || "-"}
                  </span>
                </div>
              </>
            ) : (
              <>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-slate-500">类型：</span>
                  <Badge variant="outline" className="bg-purple-50 text-purple-700 border-purple-200">
                    <Coins className="h-3 w-3 mr-1" />
                    额度兑换码
                  </Badge>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-slate-500">获得额度：</span>
                  <span className="font-medium text-slate-900">
                    {(confirmDialog.credits || 0).toLocaleString()} 积分
                  </span>
                </div>
              </>
            )}
          </div>

          <p className="text-sm text-amber-600">
            兑换后不可撤销，确认要继续吗？
          </p>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                setConfirmDialog((prev) => ({ ...prev, open: false }))
              }
              disabled={isRedeeming}
              className="cursor-pointer"
            >
              取消
            </Button>
            <Button
              type="button"
              onClick={handleConfirmRedeem}
              disabled={isRedeeming}
              className="cursor-pointer"
            >
              {isRedeeming ? "兑换中..." : "确认兑换"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
