import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: { outDir: 'dist', assetsDir: 'assets', emptyOutDir: true },
  server: { proxy: { '/api': 'http://192.168.1.1:8443' } },
})
