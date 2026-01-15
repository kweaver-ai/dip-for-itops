import { defineConfig, legacy, Navigate } from '@noya/max';
import setupProxy from './src/setupProxy';

const DEBUG_ORIGIN = 'https://192.168.201.15';
const MOCK_ORIGIN = 'http://localhost:56721';
const ORIGIN = 'http://localhost:5672';

const publicPath =
  process.env.NODE_ENV === 'development' ? '/' : '/alert-fault/';

export default defineConfig({
  mountElementId: 'aiAlertFaultRoot',
  base: '/',
  history: { type: 'browser' },
  publicPath,
  cssPublicPath: publicPath,
  server: {
    proxy: {
      '/api': {
        target: DEBUG_ORIGIN,
        changeOrigin: true,
        secure: false
      },
      '/manager': {
        target: DEBUG_ORIGIN,
        changeOrigin: true,
        secure: false
      },
      '/mock': {
        target: MOCK_ORIGIN,
        changeOrigin: true,
        secure: false,
        pathRewrite: {
          '^/mock': ''
        }
      }
    },
    onBeforeSetupMiddleware: ({ app }) => {
      if (app) {
        setupProxy({ app, DEBUG_ORIGIN, ORIGIN });
      }
    }
  },
  routes: [
    {
      path: '/',
      element: '@/pages/Home.tsx'
    },
    {
      path: '/fault-analysis',
      element: '@/pages/AlertFault/index.tsx'
    },
    {
      path: '/fault-analysis/:id',
      element: '@/pages/FaultAnalysis/index.tsx'
    }
  ],
  plugins: [
    legacy({
      qiankun: {
        base: {
          apps: []
        },
        sub: {}
      }
    })
  ]
});
