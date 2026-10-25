import { fileURLToPath, URL } from 'node:url'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const toolRoot = fileURLToPath(new URL('.', import.meta.url))
const assetsRoot = fileURLToPath(new URL('../../../assets', import.meta.url))

export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    fs: {
      allow: [toolRoot, assetsRoot],
    },
  },
})
