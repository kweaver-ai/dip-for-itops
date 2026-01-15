import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import dts from 'vite-plugin-dts';
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    dts({
      outDir: 'dist',
      tsconfigPath: './tsconfig.json'
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    lib: {
      entry: './src/index.ts',          // 库入口文件
      name: 'chatkit',                  // 全局变量名（UMD/IIFE 模式）
      fileName: (format) => `chatkit.${format}.js`, // 输出文件名
      formats: ['es', 'cjs', 'umd'],     // 打包格式（ES模块、CommonJS、UMD）
    },
    outDir: './dist',
  },
  server: {
    proxy: {
      // 将本地 /data-agent 前缀的请求代理到 Data Agent 服务,解决本地开发时的 CORS
      '/data-agent': {
        target: 'https://dip.aishu.cn:443/api/agent-app/v1',
        changeOrigin: true,
        secure: false,
        rewrite: (p) => p.replace(/^\/data-agent/, ''),
      },
      // 将本地 /api 前缀的请求代理到 AISHU 服务,用于 agent-factory API
      '/api': {
        target: 'https://dip.aishu.cn:443',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
