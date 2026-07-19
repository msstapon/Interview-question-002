import { defineStore } from 'pinia'

import { api, clearToken, getToken, setToken, type User } from '../api/client'

interface State {
  token: string | null
  username: string | null
}

export const useAuth = defineStore('auth', {
  state: (): State => ({
    token: getToken(),
    username: localStorage.getItem('username'),
  }),
  getters: {
    isAuthenticated: (s): boolean => !!s.token,
  },
  actions: {
    async register(username: string, password: string, confirmPassword: string): Promise<void> {
      await api.register(username, password, confirmPassword)
    },
    async login(username: string, password: string): Promise<void> {
      const res = await api.login(username, password)
      this.token = res.access_token
      this.username = res.user.username
      setToken(res.access_token)
      localStorage.setItem('username', res.user.username)
    },
    async fetchMe(): Promise<User> {
      const u = await api.me()
      this.username = u.username
      localStorage.setItem('username', u.username)
      return u
    },
    logout(): void {
      this.token = null
      this.username = null
      clearToken()
      localStorage.removeItem('username')
    },
  },
})
