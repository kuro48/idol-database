import createClient from 'openapi-fetch'
import type { paths } from './schema'
import { getValidAuthHeaders } from '../auth/tokenRefresh'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

const baseClient = createClient<paths>({
  baseUrl: API_BASE_URL,
})

// Middleware that injects the OIDC access token from the auth store on every request.
baseClient.use({
  async onRequest({ request }) {
    const { accessToken, idToken } = await getValidAuthHeaders()
    if (accessToken) {
      request.headers.set('Authorization', `Bearer ${accessToken}`)
    }
    if (idToken) {
      request.headers.set('X-ID-Token', idToken)
    }
    return request
  },
})

export const api = baseClient

// Separate admin client that hits /admin/* routes.
const adminBaseClient = createClient<paths>({
  baseUrl: '/',
})

adminBaseClient.use({
  async onRequest({ request }) {
    const { accessToken, idToken } = await getValidAuthHeaders()
    if (accessToken) {
      request.headers.set('Authorization', `Bearer ${accessToken}`)
    }
    if (idToken) {
      request.headers.set('X-ID-Token', idToken)
    }
    return request
  },
})

export const adminApi = adminBaseClient
