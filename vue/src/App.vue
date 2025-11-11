<template>
  <div class="min-h-screen bg-gray-50">
    <!-- App Bar - Mobile Optimized -->
    <nav class="bg-blue-600 shadow-lg sticky top-0 z-50">
      <div class="px-3 py-3">
        <div class="flex justify-between items-center">
          <div class="flex items-center flex-1 min-w-0">
            <router-link to="/" class="text-white text-base font-bold truncate">
              Thai Stock Analysis Admin
            </router-link>
          </div>
          <div class="ml-3">
            <router-link
              to="/create"
              class="bg-white text-blue-600 px-3 py-2 rounded-lg text-sm font-semibold hover:bg-blue-50 transition-colors whitespace-nowrap"
            >
              + New
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content - Mobile Optimized -->
    <main class="px-3 py-4">
      <router-view />
    </main>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>
