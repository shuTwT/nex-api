import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Megaphone } from "lucide-react";
import { AdPosition } from "@/types/ad-position";
import { rawGet } from "@/lib/raw";
import { cn } from "@/lib/utils";

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
      <Card className={cn("border-dashed bg-muted/30", className)}>
        <CardContent className={cn("flex flex-col items-center justify-center", sizeClasses[size])}>
          <Skeleton className="size-8 rounded-full" />
        </CardContent>
      </Card>
    );
  }

  if (!ad) {
    return (
      <Card className={cn("border-dashed bg-muted/30 transition-colors hover:bg-muted/50", className)}>
        <CardContent className={cn("flex flex-col items-center justify-center gap-2 text-center", sizeClasses[size])}>
          <div className="flex items-center gap-2 text-muted-foreground">
            <Megaphone className="size-4" />
            <span className="font-medium">广告位招租</span>
          </div>
          <p className="text-xs text-muted-foreground">联系我们：advertising@example.com</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <a href={ad.link} target="_blank" rel="noopener noreferrer" className="block">
      <Card className={cn("overflow-hidden transition-shadow hover:shadow-lg", className)}>
        <CardContent className="p-0">
          <div className={cn("relative w-full", sizeClasses[size])}>
            <img
              src={ad.image}
              alt={ad.title}
              className="h-full w-full object-cover"
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
