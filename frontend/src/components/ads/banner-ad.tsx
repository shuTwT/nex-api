import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Megaphone } from "lucide-react";
import { AdPosition } from "@/types/ad-position";
import { rawGet } from "@/lib/raw";

interface BannerAdProps {
  className?: string;
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

export function BannerAd({ className = "", position }: BannerAdProps) {
  const [ad, setAd] = useState<Advertisement | null>(null);
  const [isLoading, setIsLoading] = useState(true);

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
        <CardContent className="p-6">
          <div className="flex items-center justify-center min-h-[80px]">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!ad) {
    return (
      <Card className={`border-2 border-dashed border-slate-200 bg-slate-50/50 hover:border-slate-300 transition-colors ${className}`}>
        <CardContent className="p-6">
          <div className="flex flex-col items-center justify-center text-center space-y-2 min-h-[80px]">
            <div className="flex items-center gap-2 text-slate-400">
              <Megaphone className="h-5 w-5" />
              <span className="text-lg font-medium">广告位招租</span>
            </div>
            <p className="text-sm text-slate-400">联系我们：advertising@example.com</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <a href={ad.link} target="_blank" rel="noopener noreferrer" className="block">
      <Card className={`overflow-hidden border-none hover:shadow-lg transition-shadow ${className}`}>
        <CardContent className="p-0">
          <div className="relative w-full min-h-[80px]">
            <img
              src={ad.image}
              alt={ad.title}
              className="w-full h-auto object-cover"
              style={{
                width: ad.imageWidth ? `${ad.imageWidth}px` : "auto",
                height: ad.imageHeight ? `${ad.imageHeight}px` : "auto",
              }}
            />
          </div>
        </CardContent>
      </Card>
    </a>
  );
}
