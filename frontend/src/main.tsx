import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { router } from "@/router";
import { ThemeProvider } from "@/providers/theme";
import { AuthProvider } from "@/providers/auth";
import { ToastProvider } from "@/providers/toast";
import { ErrorBoundary } from "@/layouts/error-boundary";
import "@/index.css";

const container = document.getElementById("root");
if (!container) {
  throw new Error("Root element #root not found");
}

createRoot(container).render(
  <StrictMode>
    <ErrorBoundary>
      <ThemeProvider
        attribute="class"
        defaultTheme="light"
        enableSystem
        disableTransitionOnChange
      >
        <AuthProvider>
          <RouterProvider router={router} />
          <ToastProvider />
        </AuthProvider>
      </ThemeProvider>
    </ErrorBoundary>
  </StrictMode>,
);