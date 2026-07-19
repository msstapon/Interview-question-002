<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import FormFrame from '../components/FormFrame.vue'
import { useAuth } from '../stores/auth'

const router = useRouter()
const auth = useAuth()
const username = ref(auth.username ?? '')

onMounted(async () => {
  // Validate the JWT against the backend; on failure, drop the token and go to login.
  try {
    const u = await auth.fetchMe()
    username.value = u.username
  } catch {
    auth.logout()
    router.replace('/')
  }
})

function onLogout(): void {
  auth.logout()
  router.replace('/')
}
</script>

<template>
  <div class="page">
    <FormFrame title="IT 02-3">
      <div class="welcome">Welcome&nbsp;&nbsp;User : {{ username }}</div>
      <div class="link" style="text-align: center; margin-top: 8px">
        <a @click.prevent="onLogout">ออกจากระบบ</a>
      </div>
    </FormFrame>
  </div>
</template>
