import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => rewriteApiPath(path),
      },
    },
  },
});

// The frontend axios client calls everything under /api/v1, but the Go backend
// serves its domain routes at root (e.g. /goals, /roadmap). Only the ported
// FastAPI intelligence endpoints are registered at /api/v1/*.
const STRIP_PREFIX_ROUTES = [
  'auth', 'domains', 'goals', 'concepts', 'curator', 'notifications',
  'progress', 'competency', 'admin', 'profile', 'resources', 'storage',
  'search', 'telemetry', 'diagnostics', 'assessment', 'roadmap', 'projects',
];

function rewriteApiPath(path) {
  const rest = path.replace(/^\/api\/v1/, '');
  const route = rest.split('/')[1] || '';
  if (STRIP_PREFIX_ROUTES.includes(route)) {
    return rest;
  }
  // AI intelligence endpoints (goal/analyze, learning, mastery, adaptive,
  // guardrails, voice, recommendation, resource, evaluate, health) keep their
  // /api/v1 prefix because the Go backend mirrors FastAPI's exact paths.
  return path;
}
