import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    strictPort: true,
    cors: true,
    proxy: {
      // 将所有/api请求代理到后端服务器
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path,
        configure: (proxy, options) => {
          console.log("初始化代理: /api -> http://localhost:8080");

          proxy.on("error", (err, req, res) => {
            console.error("代理错误", err, req.url);
          });
          proxy.on("proxyReq", (proxyReq, req, res) => {
            console.log("代理请求:", req.method, req.url, "->", proxyReq.path);
          });
          proxy.on("proxyRes", (proxyRes, req, res) => {
            console.log(
              "代理响应:",
              proxyRes.statusCode,
              req.url,
              "头信息:",
              JSON.stringify(proxyRes.headers)
            );
          });
        },
      },
      // 健康检查端点
      "/health": {
        target: "http://localhost:8080",
        changeOrigin: true,
        secure: false,
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  optimizeDeps: {
    include: ["vue", "element-plus"],
  },
  build: {
    sourcemap: true,
    chunkSizeWarningLimit: 1500,
  },
  // 定义全局环境变量
  define: {
    __USE_MOCK__: process.env.VITE_USE_MOCK === "true" || false,
  },
});
