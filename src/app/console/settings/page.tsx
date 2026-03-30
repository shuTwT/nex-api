"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Save, Settings as SettingsIcon } from "lucide-react";
import { getSystemSettings, updateSystemSettings, getDefaultSettings } from "@/app/actions/system-settings";
import { toast } from "sonner";

interface SystemSetting {
  id: string;
  key: string;
  value: string;
  category: string;
  description: string | null;
  createdAt: string;
  updatedAt: string;
}

interface DefaultSetting {
  key: string;
  value: string;
  category: string;
  description: string;
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<Record<string, string>>({});
  const [defaultSettings, setDefaultSettings] = useState<Record<string, DefaultSetting[]>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadSettings();
    loadDefaultSettings();
  }, []);

  async function loadSettings() {
    const result = await getSystemSettings();
    if (result.success && result.data) {
      const settingsMap: Record<string, string> = {};
      result.data.forEach((s: SystemSetting) => {
        settingsMap[s.key] = s.value;
      });
      setSettings(settingsMap);
    }
    setIsLoading(false);
  }

  async function loadDefaultSettings() {
    const defaults = await getDefaultSettings();
    setDefaultSettings(defaults);
  }

  function handleSettingChange(key: string, value: string) {
    setSettings((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSave() {
    setIsSaving(true);

    const settingsToUpdate = Object.entries(settings).map(([key, value]) => ({
      key,
      value,
    }));

    const result = await updateSystemSettings(settingsToUpdate);
    if (result.success) {
      toast.success("设置保存成功");
    }
    setIsSaving(false);
  }

  function renderSetting(setting: DefaultSetting) {
    const value = settings[setting.key] ?? setting.value;

    if (value === "true" || value === "false") {
      return (
        <div key={setting.key} className="flex items-center justify-between py-4 border-b border-slate-200 last:border-0">
          <div>
            <Label className="font-medium text-slate-900">{setting.key}</Label>
            <p className="text-sm text-slate-500 mt-1">{setting.description}</p>
          </div>
          <input
            type="checkbox"
            checked={value === "true"}
            onChange={(e) =>
              handleSettingChange(setting.key, e.target.checked ? "true" : "false")
            }
            className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
          />
        </div>
      );
    }

    return (
      <div key={setting.key} className="space-y-2 py-4 border-b border-slate-200 last:border-0">
        <Label className="font-medium text-slate-900">{setting.key}</Label>
        <p className="text-sm text-slate-500">{setting.description}</p>
        <Input
          value={value}
          onChange={(e) => handleSettingChange(setting.key, e.target.value)}
          placeholder={setting.value}
        />
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">系统设置</h1>
          <p className="text-slate-500 mt-1">管理系统配置</p>
        </div>
        <Button className="gap-2 cursor-pointer" onClick={handleSave} disabled={isSaving}>
          <Save className="h-4 w-4" />
          {isSaving ? "保存中..." : "保存设置"}
        </Button>
      </div>

      <Tabs defaultValue="general" className="w-full">
        <TabsList className="grid w-full md:w-auto md:grid-cols-3">
          <TabsTrigger value="general" className="gap-2 cursor-pointer">
            <SettingsIcon className="h-4 w-4" />
            通用设置
          </TabsTrigger>
          <TabsTrigger value="operation" className="gap-2 cursor-pointer">
            <SettingsIcon className="h-4 w-4" />
            运营设置
          </TabsTrigger>
          <TabsTrigger value="payment" className="gap-2 cursor-pointer">
            <SettingsIcon className="h-4 w-4" />
            支付设置
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">通用设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.general?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="operation" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">运营设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.operation?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="payment" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">支付设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.payment?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
