import { useMemo, useState } from "react";
import { Button, Empty, Input, Popover } from "antd";
import { ChevronDown, X } from "lucide-react";
import { ICON_REGISTRY, getIcon } from "@/lib/lucide-icons";

interface IconPickerProps {
  value?: string | null;
  onChange?: (value: string) => void;
  disabled?: boolean;
}

// antd Form.Item 会自动注入 value/onChange，可像 Input 一样直接使用。
export function IconPicker({ value, onChange, disabled }: IconPickerProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    const keyword = search.trim().toLowerCase();
    if (!keyword) return ICON_REGISTRY;
    return ICON_REGISTRY.filter(
      (entry) =>
        entry.name.toLowerCase().includes(keyword) ||
        entry.keywords.includes(keyword),
    );
  }, [search]);

  const SelectedIcon = getIcon(value);
  const hasValue = !!value && !!ICON_REGISTRY.some((entry) => entry.name === value);

  function close(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) setSearch("");
  }

  return (
    <Popover
      open={open}
      onOpenChange={close}
      trigger="click"
      placement="bottomLeft"
      arrow={false}
      content={
        <div className="w-[320px]">
          <Input
            allowClear
            placeholder="搜索图标名称或用途，如 zap / 闪电"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            className="mb-2"
          />
          <div className="grid max-h-60 grid-cols-8 gap-1 overflow-y-auto pr-1">
            {filtered.map((entry) => {
              const Icon = entry.icon;
              const selected = entry.name === value;
              return (
                <button
                  key={entry.name}
                  type="button"
                  title={`${entry.name} · ${entry.keywords.split(" ").slice(-2).join(" ")}`}
                  onClick={() => {
                    onChange?.(entry.name);
                    close(false);
                  }}
                  className={`flex h-9 w-9 items-center justify-center rounded-md border transition-colors ${
                    selected
                      ? "border-cyan-500 bg-cyan-50 text-cyan-600"
                      : "border-transparent text-slate-700 hover:border-slate-200 hover:bg-slate-100"
                  }`}
                >
                  <Icon size={18} />
                </button>
              );
            })}
          </div>
          {filtered.length === 0 && (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有匹配的图标" className="my-4" />
          )}
        </div>
      }
    >
      <Button disabled={disabled} className="w-full justify-between [&>span]:w-full">
        <span className="flex items-center justify-between">
          <span className="flex items-center gap-2">
            <SelectedIcon size={16} className="text-cyan-600" />
            <span className={hasValue ? "font-mono text-xs" : "text-slate-400"}>
              {hasValue ? value : "点击选择图标"}
            </span>
          </span>
          {hasValue ? (
            <X
              size={14}
              className="text-slate-400 hover:text-slate-600"
              onClick={(event) => {
                event.stopPropagation();
                onChange?.("");
              }}
            />
          ) : (
            <ChevronDown size={14} className="text-slate-400" />
          )}
        </span>
      </Button>
    </Popover>
  );
}
