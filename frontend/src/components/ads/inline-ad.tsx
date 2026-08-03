import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Megaphone } from "lucide-react";
import { AdPosition } from "@/types/ad-position";
import { rawGet } from "@/lib/raw";

interface InlineAdProps {
  className?: string;
  size?: "sm" | "md" | "lg";
  position: AdPosition;
}

interface Advertisement {
  id: string;
  image: string;
  imageWidth?: number;
  imageHeight?: number;
  link: string;
  title: string;
  isActive: boolean;
}

export function InlineAd({ className = "", size = "md", position }: InlineAdProps) {
  const [ad, setAd] = useState<Advertisement | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const sizeClasses = {
    sm: "min-h-[100px]",
    md: "min-h-[150px]",
    lg: "min-h-[200px]",
  };

  useEffect(() => {
    let cancelled = false;
    async function loadAd() {
      try {
        const result = await rawGet<{ success: boolean; data?: Advertisement }>(
          `/api/advertisements/by-position/${position}`,
        );
        if (!cancelled) {
          setAd(result.success && result.data ? result.data : null);
        }
      } catch {
        if (!cancelled) setAd(null);
      }
      if (!cancelled) setIsLoading(false);
    }
    loadAd();
    return () => {
      cancelled = true;
    };
  }, [position]);

  if (isLoading) {
    return (
      <Card className={`border-2 border-dashed border-slate-200 bg-slate-50/50 ${className}`}>
        <CardContent className={`p-4 flex flex-col items-center justify-center ${sizeClasses[size]}`}>
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600" />
        </CardContent>
      </Card>
    );
  }

  if (!ad) {
    return (
      <Card className={`border-2 border-dashed border-slate-200 bg-slate-50/50 hover:border-slate-300 transition-colors ${className}`}>
        <CardContent className={`p-4 flex flex-col items-center justify-center text-center ${sizeClasses[size]}`}>
          <div className="flex items-center gap-2 text-slate-400 mb-2">
            <Megaphone className="h-4 w-4" />
            <span className="font-medium">广告位招租</span>
          </div>
          <p className="text-xs text-slate-400">联系我们：advertising@example.com</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <a href={ad.link} target="_blank" rel="noopener noreferrer" className="block">
      <Card className={`overflow-hidden border-none hover:shadow-lg transition-shadow ${className}`}>
        <CardContent className="p-0">
          <div className={`relative w-full ${sizeClasses[size]}`}>
            <img
              src={ad.image}
              alt={ad.title}
              className="w-full h-full object-cover"
              style={{
                width: ad.imageWidth ? `${ad.imageWidth}px` : "100%",
                height: ad.imageHeight ? `${ad.imageHeight}px` : "auto",
              }}
            />
          </div>
        </CardContent>
      </Card>
    </a>
  );
}
