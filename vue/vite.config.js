import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
	plugins: [vue()],
	base: '/', // Base path for production assets
	build: {
		outDir: 'dist',
		assetsDir: 'assets',
		// Generate source maps for debugging
		sourcemap: false,
		// Optimize chunks
		rollupOptions: {
			output: {
				manualChunks: {
					'vue-vendor': ['vue', 'vue-router'],
					'axios-vendor': ['axios']
				}
			}
		}
	},
	server: {
		port: 3000,
		proxy: {
			'/api': {
				target: 'http://localhost:7777',
				changeOrigin: true,
				secure: false
			}
		}
	}
})
