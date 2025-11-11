import axios from 'axios'

const api = axios.create({
	baseURL: '/api',
	headers: {
		'Content-Type': 'application/json'
	}
})

// API service for Thai Stock Market Analysis
export const articleAPI = {
	// Submit morning opening data
	submitMorningOpen(date, index, change, highlights) {
		return api.post('/market-data-analysis', {
			date,
			morning_open: {
				index: parseFloat(index),
				change: parseFloat(change),
				highlights
			}
		})
	},

	// Submit morning closing data
	submitMorningClose(date, index, change) {
		return api.post('/market-data-close', {
			date,
			morning_close: {
				index: parseFloat(index),
				change: parseFloat(change)
			}
		})
	},

	// Submit afternoon opening data
	submitAfternoonOpen(date, index, change, highlights) {
		return api.post('/market-data-analysis', {
			date,
			afternoon_open: {
				index: parseFloat(index),
				change: parseFloat(change),
				highlights
			}
		})
	},

	// Submit afternoon closing data
	submitAfternoonClose(date, index, change) {
		return api.post('/market-data-close', {
			date,
			afternoon_close: {
				index: parseFloat(index),
				change: parseFloat(change)
			}
		})
	},

	// Get all articles
	getArticles() {
		return api.get('/articles')
	},

	// Get article by date (slug) - reads markdown file data
	getArticle(date) {
		return api.get(`/articles/${date}`)
	}
}

export default api
