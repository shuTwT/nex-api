import { useState } from "react";
import { Alert, Card, Checkbox, Form, Input, InputNumber, Select } from "antd";
import { api } from "@/lib/api";
import { ImageUpload } from "@/components/image-upload";
import { AdPositionOptions } from "@/types/ad-position";

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
interface AdvertisementFormValues {
  title: string;
  image: string;
  imageWidth?: number;
  imageHeight?: number;
  link: string;
  position: string;
  isActive: boolean;
}

export function AdvertisementForm({
  advertisement,
  onSuccess,
  formId,
}: AdvertisementFormProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isEdit = !!advertisement;

  async function handleFinish(values: AdvertisementFormValues) {
    setIsLoading(true);
    setError(null);
    const body = {
      ...values,
      imageWidth: values.imageWidth ?? 0,
      imageHeight: values.imageHeight ?? 0,
    };
    try {
      const result = isEdit
        ? await api.advertisements_id_route_put(
            { id: advertisement!.id! },
            body,
          )
        : await api.advertisements_route_post(body);
      if (result.success) onSuccess();
      else setError(result.error || "操作失败");
    } catch {
      setError("操作失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Form<AdvertisementFormValues>
      id={formId}
      layout="vertical"
      onFinish={handleFinish}
      disabled={isLoading}
      initialValues={{
        title: advertisement?.title ?? "",
        image: advertisement?.image ?? "",
        imageWidth: advertisement?.imageWidth ?? 0,
        imageHeight: advertisement?.imageHeight ?? 0,
        link: advertisement?.link ?? "",
        position: advertisement?.position ?? "",
        isActive: advertisement?.isActive ?? true,
      }}
    >
      <Card styles={{ body: { padding: 24 } }}>
        <Form.Item
          name="title"
          label="广告标题"
          rules={[{ required: true, message: "请输入广告标题" }]}
        >
          <Input placeholder="输入广告标题" />
        </Form.Item>
        <Form.Item
          name="image"
          label="广告图片"
          valuePropName="value"
          rules={[{ required: true, message: "请上传广告图片" }]}
        >
          <ImageUpload disabled={isLoading} />
        </Form.Item>
        <div className="grid gap-4 md:grid-cols-2">
          <Form.Item name="imageWidth" label="图片宽度 (px)">
            <InputNumber min={0} className="w-full" placeholder="例如：1200" />
          </Form.Item>
          <Form.Item name="imageHeight" label="图片高度 (px)">
            <InputNumber min={0} className="w-full" placeholder="例如：300" />
          </Form.Item>
        </div>
        <Form.Item
          name="link"
          label="跳转链接"
          rules={[
            { required: true, message: "请输入跳转链接" },
            { type: "url", message: "请输入有效的 URL" },
          ]}
        >
          <Input placeholder="https://example.com" />
        </Form.Item>
        <Form.Item
          name="position"
          label="广告位"
          rules={[{ required: true, message: "请选择广告位" }]}
        >
          <Select
            placeholder="选择广告位"
            options={AdPositionOptions.map((option) => ({
              value: option.value,
              label: option.label,
            }))}
          />
        </Form.Item>
        <Form.Item name="isActive" valuePropName="checked">
          <Checkbox>启用</Checkbox>
        </Form.Item>
        {error && <Alert type="error" message={error} showIcon />}
      </Card>
    </Form>
  );
}
