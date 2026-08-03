import { ConfigProvider, theme as antdTheme } from "antd";
import type { ReactNode } from "react";
import { useTheme } from "@/providers/theme";

interface AntdThemeProviderProps {
  children: ReactNode;
}

export function AntdThemeProvider({ children }: AntdThemeProviderProps) {
  const { resolvedTheme } = useTheme();

  return (
    <ConfigProvider
      theme={{
        algorithm:
          resolvedTheme === "dark"
            ? antdTheme.darkAlgorithm
            : antdTheme.defaultAlgorithm,
      }}
    >
      {children}
    </ConfigProvider>
  );
}
