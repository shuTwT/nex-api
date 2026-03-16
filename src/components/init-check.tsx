"use client";

import { useEffect, useState } from "react";
import { checkSystemInitialized } from "@/app/actions/system";
import { InitializationForm } from "@/components/initialization-form";

export function InitCheck({ children }: { children: React.ReactNode }) {
  const [isInitialized, setIsInitialized] = useState<boolean | null>(null);

  useEffect(() => {
    async function checkInit() {
      const result = await checkSystemInitialized();
      setIsInitialized(result.initialized);
    }
    checkInit();
  }, []);

  if (isInitialized === null) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!isInitialized) {
    return <InitializationForm />;
  }

  return <>{children}</>;
}
