import { useEffect, useState } from "react";
import { rawGet } from "@/lib/raw";

export interface OAuthProvider {
  id: string;
  name: string;
}

interface OAuthProvidersResponse {
  success: boolean;
  data?: OAuthProvider[] | null;
}

export function useOAuthProviders() {
  const [providers, setProviders] = useState<OAuthProvider[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let active = true;

    void rawGet<OAuthProvidersResponse>("/api/auth/providers")
      .then((result) => {
        if (active && result.success && Array.isArray(result.data)) {
          setProviders(result.data);
        }
      })
      .catch(() => {
        // OAuth is optional. A failed provider discovery must not block password login.
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  return { providers, isLoading };
}
