import type { Config } from "tailwindcss";

// Tailwind CSS 4 uses CSS-first configuration via @theme inline in src/index.css.
// This file is kept for tooling compatibility and future content/plugin overrides.
const config: Partial<Config> = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
};

export default config satisfies Config;