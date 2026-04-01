"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getCurrentUserProfile } from "@/app/actions/personal";
import {
  User,
  CreditCard,
  Activity,
  Calendar,
  Shield,
  Coins,
  Plus,
} from "lucide-react";
import { RechargeDialog } from "@/components/recharge-dialog";

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

  useEffect(() => {
    loadProfile();
  }, []);

  async function loadProfile() {
    setIsLoading(true);
    const result = await getCurrentUserProfile();
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

  const displayName = profile?.name || profile?.username;
  const initial = displayName ? displayName.charAt(0).toUpperCase() : "U";

  const statsCards = [
    {
      title: "当前余额",
      value: profile?.credits.toLocaleString() || 0,
      icon: Coins,
      color: "blue",
    },
    {
      title: "历史消耗",
      value: profile?.totalCreditsSpent.toLocaleString() || 0,
      icon: Activity,
      color: "green",
    },
    {
      title: "请求次数",
      value: profile?.totalRequests.toLocaleString() || 0,
      icon: CreditCard,
      color: "purple",
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

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">统计信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-3">
                {statsCards.map((stat, index) => {
                  const Icon = stat.icon;
                  const colorClasses = {
                    blue: "bg-blue-50 text-blue-600",
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
                            className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}
                          >
                            <Icon className="h-5 w-5" />
                          </div>
                        </div>
                        {index === 0 && (
                          <Button
                            size="sm"
                            className="w-full mt-3"
                            onClick={() => setRechargeDialogOpen(true)}
                          >
                            <Plus className="mr-1 h-4 w-4" />
                            充值
                          </Button>
                        )}
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
