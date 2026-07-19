import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { getToken } from '../api/client'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import WelcomeView from '../views/WelcomeView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'login', component: LoginView, meta: { guestOnly: true } }, // IT 02-1
  { path: '/register', name: 'register', component: RegisterView }, // IT 02-2
  { path: '/welcome', name: 'welcome', component: WelcomeView, meta: { requiresAuth: true } }, // IT 02-3
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Guard reads localStorage directly so it never depends on Pinia activation order.
router.beforeEach((to) => {
  const hasToken = !!getToken()
  if (to.meta.requiresAuth && !hasToken) return { path: '/' }
  if (to.meta.guestOnly && hasToken) return { path: '/welcome' }
  return true
})

export default router
