import axios from 'axios'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:3000',
  withCredentials: true,
})

// Auth state callbacks registered by AuthProvider
let getAccessToken: () => string | null = () => null
let updateAccessToken: (token: string) => void = () => {}
let clearAuth: () => void = () => {}

export function registerAuthCallbacks(callbacks: {
  getAccessToken: () => string | null
  updateAccessToken: (token: string) => void
  clearAuth: () => void
}) {
  getAccessToken = callbacks.getAccessToken
  updateAccessToken = callbacks.updateAccessToken
  clearAuth = callbacks.clearAuth
}

// Queue for concurrent 401 requests while refreshing
let isRefreshing = false
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (err: unknown) => void
}> = []

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token!)
    }
  })
  failedQueue = []
}

// Attach access token to every request
apiClient.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// On 401: refresh token then retry, queue concurrent 401s
apiClient.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config as typeof error.config & { _retry?: boolean }

    // Never retry the refresh endpoint itself — avoids deadlock
    if (
      error.response?.status !== 401 ||
      original._retry ||
      original.url === '/auth/refresh'
    ) {
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        original.headers.Authorization = `Bearer ${token}`
        return apiClient(original)
      })
    }

    original._retry = true
    isRefreshing = true

    try {
      const { data } = await apiClient.post<{ access_token: string }>(
        '/auth/refresh'
      )
      updateAccessToken(data.access_token)
      processQueue(null, data.access_token)
      original.headers.Authorization = `Bearer ${data.access_token}`
      return apiClient(original)
    } catch (err) {
      processQueue(err, null)
      clearAuth()
      return Promise.reject(err)
    } finally {
      isRefreshing = false
    }
  }
)
