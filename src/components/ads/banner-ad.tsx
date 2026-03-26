import { Card, CardContent } from "@/components/ui/card";
import { Megaphone } from "lucide-react";

interface BannerAdProps {
  className?: string;
}

export function BannerAd({ className = "" }: BannerAdProps) {
  return (
    <Card className={`border-2 border-dashed border-slate-200 bg-slate-50/50 hover:border-slate-300 transition-colors ${className}`}>
      <CardContent className="p-6">
        <div className="flex flex-col items-center justify-center text-center space-y-2 min-h-[80px]">
          <div className="flex items-center gap-2 text-slate-400">
            <Megaphone className="h-5 w-5" />
            <span className="text-lg font-medium">广告位招租</span>
          </div>
          <p className="text-sm text-slate-400">
            联系我们：advertising@example.com
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
