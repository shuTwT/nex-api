"use client";

import { useState, useRef, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Upload, X, Image as ImageIcon, Loader2 } from "lucide-react";

interface ImageUploadProps {
  value?: string;
  onChange: (url: string) => void;
  accept?: string;
  maxSize?: number;
  className?: string;
  disabled?: boolean;
}

export function ImageUpload({
  value,
  onChange,
  accept = "image/jpeg,image/png,image/gif,image/webp",
  maxSize = 10 * 1024 * 1024,
  className = "",
  disabled = false,
}: ImageUploadProps) {
  const [isUploading, setIsUploading] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleUpload = useCallback(async (file: File) => {
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
      const formData = new FormData();
      formData.append("file", file);

      const response = await fetch("/api/upload", {
        method: "POST",
        body: formData,
      });

      const result = await response.json();

      if (result.success && result.data?.url) {
        onChange(result.data.url);
      } else {
        setError(result.error || "上传失败");
      }
    } catch (err) {
      console.error("Upload error:", err);
      setError("上传失败，请重试");
    } finally {
      setIsUploading(false);
    }
  }, [maxSize, onChange, disabled]);

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleUpload(file);
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }, [handleUpload]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    if (!disabled) {
      setIsDragging(true);
    }
  }, [disabled]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);

    if (disabled) return;

    const file = e.dataTransfer.files[0];
    if (file) {
      handleUpload(file);
    }
  }, [handleUpload, disabled]);

  const handleClick = useCallback(() => {
    if (!disabled) {
      fileInputRef.current?.click();
    }
  }, [disabled]);

  const handleRemove = useCallback(() => {
    onChange("");
    setError(null);
  }, [onChange]);

  return (
    <div className={`space-y-2 ${className}`}>
      <div
        className={`
          relative border-2 border-dashed rounded-lg transition-colors
          ${isDragging ? "border-blue-500 bg-blue-50" : "border-slate-200"}
          ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer hover:border-slate-300"}
          ${value ? "p-2" : "p-6"}
        `}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleClick}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept={accept}
          onChange={handleFileSelect}
          className="hidden"
          disabled={disabled}
        />

        {isUploading ? (
          <div className="flex flex-col items-center justify-center py-4">
            <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
            <p className="mt-2 text-sm text-slate-600">上传中...</p>
          </div>
        ) : value ? (
          <div className="relative group">
            <img
              src={value}
              alt="Preview"
              className="w-full h-40 object-contain rounded-lg"
            />
            <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-all rounded-lg flex items-center justify-center pointer-events-none group-hover:pointer-events-auto">
              <Button
                type="button"
                variant="destructive"
                size="sm"
                className="opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={(e) => {
                  e.stopPropagation();
                  handleRemove();
                }}
                disabled={disabled}
              >
                <X className="h-4 w-4 mr-1" />
                删除
              </Button>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center">
            <div className="h-12 w-12 rounded-full bg-slate-100 flex items-center justify-center mb-3">
              {isDragging ? (
                <Upload className="h-6 w-6 text-blue-600" />
              ) : (
                <ImageIcon className="h-6 w-6 text-slate-400" />
              )}
            </div>
            <p className="text-sm text-slate-600 mb-1">
              {isDragging ? "松开以上传" : "点击或拖拽上传图片"}
            </p>
            <p className="text-xs text-slate-400">
              支持 JPG、PNG、GIF、WebP，最大 {(maxSize / 1024 / 1024).toFixed(0)}MB
            </p>
          </div>
        )}
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </div>
  );
}
