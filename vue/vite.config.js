import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
	plugins: [vue()],
	server: {
		port: 3000,
		proxy: {
			'/api': {
				target: process.env.VITE_API_URL || 'https://thaistockanalysis.com',
				changeOrigin: true,
				secure: true
			}
		}
	}
})
