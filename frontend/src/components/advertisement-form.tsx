import { useState, useEffect, type FormEvent } from "react";
import { Alert, Card, Checkbox, Input, InputNumber, Select } from "antd";
import { api } from "@/lib/api";
import { AdPositionOptions } from "@/types/ad-position";
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

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
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
    const body: Record<string, string | number | boolean> = {};
    formData.forEach((value, key) => {
      body[key] = String(value);
    });
    body.imageWidth = Number(formData.get("imageWidth") ?? 0);
    body.imageHeight = Number(formData.get("imageHeight") ?? 0);
    body.isActive = formData.has("isActive");

    try {
      const result = isEdit
        ? await api.advertisements_id_route_put({ id: body.id ?? advertisement?.id ?? "" }, body)
        : await api.advertisements_route_post(body);

      if (result.success) {
        onSuccess();
      } else {
        setError(result.error || "操作失败");
      }
    } catch {
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

      <Card styles={{ body: { padding: 24 } }}>
        <div className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="title">广告标题 *</label>
            <Input
              id="title"
              name="title"
              defaultValue={advertisement?.title || ""}
              required
              placeholder="输入广告标题"
            />
          </div>

          <div className="space-y-2">
            <label>广告图片 *</label>
            <ImageUpload
              value={imageUrl}
              onChange={setImageUrl}
              disabled={isLoading}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <label htmlFor="imageWidth">图片宽度 (px)</label>
              <InputNumber
                id="imageWidth"
                name="imageWidth"
                type="number"
                min={0}
                defaultValue={advertisement?.imageWidth || 0}
                placeholder="例如：1200"
                className="w-full"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="imageHeight">图片高度 (px)</label>
              <InputNumber
                id="imageHeight"
                name="imageHeight"
                type="number"
                min={0}
                defaultValue={advertisement?.imageHeight || 0}
                placeholder="例如：300"
                className="w-full"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="link">跳转链接 *</label>
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
            <label htmlFor="position">广告位 *</label>
            <Select
              id="position"
              value={position}
              onChange={setPosition}
              disabled={isLoading}
              placeholder="选择广告位"
              options={AdPositionOptions.map((option) => ({ value: option.value, label: option.label }))}
            />
          </div>

          <div className="flex items-center gap-2">
            <Checkbox
              id="isActive"
              name="isActive"
              defaultChecked={advertisement?.isActive ?? true}
            >启用</Checkbox>
          </div>

          {error && (
            <Alert type="error" message={error} showIcon />
          )}
        </div>
      </Card>
    </form>
  );
}
