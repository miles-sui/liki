import { ofetch } from 'ofetch'

const API_BASE = '/api'

export function getAuthToken(): string | null {
  return localStorage.getItem('token')
}

export const api = ofetch.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
  onRequest({ options }) {
    const token = getAuthToken()
    if (token) {
      const h = new Headers(options.headers as HeadersInit)
      h.set('Authorization', `Bearer ${token}`)
      options.headers = h
    }
  },
  onResponseError({ response }) {
    const body = response._data
    const message = body?.error?.message || body?.message || `HTTP ${response.status}`
    throw new Error(message)
  },
})

export function setAuthToken(token: string | null) {
  if (token) {
    localStorage.setItem('token', token)
  } else {
    localStorage.removeItem('token')
  }
}
