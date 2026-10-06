import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      "/api/": {
        target: "http://localhost:8080",
        changeOrigin: true,
        // 转发 X-Forwarded-For，支付下单需要真实客户端 IP（易支付 clientip）
        xfwd: true,
      },
    },
  },
  build: {
    outDir: "dist",
    sourcemap: true,
  },
});