"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { api } from "@/lib/api-client";
import { AdPosition, AdPositionOptions } from "@/types/ad-position";
import { ImageUpload } from "@/components/image-upload";

interface Advertisement {
  id?: string;
  image: string;
  imageWidth: number;
  imageHeight: number;
  link: string;
  title: string;
  position: string;
  isActive: boolean;
}

interface AdvertisementFormProps {
  advertisement?: Advertisement;
  onSuccess: () => void;
  formId: string;
}

export function AdvertisementForm({ advertisement, onSuccess, formId }: AdvertisementFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [position, setPosition] = useState(advertisement?.position || "");
  const [imageUrl, setImageUrl] = useState(advertisement?.image || "");

  useEffect(() => {
    if (advertisement) {
      setPosition(advertisement.position || "");
      setImageUrl(advertisement.image || "");
    } else {
      setPosition("");
      setImageUrl("");
    }
    setError(null);
  }, [advertisement]);

  const isEdit = !!advertisement;

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const formData = new FormData(e.currentTarget);
    if (position) formData.set("position", position);

    if (!imageUrl) {
      setError("请上传广告图片");
      setIsLoading(false);
      return;
    }

    formData.set("image", imageUrl);
    const body = Object.fromEntries(formData.entries());

    try {
      const result = isEdit
        ? await api.put(`/api/advertisements/${body.id}`, body)
        : await api.post("/api/advertisements", body);

      if (result.success) {
        onSuccess();
      } else {
        setError(result.error || "操作失败");
      }
    } catch (err) {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form id={formId} onSubmit={handleSubmit} className="space-y-4">
      {advertisement?.id && (
        <input type="hidden" name="id" value={advertisement.id} />
      )}

      <Card>
        <CardContent className="p-6 space-y-4">
          <div className="space-y-2">
            <Label htmlFor="title">广告标题 *</Label>
            <Input
              id="title"
              name="title"
              defaultValue={advertisement?.title || ""}
              required
              placeholder="输入广告标题"
            />
          </div>

          <div className="space-y-2">
            <Label>广告图片 *</Label>
            <ImageUpload
              value={imageUrl}
              onChange={setImageUrl}
              disabled={isLoading}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="imageWidth">图片宽度 (px)</Label>
              <Input
                id="imageWidth"
                name="imageWidth"
                type="number"
                min="0"
                defaultValue={advertisement?.imageWidth || 0}
                placeholder="例如：1200"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="imageHeight">图片高度 (px)</Label>
              <Input
                id="imageHeight"
                name="imageHeight"
                type="number"
                min="0"
                defaultValue={advertisement?.imageHeight || 0}
                placeholder="例如：300"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="link">跳转链接 *</Label>
            <Input
              id="link"
              name="link"
              type="url"
              defaultValue={advertisement?.link || ""}
              required
              placeholder="https://example.com"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="position">广告位 *</Label>
            <Select value={position} onValueChange={setPosition} disabled={isLoading}>
              <SelectTrigger id="position">
                <SelectValue placeholder="选择广告位" />
              </SelectTrigger>
              <SelectContent>
                {AdPositionOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="isActive"
              name="isActive"
              value="true"
              defaultChecked={advertisement?.isActive ?? true}
              className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
            />
            <Label htmlFor="isActive">启用</Label>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}
        </CardContent>
      </Card>
    </form>
  );
}
