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
      <h1 class="text-xl font-bold text-gray-900">Create New Article</h1>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="bg-green-50 border border-green-200 rounded-lg p-3 mb-4 text-sm">
      <p class="text-green-800">{{ successMessage }}</p>
    </div>

    <!-- Error Message -->
    <div v-if="errorMessage" class="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 text-sm">
      <p class="text-red-800">{{ errorMessage }}</p>
    </div>

    <div class="bg-white rounded-lg shadow p-4">
      <!-- Date Selection -->
      <div class="mb-4 pb-4 border-b border-gray-200">
        <label class="block text-xs font-medium text-gray-700 mb-2">Article Date</label>
        <input
          v-model="selectedDate"
          type="date"
          class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        <p class="text-xs text-gray-500 mt-2">Today's date is auto-selected</p>
      </div>

      <!-- Morning Session -->
      <div class="mb-5">
        <h2 class="text-base font-bold text-gray-900 mb-3">
          <span class="bg-yellow-100 text-yellow-800 px-2 py-1 rounded-lg text-sm">Morning Session</span>
        </h2>

        <!-- Morning Open -->
        <div class="bg-yellow-50 rounded-lg p-3 mb-3 border border-yellow-200">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Opening</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index *</label>
              <input
                v-model="morningOpen.index"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1287.01"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change *</label>
              <input
                v-model="morningOpen.change"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="4.47"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Highlights *</label>
              <textarea
                v-model="morningOpen.highlights"
                rows="3"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="7 => +79, +75 :: 4 => +49, +45"
              ></textarea>
            </div>
          </div>
          <button
            @click="submitMorningOpen"
            :disabled="submitting || !isValidMorningOpen"
            class="w-full bg-blue-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {{ submitting ? 'Submitting...' : 'Submit Morning Open' }}
          </button>
        </div>

        <!-- Morning Close -->
        <div class="bg-yellow-50 rounded-lg p-3 border border-yellow-200">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Closing</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index *</label>
              <input
                v-model="morningClose.index"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1281.04"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change *</label>
              <input
                v-model="morningClose.change"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="-1.50"
              />
            </div>
          </div>
          <button
            @click="submitMorningClose"
            :disabled="submitting || !isValidMorningClose"
            class="w-full bg-green-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-green-700 active:bg-green-800 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {{ submitting ? 'Submitting...' : 'Submit Morning Close' }}
          </button>
        </div>
      </div>

      <!-- Afternoon Session -->
      <div>
        <h2 class="text-base font-bold text-gray-900 mb-3">
          <span class="bg-orange-100 text-orange-800 px-2 py-1 rounded-lg text-sm">Afternoon Session</span>
        </h2>

        <!-- Afternoon Open -->
        <div class="bg-orange-50 rounded-lg p-3 mb-3 border border-orange-200">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Opening</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index *</label>
              <input
                v-model="afternoonOpen.index"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1279.48"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change *</label>
              <input
                v-model="afternoonOpen.change"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="-8.59"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Highlights *</label>
              <textarea
                v-model="afternoonOpen.highlights"
                rows="3"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="7 => +79, +75 :: 4 => +49, +45"
              ></textarea>
            </div>
          </div>
          <button
            @click="submitAfternoonOpen"
            :disabled="submitting || !isValidAfternoonOpen"
            class="w-full bg-blue-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {{ submitting ? 'Submitting...' : 'Submit Afternoon Open' }}
          </button>
        </div>

        <!-- Afternoon Close -->
        <div class="bg-orange-50 rounded-lg p-3 border border-orange-200">
          <h3 class="text-sm font-semibold text-gray-800 mb-3">Market Closing</h3>
          <div class="space-y-3 mb-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Index *</label>
              <input
                v-model="afternoonClose.index"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1275.20"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Change *</label>
              <input
                v-model="afternoonClose.change"
                type="number"
                step="0.01"
                required
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="-3.28"
              />
            </div>
          </div>
          <button
            @click="submitAfternoonClose"
            :disabled="submitting || !isValidAfternoonClose"
            class="w-full bg-green-600 text-white px-4 py-2 text-sm rounded-lg hover:bg-green-700 active:bg-green-800 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {{ submitting ? 'Submitting...' : 'Submit Afternoon Close' }}
          </button>
        </div>
      </div>

      <!-- Helper Text -->
      <div class="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
        <h4 class="text-xs font-semibold text-blue-900 mb-2">Instructions</h4>
        <ul class="text-xs text-blue-800 space-y-1">
          <li>• Select article date (today pre-selected)</li>
          <li>• Fill each section separately and submit</li>
          <li>• Open sections: Index, Change, Highlights</li>
          <li>• Close sections: Index and Change only</li>
          <li>• Highlights format: "7 => +79, +75 :: 4 => +49, +45"</li>
          <li>• AI analysis generated automatically</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { articleAPI } from '../services/api'

