"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { getCurrentUserProfile } from "@/app/actions/personal";
import { User, CreditCard, Activity, Calendar, Shield, Coins } from "lucide-react";

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

  useEffect(() => {
    loadProfile();
  }, []);

  async function loadProfile() {
    setIsLoading(true);
    const result = await getCurrentUserProfile();
    if (result.success && result.data) {
      setProfile({
        ...result.data,
        createdAt: result.data.createdAt instanceof Date 
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

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardContent className="p-6">
            <div className="flex flex-col items-center text-center">
              <div className="relative">
                {profile?.image ? (
                  <img
                    src={profile.image}
                    alt={displayName || "Avatar"}
                    className="h-24 w-24 rounded-full object-cover"
                  />
                ) : (
                  <div className="h-24 w-24 rounded-full bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
                    <span className="text-white text-3xl font-bold">{initial}</span>
                  </div>
                )}
              </div>
              <h2 className="mt-4 text-xl font-bold text-slate-900">{displayName}</h2>
              <p className="text-slate-500">{profile?.email}</p>
              <div className="mt-4">
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
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="lg:col-span-2 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">统计信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-3">
                {statsCards.map((stat) => {
                  const Icon = stat.icon;
                  const colorClasses = {
                    blue: "bg-blue-50 text-blue-600",
                    green: "bg-green-50 text-green-600",
                    purple: "bg-purple-50 text-purple-600",
                  };
                  
                  return (
                    <Card key={stat.title} className="hover:shadow-md transition-shadow">
                      <CardContent className="p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="text-sm text-slate-500">{stat.title}</p>
                            <p className="text-2xl font-bold text-slate-900 mt-1">{stat.value}</p>
                          </div>
                          <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
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

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">基本信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center gap-3 py-3 border-b border-slate-200">
                  <User className="h-5 w-5 text-slate-400" />
                  <div className="flex-1">
                    <p className="text-sm text-slate-500">用户名</p>
                    <p className="text-sm font-medium text-slate-900">{profile?.username}</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 py-3 border-b border-slate-200">
                  <Shield className="h-5 w-5 text-slate-400" />
                  <div className="flex-1">
                    <p className="text-sm text-slate-500">用户 ID</p>
                    <p className="text-sm font-medium text-slate-900 font-mono">{profile?.id}</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 py-3">
                  <Calendar className="h-5 w-5 text-slate-400" />
                  <div className="flex-1">
                    <p className="text-sm text-slate-500">注册时间</p>
                    <p className="text-sm font-medium text-slate-900">
                      {profile?.createdAt ? new Date(profile.createdAt).toLocaleDateString("zh-CN") : "-"}
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
