import { Card, CardContent } from "@/components/ui/card";
import { Megaphone } from "lucide-react";

interface InlineAdProps {
  className?: string;
  size?: "sm" | "md" | "lg";
}

export function InlineAd({ className = "", size = "md" }: InlineAdProps) {
  const sizeClasses = {
    sm: "min-h-[100px]",
    md: "min-h-[150px]",
    lg: "min-h-[200px]",
  };

  return (
    <Card className={`border-2 border-dashed border-slate-200 bg-slate-50/50 hover:border-slate-300 transition-colors ${className}`}>
      <CardContent className={`p-4 flex flex-col items-center justify-center text-center ${sizeClasses[size]}`}>
        <div className="flex items-center gap-2 text-slate-400 mb-2">
          <Megaphone className="h-4 w-4" />
          <span className="font-medium">广告位招租</span>
        </div>
        <p className="text-xs text-slate-400">
          联系我们：advertising@example.com
        </p>
      </CardContent>
    </Card>
  );
}
