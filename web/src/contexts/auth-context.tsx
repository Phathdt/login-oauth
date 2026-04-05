import {
  createContext,
  useCallback,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { apiClient, registerAuthCallbacks } from '@/lib/axios-client'
import { queryClient } from '@/lib/query-client'

export interface User {
  id: string
  email: string
  name: string
  picture?: string
}

interface AuthState {
  user: User | null
  accessToken: string | null
  isLoading: boolean
}

export interface AuthContextValue extends AuthState {
  loginWithFirebase: (idToken: string) => Promise<void>
  updateAccessToken: (token: string) => void
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    accessToken: null,
    isLoading: true,
  })

  const tokenRef = useRef<string | null>(null)

  const setTokens = useCallback((accessToken: string, user: User) => {
    tokenRef.current = accessToken
    setState({ user, accessToken, isLoading: false })
  }, [])

  const updateAccessToken = useCallback((token: string) => {
    tokenRef.current = token
    setState((prev) => ({ ...prev, accessToken: token }))
  }, [])

  const clearAuth = useCallback(() => {
    tokenRef.current = null
    setState({ user: null, accessToken: null, isLoading: false })
  }, [])

  const loginWithFirebase = useCallback(async (idToken: string) => {
    const res = await apiClient.post<{ access_token: string; user: User }>('/auth/firebase', {
      id_token: idToken,
    })
    setTokens(res.data.access_token, res.data.user)
  }, [setTokens])

  const logout = useCallback(async () => {
    try {
      await apiClient.post('/auth/logout')
    } catch {
      // ignore errors — clear auth regardless
    } finally {
      queryClient.clear()
      clearAuth()
    }
  }, [clearAuth])

  useEffect(() => {
    registerAuthCallbacks({
      getAccessToken: () => tokenRef.current,
      updateAccessToken,
      clearAuth,
    })
  }, [updateAccessToken, clearAuth])

  // Restore session from HTTP-only refresh cookie via /auth/me on mount
  useEffect(() => {
    apiClient
      .get<{ user: User; access_token: string }>('/auth/me')
      .then((res) => {
        tokenRef.current = res.data.access_token
        setState({
          user: res.data.user,
          accessToken: res.data.access_token,
          isLoading: false,
        })
      })
      .catch(() => {
        setState((prev) => ({ ...prev, isLoading: false }))
      })
  }, [])

  return (
    <AuthContext.Provider
      value={{ ...state, loginWithFirebase, updateAccessToken, logout }}
    >
      {children}
    </AuthContext.Provider>
  )
}
