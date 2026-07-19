<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { ApiError } from '../api/client'
import FormFrame from '../components/FormFrame.vue'
import { useAuth } from '../stores/auth'

const router = useRouter()
const auth = useAuth()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

async function onSubmit(): Promise<void> {
  error.value = ''
  // Client-side guard: Password must equal Confirm Password (also enforced server-side).
  if (password.value !== confirmPassword.value) {
    error.value = 'Password และ Confirm Password ไม่ตรงกัน'
    return
  }
  loading.value = true
  try {
    await auth.register(username.value.trim(), password.value, confirmPassword.value)
    // Success → go back to IT 02-1 to log in.
    router.push({ path: '/', query: { registered: '1' } })
  } catch (e) {
    if (e instanceof ApiError && e.code === 'USERNAME_TAKEN') {
      error.value = 'ชื่อผู้ใช้นี้ถูกใช้แล้ว'
    } else if (e instanceof ApiError) {
      error.value = e.message
    } else {
      error.value = 'สมัครสมาชิกไม่สำเร็จ'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <FormFrame title="IT 02-2">
      <form class="form" @submit.prevent="onSubmit">
        <div class="row">
          <label for="username">User</label>
          <input id="username" v-model="username" type="text" autocomplete="username" />
        </div>
        <div class="row">
          <label for="password">Password</label>
          <input id="password" v-model="password" type="password" autocomplete="new-password" />
        </div>
        <div class="row">
          <label for="confirm">Confirm Password</label>
          <input id="confirm" v-model="confirmPassword" type="password" autocomplete="new-password" />
        </div>
        <p class="msg msg--error">{{ error }}</p>
        <div class="actions">
          <button type="submit" :disabled="loading">สมัครสมาชิก</button>
        </div>
      </form>
    </FormFrame>
  </div>
</template>
