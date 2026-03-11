import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    plugins: [vue()],
    server: {
      port: 3000,
      host: true,
      strictPort: true,
      watch: {
        usePolling: true
      },
      proxy: {
        '/api': {
          target: 'http://localhost',
          changeOrigin: true
        }
      }
    }
  }
})
