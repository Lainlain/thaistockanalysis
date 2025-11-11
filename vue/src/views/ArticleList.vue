<template>
  <div>
    <div class="mb-4">
      <h1 class="text-xl font-bold text-gray-900">Article Management</h1>
      <p class="text-sm text-gray-600 mt-1">Manage stock market articles</p>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600"></div>
      <p class="mt-3 text-sm text-gray-600">Loading articles...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-3 text-sm">
      <p class="text-red-800">{{ error }}</p>
    </div>

    <!-- Article List -->
    <div v-else class="bg-white rounded-lg shadow overflow-hidden">
      <div v-if="articles.length === 0" class="text-center py-8">
        <p class="text-sm text-gray-500">No articles found. Create your first article!</p>
      </div>

      <div v-else class="divide-y divide-gray-200">
        <div
          v-for="article in articles"
          :key="article.date"
          @click="viewArticle(article.date)"
          class="p-4 hover:bg-gray-50 active:bg-gray-100 cursor-pointer transition-colors"
        >
          <div class="flex justify-between items-start">
            <div class="flex-1 min-w-0 pr-3">
              <h3 class="text-base font-semibold text-gray-900 truncate">
                {{ article.title }}
              </h3>
              <p class="text-xs text-gray-600 mt-1">{{ article.date }}</p>
              <p class="text-sm text-gray-700 mt-2 line-clamp-2">{{ article.summary }}</p>
            </div>
            <div class="flex-shrink-0">
              <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

export default {
  name: 'ArticleList',
  setup() {
    const router = useRouter()
    const articles = ref([])
    const loading = ref(true)
    const error = ref(null)

    const loadArticles = async () => {
      try {
        loading.value = true
        error.value = null

        // Fetch articles from the API endpoint
        const response = await axios.get('/api/articles')

        articles.value = response.data
      } catch (err) {
        error.value = 'Failed to load articles. Make sure the Go server is running on port 7777.'
        console.error('Error loading articles:', err)
      } finally {
        loading.value = false
      }
    }

    const viewArticle = (date) => {
      router.push(`/article/${date}`)
    }

    onMounted(() => {
      loadArticles()
    })

    return {
      articles,
      loading,
      error,
      viewArticle
    }
  }
}
</script>
