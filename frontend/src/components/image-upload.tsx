import { useCallback, useState } from "react";
import { Alert, Button, Image, Upload } from "antd";
import { Upload as UploadIcon, X } from "lucide-react";
import { uploadFile } from "@/lib/raw";

interface ImageUploadProps {
  value?: string;
  onChange?: (url: string) => void;
  accept?: string;
  maxSize?: number;
  className?: string;
  disabled?: boolean;
}

export function ImageUpload({
  value,
  onChange = () => undefined,
  accept = "image/jpeg,image/png,image/gif,image/webp",
  maxSize = 10 * 1024 * 1024,
  className = "",
  disabled = false,
}: ImageUploadProps) {
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleUpload = useCallback(
    async (file: File) => {
      if (disabled) return;

      setError(null);
      if (!file.type.startsWith("image/")) {
        setError("请上传图片文件");
        return;
      }
      if (file.size > maxSize) {
        setError(`文件大小不能超过 ${(maxSize / 1024 / 1024).toFixed(0)}MB`);
        return;
      }

      setIsUploading(true);
      try {
        const result = await uploadFile(file);
        if (
          result.success &&
          (result.data as { url?: string } | undefined)?.url
        ) {
          onChange((result.data as { url: string }).url);
        } else {
          setError(result.error || "上传失败");
        }
      } catch {
        setError("上传失败，请重试");
      } finally {
        setIsUploading(false);
      }
    },
    [disabled, maxSize, onChange],
  );

  return (
    <div className={`space-y-2 ${className}`}>
      {value ? (
        <div className="relative rounded-lg border border-slate-200 p-2">
          <Image
            src={value}
            alt="广告图片预览"
            className="h-40 w-full object-contain"
          />
          <Button
            type="primary"
            danger
            size="small"
            icon={<X className="h-4 w-4" />}
            onClick={() => {
              onChange("");
              setError(null);
            }}
            disabled={disabled}
            className="absolute right-4 top-4"
          >
            删除
          </Button>
        </div>
      ) : (
        <Upload.Dragger
          accept={accept}
          disabled={disabled || isUploading}
          showUploadList={false}
          multiple={false}
          beforeUpload={(file) => {
            void handleUpload(file as File);
            return false;
          }}
          className="!p-4"
        >
          <p className="ant-upload-drag-icon">
            <UploadIcon className="mx-auto h-8 w-8 text-blue-600" />
          </p>
          <p className="ant-upload-text">点击或拖拽上传图片</p>
          <p className="ant-upload-hint">
            支持 JPG、PNG、GIF、WebP，最大 {(maxSize / 1024 / 1024).toFixed(0)}
            MB
          </p>
        </Upload.Dragger>
      )}
      {isUploading && <p className="text-sm text-slate-500">上传中...</p>}
      {error && <Alert type="error" message={error} showIcon />}
    </div>
  );
}
