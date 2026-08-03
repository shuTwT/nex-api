import { createBrowserRouter, type RouteObject } from "react-router";
import { PublicLayout } from "@/layouts/public";
import { ConsoleLayout } from "@/layouts/console";
import { ErrorPage } from "@/pages/error";
import { InitializationGuard } from "@/components/initialization-guard";

/**
 * Route tree with 30 page URLs.
 *
 * Uses React Router v7 `lazy` for automatic code-splitting — each page
 * module is loaded on demand and React Router handles the Suspense boundary.
 */

const lazyRoute = (
  importer: () => Promise<{ default: React.ComponentType }>,
): Pick<RouteObject, "lazy"> => ({
  lazy: async () => {
    const module = await importer();
    return { Component: module.default };
  },
});

const publicRoutes: RouteObject[] = [
  {
    index: true,
    ...lazyRoute(() => import("@/pages/home")),
  },
  {
    path: "api-detail",
    ...lazyRoute(() => import("@/pages/api-detail")),
  },
  {
    path: "api-market",
    ...lazyRoute(() => import("@/pages/api-market")),
  },
  {
    path: "mcp-market",
    ...lazyRoute(() => import("@/pages/mcp-market")),
  },
  {
    path: "pricing",
    ...lazyRoute(() => import("@/pages/pricing")),
  },
  {
    path: "payment",
    ...lazyRoute(() => import("@/pages/payment")),
  },
  {
    path: "payment/mock",
    ...lazyRoute(() => import("@/pages/payment-mock")),
  },
  {
    path: "payment/result",
    ...lazyRoute(() => import("@/pages/payment-result")),
  },
  {
    path: "auth/signin",
    ...lazyRoute(() => import("@/pages/auth-signin")),
  },
  {
    path: "auth/signout",
    ...lazyRoute(() => import("@/pages/auth-signout")),
  },
  {
    path: "auth/verify-request",
    ...lazyRoute(() => import("@/pages/auth-verify-request")),
  },
  {
    path: "auth/error",
    ...lazyRoute(() => import("@/pages/auth-error")),
  },
  {
    path: "initialize",
    ...lazyRoute(() => import("@/pages/initialize")),
  },
  {
    path: "privacy",
    ...lazyRoute(() => import("@/pages/privacy")),
  },
  {
    path: "terms",
    ...lazyRoute(() => import("@/pages/terms")),
  },
  {
    path: "unauthorized",
    ...lazyRoute(() => import("@/pages/unauthorized")),
  },
  {
    path: "forbidden",
    ...lazyRoute(() => import("@/pages/forbidden")),
  },
];

const consoleRoutes: RouteObject[] = [
  {
    index: true,
    ...lazyRoute(() => import("@/pages/console/dashboard")),
  },
  {
    path: "personal",
    ...lazyRoute(() => import("@/pages/console/personal")),
  },
  {
    path: "membership",
    ...lazyRoute(() => import("@/pages/console/membership")),
  },
  {
    path: "tokens",
    ...lazyRoute(() => import("@/pages/console/tokens")),
  },
  {
    path: "api-management",
    ...lazyRoute(() => import("@/pages/console/api-management")),
  },
  {
    path: "mcp-services",
    ...lazyRoute(() => import("@/pages/console/mcp-services")),
  },
  {
    path: "subscription-plans",
    ...lazyRoute(() => import("@/pages/console/subscription-plans")),
  },
  {
    path: "redemption-codes",
    ...lazyRoute(() => import("@/pages/console/redemption-codes")),
  },
  {
    path: "advertisements",
    ...lazyRoute(() => import("@/pages/console/advertisements")),
  },
  {
    path: "usage",
    ...lazyRoute(() => import("@/pages/console/usage")),
  },
  {
    path: "audit-logs",
    ...lazyRoute(() => import("@/pages/console/audit-logs")),
  },
  {
    path: "users",
    ...lazyRoute(() => import("@/pages/console/users")),
  },
  {
    path: "settings",
    ...lazyRoute(() => import("@/pages/console/settings")),
  },
];

export const router = createBrowserRouter([
  {
    element: <InitializationGuard />,
    children: [
      {
        path: "/",
        element: <PublicLayout />,
        errorElement: <ErrorPage />,
        children: [
          ...publicRoutes,
          {
            path: "console",
            element: <ConsoleLayout />,
            children: consoleRoutes,
          },
        ],
      },
      {
        path: "*",
        element: <ErrorPage />,
      },
    ],
  },
]);
