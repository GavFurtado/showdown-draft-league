import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tailwindcss(),
    react()
  ],
  server: {
    port: 5173, // Ensure this matches your frontend's port
    proxy: {
      '/api': { // Proxy requests starting with /api
        target: 'http://localhost:8080', // Your backend server
        changeOrigin: true, // Needed for virtual hosted sites
        rewrite: (path) => path.replace(/^\/api/, '/api'), // Keep /api prefix for backend
      },
      '/auth': { // Proxy requests starting with /auth (for Discord OAuth)
        target: 'http://localhost:8080', // Your backend server
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/auth/, '/auth'), // Keep /auth prefix for backend
      },
    },
  },
});
