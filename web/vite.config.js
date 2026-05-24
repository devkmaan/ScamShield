import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/v1': 'http://localhost:8081',
      '/internal': 'http://localhost:8081',
      '/webhooks': 'http://localhost:8081',
      '/ready': 'http://localhost:8081',
      '/health': 'http://localhost:8081'
    }
  }
});

