import { defineConfig } from 'vite';

export default defineConfig({
  root: '.',
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/tiles': 'http://localhost:8080',
    },
  },
});
