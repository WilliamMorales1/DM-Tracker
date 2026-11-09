import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const backend = "http://localhost:8080";

export default defineConfig({
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  server: {
    proxy: {
      "/api": backend,
      "/images": backend,
      "/maps": backend,
    },
  },
  build: { outDir: "dist", sourcemap: true },
  test: { include: ["src/**/*.test.ts"] },
});
