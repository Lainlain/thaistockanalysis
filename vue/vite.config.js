import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
	plugins: [vue()],
	server: {
		port: 3000,
		proxy: {
			'/api': {
				target: 'http://localhost:7777',
				changeOrigin: true
			}
		}
	}
})
