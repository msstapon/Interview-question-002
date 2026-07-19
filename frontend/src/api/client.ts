const BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1'
const TOKEN_KEY = 'token'

export interface User {
  id: string
  username: string
  created_at: string
}

export interface LoginResponse {
  access_token: string
  expires_in: number
  user: User
}

interface Envelope<T> {
  success: boolean
  data?: T
  error?: { code: string; message: string; details?: unknown }
}

// ApiError carries the backend {code, message} so views can show a useful message.
export class ApiError extends Error {
  status: number
  code: string
  details?: unknown
  constructor(status: number, code: string, message: string, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}
export function setToken(t: string): void {
  localStorage.setItem(TOKEN_KEY, t)
}
export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  let env: Envelope<T> | null = null
  try {
    env = (await res.json()) as Envelope<T>
  } catch {
    // non-JSON response (e.g. network/proxy error)
  }

  if (!res.ok || !env || !env.success) {
    const e = env?.error
    throw new ApiError(
      res.status,
      e?.code ?? 'ERROR',
      e?.message ?? `request failed (${res.status})`,
      e?.details,
    )
  }
  return env.data as T
}

export const api = {
  register: (username: string, password: string, confirmPassword: string) =>
    request<User>('POST', '/auth/register', {
      username,
      password,
      confirm_password: confirmPassword,
    }),
  login: (username: string, password: string) =>
    request<LoginResponse>('POST', '/auth/login', { username, password }),
  me: () => request<User>('GET', '/me'),
}