export default {
  name: 'CreateArticle',
  setup() {
    const selectedDate = ref('')
    const submitting = ref(false)
    const successMessage = ref('')
    const errorMessage = ref('')

    const morningOpen = ref({ index: '', change: '', highlights: '' })
    const morningClose = ref({ index: '', change: '' })
    const afternoonOpen = ref({ index: '', change: '', highlights: '' })
    const afternoonClose = ref({ index: '', change: '' })

    // Validation computed properties
    const isValidMorningOpen = computed(() => {
      return morningOpen.value.index && morningOpen.value.change && morningOpen.value.highlights
    })

    const isValidMorningClose = computed(() => {
      return morningClose.value.index && morningClose.value.change
    })

    const isValidAfternoonOpen = computed(() => {
      return afternoonOpen.value.index && afternoonOpen.value.change && afternoonOpen.value.highlights
    })

    const isValidAfternoonClose = computed(() => {
      return afternoonClose.value.index && afternoonClose.value.change
    })

    // Set today's date in YYYY-MM-DD format
    onMounted(() => {
      const today = new Date()
      selectedDate.value = today.toISOString().split('T')[0]
    })

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
      if (!isValidMorningOpen.value) {
        showMessage('Please fill in all morning opening fields', true)
        return
      }

      try {
        submitting.value = true
        await articleAPI.submitMorningOpen(
          selectedDate.value,
          morningOpen.value.index,
          morningOpen.value.change,
          morningOpen.value.highlights
        )
        showMessage('Morning opening data submitted successfully! ✅')
        // Don't clear the form - user might want to reference values
      } catch (error) {
        showMessage('Failed to submit morning opening data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    const submitMorningClose = async () => {
      if (!isValidMorningClose.value) {
        showMessage('Please fill in all morning closing fields', true)
        return
      }

      try {
        submitting.value = true
        await articleAPI.submitMorningClose(
          selectedDate.value,
          morningClose.value.index,
          morningClose.value.change
        )
        showMessage('Morning closing data submitted successfully! ✅')
      } catch (error) {
        showMessage('Failed to submit morning closing data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    const submitAfternoonOpen = async () => {
      if (!isValidAfternoonOpen.value) {
        showMessage('Please fill in all afternoon opening fields', true)
        return
      }

      try {
        submitting.value = true
        await articleAPI.submitAfternoonOpen(
          selectedDate.value,
          afternoonOpen.value.index,
          afternoonOpen.value.change,
          afternoonOpen.value.highlights
        )
        showMessage('Afternoon opening data submitted successfully! ✅')
      } catch (error) {
        showMessage('Failed to submit afternoon opening data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    const submitAfternoonClose = async () => {
      if (!isValidAfternoonClose.value) {
        showMessage('Please fill in all afternoon closing fields', true)
        return
      }

      try {
        submitting.value = true
        await articleAPI.submitAfternoonClose(
          selectedDate.value,
          afternoonClose.value.index,
          afternoonClose.value.change
        )
        showMessage('Afternoon closing data submitted successfully! ✅')
      } catch (error) {
        showMessage('Failed to submit afternoon closing data: ' + error.message, true)
      } finally {
        submitting.value = false
      }
    }

    return {
      selectedDate,
      morningOpen,
      morningClose,
      afternoonOpen,
      afternoonClose,
      submitting,
      successMessage,
      errorMessage,
      isValidMorningOpen,
      isValidMorningClose,
      isValidAfternoonOpen,
      isValidAfternoonClose,
      submitMorningOpen,
      submitMorningClose,
      submitAfternoonOpen,
      submitAfternoonClose
    }
  }
}
</script>
