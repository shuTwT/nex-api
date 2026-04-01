"use client";

import { useEffect, useState, useRef } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Save, Settings as SettingsIcon, Upload, Plus, Trash2, ChevronDown, ChevronUp } from "lucide-react";
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

interface PaymentSettings {
  basic: DefaultSetting[];
  alipay: DefaultSetting[];
  wechat: DefaultSetting[];
}

interface OperationSettings {
  basic: DefaultSetting[];
  announcement: DefaultSetting[];
}

interface DefaultSettings {
  general: DefaultSetting[];
  operation: OperationSettings;
  payment: PaymentSettings;
  oauth?: {
    basic: DefaultSetting[];
  };
}

interface OAuthProvider {
  id: string;
  name: string;
  clientId: string;
  clientSecret: string;
  authorizationUrl: string;
  tokenUrl: string;
  userInfoUrl: string;
  scopes: string;
  userIdField: string;
  emailField: string;
  usernameField: string;
  roleField: string;
}

const PEM_FILE_FIELDS = [
  'wechatPayPrivateKey',
  'wechatPayPublicKey', 
  'wechatPayPaymentPublicKey'
];

export default function SettingsPage() {
  const [settings, setSettings] = useState<Record<string, string>>({});
  const [defaultSettings, setDefaultSettings] = useState<DefaultSettings | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const fileInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const [oauthProviders, setOauthProviders] = useState<OAuthProvider[]>([]);
  const [expandedProviders, setExpandedProviders] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadSettings();
    loadDefaultSettings();
  }, []);

  async function loadSettings() {
    const result = await getSystemSettings();
    if (result.success && result.data) {
      const settingsMap: Record<string, string> = {};
      result.data.forEach((s: any) => {
        settingsMap[s.key] = s.value;
      });
      setSettings(settingsMap);
      
      try {
        if (settingsMap.oauthProviders) {
          setOauthProviders(JSON.parse(settingsMap.oauthProviders));
        }
      } catch (e) {
        console.error("Failed to parse oauthProviders:", e);
      }
    }
    setIsLoading(false);
  }

  async function loadDefaultSettings() {
    const defaults = await getDefaultSettings();
    setDefaultSettings(defaults as DefaultSettings);
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
    
    settingsToUpdate.push({
      key: "oauthProviders",
      value: JSON.stringify(oauthProviders),
    });

    const result = await updateSystemSettings(settingsToUpdate);
    if (result.success) {
      toast.success("设置保存成功");
    }
    setIsSaving(false);
  }

  function addOAuthProvider() {
    const newProvider: OAuthProvider = {
      id: crypto.randomUUID(),
      name: "",
      clientId: "",
      clientSecret: "",
      authorizationUrl: "",
      tokenUrl: "",
      userInfoUrl: "",
      scopes: "",
      userIdField: "",
      emailField: "",
      usernameField: "",
      roleField: "",
    };
    setOauthProviders([...oauthProviders, newProvider]);
    setExpandedProviders(new Set([...expandedProviders, newProvider.id]));
  }

  function removeOAuthProvider(id: string) {
    setOauthProviders(oauthProviders.filter(p => p.id !== id));
    const newExpanded = new Set(expandedProviders);
    newExpanded.delete(id);
    setExpandedProviders(newExpanded);
  }

  function updateOAuthProvider(id: string, field: keyof OAuthProvider, value: string) {
    setOauthProviders(oauthProviders.map(p => 
      p.id === id ? { ...p, [field]: value } : p
    ));
  }

  function toggleProviderExpand(id: string) {
    const newExpanded = new Set(expandedProviders);
    if (newExpanded.has(id)) {
      newExpanded.delete(id);
    } else {
      newExpanded.add(id);
    }
    setExpandedProviders(newExpanded);
  }

  function handleFileUpload(key: string) {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.pem';
    
    input.onchange = (e: Event) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = (event) => {
        const content = event.target?.result as string;
        handleSettingChange(key, content.trim());
        toast.success(`已成功读取 ${file.name}`);
      };
      reader.onerror = () => {
        toast.error('文件读取失败');
      };
      reader.readAsText(file);
    };

    input.click();
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

    if (setting.key === "announcementContent") {
      return (
        <div key={setting.key} className="space-y-2 py-4 border-b border-slate-200 last:border-0">
          <Label className="font-medium text-slate-900">{setting.key}</Label>
          <p className="text-sm text-slate-500">{setting.description}</p>
          <textarea
            value={value}
            onChange={(e) => handleSettingChange(setting.key, e.target.value)}
            placeholder={setting.value}
            rows={8}
            className="w-full px-3 py-2 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-y"
          />
        </div>
      );
    }

    if (PEM_FILE_FIELDS.includes(setting.key)) {
      return (
        <div key={setting.key} className="space-y-2 py-4 border-b border-slate-200 last:border-0">
          <div className="flex items-center justify-between">
            <div>
              <Label className="font-medium text-slate-900">{setting.key}</Label>
              <p className="text-sm text-slate-500">{setting.description}</p>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => handleFileUpload(setting.key)}
              className="gap-2"
            >
              <Upload className="h-4 w-4" />
              上传 .pem 文件
            </Button>
          </div>
          <textarea
            value={value}
            onChange={(e) => handleSettingChange(setting.key, e.target.value)}
            placeholder={setting.value}
            rows={6}
            className="w-full px-3 py-2 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono resize-y"
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

  if (isLoading || !defaultSettings) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
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
        <TabsList className="grid w-full md:w-auto md:grid-cols-4">
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
          <TabsTrigger value="oauth" className="gap-2 cursor-pointer">
            <SettingsIcon className="h-4 w-4" />
            OAuth 设置
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

        <TabsContent value="operation" className="mt-6 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">基本设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.operation?.basic?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">公告设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.operation?.announcement?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="payment" className="mt-6 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">基本设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.payment.basic?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">支付宝设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.payment.alipay?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">微信设置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {defaultSettings.payment.wechat?.map(renderSetting)}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="oauth" className="mt-6 space-y-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-lg">OAuth 提供商</CardTitle>
              <Button className="gap-2 cursor-pointer" onClick={addOAuthProvider}>
                <Plus className="h-4 w-4" />
                添加提供商
              </Button>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {oauthProviders.length === 0 ? (
                  <p className="text-sm text-slate-500 text-center py-8">
                    暂无 OAuth 提供商，点击上方按钮添加
                  </p>
                ) : (
                  oauthProviders.map((provider) => (
                    <Card key={provider.id} className="border border-slate-200">
                      <CardHeader className="py-4 px-4 flex flex-row items-center justify-between">
                        <div className="flex items-center gap-3">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="p-0 h-8 w-8"
                            onClick={() => toggleProviderExpand(provider.id)}
                          >
                            {expandedProviders.has(provider.id) ? (
                              <ChevronUp className="h-4 w-4" />
                            ) : (
                              <ChevronDown className="h-4 w-4" />
                            )}
                          </Button>
                          <div>
                            <CardTitle className="text-base">
                              {provider.name || "未命名提供商"}
                            </CardTitle>
                          </div>
                        </div>
                        <Button
                          variant="destructive"
                          size="sm"
                          className="gap-2 cursor-pointer"
                          onClick={() => removeOAuthProvider(provider.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                          删除
                        </Button>
                      </CardHeader>
                      {expandedProviders.has(provider.id) && (
                        <CardContent className="pt-0 px-4 pb-4 space-y-4">
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="space-y-2">
                              <Label htmlFor={`name-${provider.id}`}>提供商名称</Label>
                              <Input
                                id={`name-${provider.id}`}
                                value={provider.name}
                                onChange={(e) => updateOAuthProvider(provider.id, "name", e.target.value)}
                                placeholder="例如: GitHub, Google"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor={`clientId-${provider.id}`}>Client ID</Label>
                              <Input
                                id={`clientId-${provider.id}`}
                                value={provider.clientId}
                                onChange={(e) => updateOAuthProvider(provider.id, "clientId", e.target.value)}
                                placeholder="Client ID"
                              />
                            </div>
                            <div className="space-y-2 md:col-span-2">
                              <Label htmlFor={`clientSecret-${provider.id}`}>Client Secret</Label>
                              <Input
                                id={`clientSecret-${provider.id}`}
                                type="password"
                                value={provider.clientSecret}
                                onChange={(e) => updateOAuthProvider(provider.id, "clientSecret", e.target.value)}
                                placeholder="Client Secret"
                              />
                            </div>
                            <div className="space-y-2 md:col-span-2">
                              <Label htmlFor={`authorizationUrl-${provider.id}`}>AUTHORIZATION_URL</Label>
                              <Input
                                id={`authorizationUrl-${provider.id}`}
                                value={provider.authorizationUrl}
                                onChange={(e) => updateOAuthProvider(provider.id, "authorizationUrl", e.target.value)}
                                placeholder="https://example.com/oauth/authorize"
                              />
                            </div>
                            <div className="space-y-2 md:col-span-2">
                              <Label htmlFor={`tokenUrl-${provider.id}`}>TOKEN_URL</Label>
                              <Input
                                id={`tokenUrl-${provider.id}`}
                                value={provider.tokenUrl}
                                onChange={(e) => updateOAuthProvider(provider.id, "tokenUrl", e.target.value)}
                                placeholder="https://example.com/oauth/token"
                              />
                            </div>
                            <div className="space-y-2 md:col-span-2">
                              <Label htmlFor={`userInfoUrl-${provider.id}`}>USER_INFO_URL</Label>
                              <Input
                                id={`userInfoUrl-${provider.id}`}
                                value={provider.userInfoUrl}
                                onChange={(e) => updateOAuthProvider(provider.id, "userInfoUrl", e.target.value)}
                                placeholder="https://example.com/api/user"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor={`scopes-${provider.id}`}>Scopes</Label>
                              <Input
                                id={`scopes-${provider.id}`}
                                value={provider.scopes}
                                onChange={(e) => updateOAuthProvider(provider.id, "scopes", e.target.value)}
                                placeholder="read,write,profile"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor={`userIdField-${provider.id}`}>用户 ID 字段</Label>
                              <Input
                                id={`userIdField-${provider.id}`}
                                value={provider.userIdField}
                                onChange={(e) => updateOAuthProvider(provider.id, "userIdField", e.target.value)}
                                placeholder="id"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor={`emailField-${provider.id}`}>邮箱字段</Label>
                              <Input
                                id={`emailField-${provider.id}`}
                                value={provider.emailField}
                                onChange={(e) => updateOAuthProvider(provider.id, "emailField", e.target.value)}
                                placeholder="email"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor={`usernameField-${provider.id}`}>用户名字段</Label>
                              <Input
                                id={`usernameField-${provider.id}`}
                                value={provider.usernameField}
                                onChange={(e) => updateOAuthProvider(provider.id, "usernameField", e.target.value)}
                                placeholder="username"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor={`roleField-${provider.id}`}>角色字段</Label>
                              <Input
                                id={`roleField-${provider.id}`}
                                value={provider.roleField}
                                onChange={(e) => updateOAuthProvider(provider.id, "roleField", e.target.value)}
                                placeholder="role"
                              />
                            </div>
                          </div>
                        </CardContent>
                      )}
                    </Card>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
