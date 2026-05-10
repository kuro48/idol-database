import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8081', changeOrigin: true },
    },
  },
  build: {
    outDir: 'dist',
    rollupOptions: {
      output: {
        manualChunks: (id: string) => {
          if (id.includes('recharts')) return 'charts'
          if (id.includes('@tanstack/react-query')) return 'query'
          if (
            id.includes('react-dom') ||
            id.includes('react-router-dom') ||
            id.includes('/react/')
          )
            return 'vendor'
          return undefined
        },
      },
    },
  },
})
