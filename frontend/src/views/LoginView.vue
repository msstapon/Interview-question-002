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
const error = ref('')
const loading = ref(false)

async function onSubmit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    router.push('/welcome')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'เข้าสู่ระบบไม่สำเร็จ'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <FormFrame title="IT 02-1">
      <form class="form" @submit.prevent="onSubmit">
        <div class="row">
          <label for="username">User</label>
          <input id="username" v-model="username" type="text" autocomplete="username" />
        </div>
        <div class="row">
          <label for="password">Password</label>
          <input id="password" v-model="password" type="password" autocomplete="current-password" />
        </div>
        <p class="msg msg--error">{{ error }}</p>
        <div class="actions">
          <button type="submit" :disabled="loading">ลงชื่อเข้าใช้งาน</button>
        </div>
        <div class="link">
          <router-link to="/register">สมัครสมาชิก</router-link>
        </div>
      </form>
    </FormFrame>
  </div>
</template>
