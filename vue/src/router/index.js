import { createRouter, createWebHistory } from 'vue-router'
import ArticleDetail from '../views/ArticleDetail.vue'
import ArticleList from '../views/ArticleList.vue'
import CreateArticle from '../views/CreateArticle.vue'

const routes = [
	{
		path: '/',
		name: 'ArticleList',
		component: ArticleList
	},
	{
		path: '/article/:date',
		name: 'ArticleDetail',
		component: ArticleDetail,
		props: true
	},
	{
		path: '/create',
		name: 'CreateArticle',
		component: CreateArticle
	}
]

const router = createRouter({
	history: createWebHistory(),
	routes
})

export default router
