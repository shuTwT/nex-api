import { useEffect, useState } from "react";
import { rawGet } from "@/lib/raw";
import { isOk } from "@/lib/api";

export interface OAuthProvider {
  id: string;
  name: string;
}

interface OAuthProvidersResponse {
  code: number;
  msg: string;
  data?: OAuthProvider[] | null;
}

export function useOAuthProviders() {
  const [providers, setProviders] = useState<OAuthProvider[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let active = true;

    void rawGet<OAuthProvidersResponse>("/api/auth/providers")
      .then((result) => {
        if (active && isOk(result) && Array.isArray(result.data)) {
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
