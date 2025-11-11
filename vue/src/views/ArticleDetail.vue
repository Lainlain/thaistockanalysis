<template>
  <div>
    <div class="mb-4">
      <button
        @click="$router.push('/')"
        class="text-blue-600 hover:text-blue-800 flex items-center mb-3 text-sm"
      >
        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back
      </button>
      <h1 class="text-xl font-bold text-gray-900">Article - {{ date }}</h1>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="bg-green-50 border border-green-200 rounded-lg p-3 mb-4 text-sm">
      <p class="text-green-800">{{ successMessage }}</p>
    </div>

    <!-- Error Message -->
    <div v-if="errorMessage" class="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 text-sm">
      <p class="text-red-800">{{ errorMessage }}</p>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600"></div>
      <p class="mt-3 text-sm text-gray-600">Loading article data...</p>
    </div>

    <div v-else class="bg-white rounded-lg shadow p-4">
      <!-- Morning Session -->
      <div class="mb-6">
        <h2 class="text-lg font-bold text-gray-900 mb-3">
          <span class="bg-yellow-100 text-yellow-800 px-3 py-1 rounded-lg text-sm">Morning Session</span>
        </h2>

        <!-- Morning Open -->
        <div class="bg-gray-50 rounded-lg p-3 mb-3">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Opening</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index</label>
              <input
                v-model="morningOpen.index"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1287.01"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change</label>
              <input
                v-model="morningOpen.change"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="4.47"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Highlights</label>
              <textarea
                v-model="morningOpen.highlights"
                rows="3"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="7 => +79, +75, +78 :: 4 => +49, +45"
              ></textarea>
            </div>
          </div>
          <button
            @click="submitMorningOpen"
            :disabled="submitting"
            class="w-full bg-blue-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors disabled:bg-gray-400"
          >
            {{ submitting ? 'Submitting...' : 'Update Morning Open' }}
          </button>
        </div>

        <!-- Morning Close -->
        <div class="bg-gray-50 rounded-lg p-3">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Closing</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index</label>
              <input
                v-model="morningClose.index"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1281.04"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change</label>
              <input
                v-model="morningClose.change"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="-1.50"
              />
            </div>
          </div>
          <button
            @click="submitMorningClose"
            :disabled="submitting"
            class="w-full bg-green-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-green-700 active:bg-green-800 transition-colors disabled:bg-gray-400"
          >
            {{ submitting ? 'Submitting...' : 'Update Morning Close' }}
          </button>
        </div>
      </div>

      <!-- Afternoon Session -->
      <div>
        <h2 class="text-lg font-bold text-gray-900 mb-3">
          <span class="bg-orange-100 text-orange-800 px-3 py-1 rounded-lg text-sm">Afternoon Session</span>
        </h2>

        <!-- Afternoon Open -->
        <div class="bg-gray-50 rounded-lg p-3 mb-3">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Opening</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index</label>
              <input
                v-model="afternoonOpen.index"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1279.48"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change</label>
              <input
                v-model="afternoonOpen.change"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="-8.59"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Highlights</label>
              <textarea
                v-model="afternoonOpen.highlights"
                rows="3"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="7 => +79, +75, +78 :: 4 => +49, +45"
              ></textarea>
            </div>
          </div>
          <button
            @click="submitAfternoonOpen"
            :disabled="submitting"
            class="w-full bg-blue-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors disabled:bg-gray-400"
          >
            {{ submitting ? 'Submitting...' : 'Update Afternoon Open' }}
          </button>
        </div>

        <!-- Afternoon Close -->
        <div class="bg-gray-50 rounded-lg p-3">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Closing</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index</label>
              <input
                v-model="afternoonClose.index"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1275.20"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change</label>
              <input
                v-model="afternoonClose.change"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="-3.28"
              />
            </div>
          </div>
          <button
            @click="submitAfternoonClose"
            :disabled="submitting"
            class="w-full bg-green-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-green-700 active:bg-green-800 transition-colors disabled:bg-gray-400"
          >
            {{ submitting ? 'Submitting...' : 'Update Afternoon Close' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { articleAPI } from '../services/api'

export default {
  name: 'ArticleDetail',
  setup() {
    const route = useRoute()
    const date = ref(route.params.date)
    const submitting = ref(false)
    const successMessage = ref('')
    const errorMessage = ref('')
    const loading = ref(true)

    const morningOpen = ref({ index: '', change: '', highlights: '' })
    const morningClose = ref({ index: '', change: '' })
    const afternoonOpen = ref({ index: '', change: '', highlights: '' })
    const afternoonClose = ref({ index: '', change: '' })

    // Load existing article data
    const loadArticleData = async () => {
      try {
        loading.value = true
        const response = await articleAPI.getArticle(date.value)
        const data = response.data

        // Populate form fields with existing data
        if (data.morning_open) {
          morningOpen.value = {
            index: data.morning_open.index || '',
            change: data.morning_open.change || '',
            highlights: data.morning_open.highlights || ''
          }
        }
        if (data.morning_close) {
          morningClose.value = {
            index: data.morning_close.index || '',
            change: data.morning_close.change || ''
          }
        }
        if (data.afternoon_open) {
          afternoonOpen.value = {
            index: data.afternoon_open.index || '',
            change: data.afternoon_open.change || '',
            highlights: data.afternoon_open.highlights || ''
          }
        }
        if (data.afternoon_close) {
          afternoonClose.value = {
            index: data.afternoon_close.index || '',
            change: data.afternoon_close.change || ''
          }
        }
      } catch (error) {
        console.error('Error loading article data:', error)
        showMessage('Failed to load article data', true)
      } finally {
        loading.value = false
      }
    }

    const showMessage = (message, isError = false) => {
      if (isError) {
        errorMessage.value = message
        successMessage.value = ''
      } else {
        successMessage.value = message
        errorMessage.value = ''
      }
      setTimeout(() => {
        successMessage.value = ''
        errorMessage.value = ''
      }, 5000)
    }

    const submitMorningOpen = async () => {
      try {
        submitting.value = true
        await articleAPI.submitMorningOpen(
          date.value,
          morningOpen.value.index,
          morningOpen.value.change,
          morningOpen.value.highlights
        )
        showMessage('Morning opening data updated successfully!')
      } catch (error) {
        showMessage('Failed to update morning opening data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    const submitMorningClose = async () => {
      try {
        submitting.value = true
        await articleAPI.submitMorningClose(
          date.value,
          morningClose.value.index,
          morningClose.value.change
        )
        showMessage('Morning closing data updated successfully!')
      } catch (error) {
        showMessage('Failed to update morning closing data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    const submitAfternoonOpen = async () => {
      try {
        submitting.value = true
        await articleAPI.submitAfternoonOpen(
          date.value,
          afternoonOpen.value.index,
          afternoonOpen.value.change,
          afternoonOpen.value.highlights
        )
        showMessage('Afternoon opening data updated successfully!')
      } catch (error) {
        showMessage('Failed to update afternoon opening data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    const submitAfternoonClose = async () => {
      try {
        submitting.value = true
        await articleAPI.submitAfternoonClose(
          date.value,
          afternoonClose.value.index,
          afternoonClose.value.change
        )
        showMessage('Afternoon closing data updated successfully!')
      } catch (error) {
        showMessage('Failed to update afternoon closing data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    // Load article data when component mounts
    onMounted(() => {
      loadArticleData()
    })

    return {
      date,
      morningOpen,
      morningClose,
      afternoonOpen,
      afternoonClose,
      submitting,
      successMessage,
      errorMessage,
      loading,
      submitMorningOpen,
      submitMorningClose,
      submitAfternoonOpen,
      submitAfternoonClose
    }
  }
}
</script>
