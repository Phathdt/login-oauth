import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { apiClient } from '@/lib/axios-client'
import { useAuth } from '@/hooks/use-auth'
import type { User } from '@/contexts/auth-context'

export default function AuthCallback() {
  const { setTokens } = useAuth()
  const navigate = useNavigate()
  const called = useRef(false)

  useEffect(() => {
    // Prevent double-execution in React StrictMode
    if (called.current) return
    called.current = true

    const params = new URLSearchParams(window.location.search)
    const accessToken = params.get('access_token')

    if (!accessToken) {
      navigate('/login', { replace: true })
      return
    }

    // Fetch user info using the new access token
    apiClient
      .get<{ user: User; access_token: string }>('/auth/me', {
        headers: { Authorization: `Bearer ${accessToken}` },
      })
      .then((res) => {
        setTokens(res.data.access_token, res.data.user)
        navigate('/products', { replace: true })
      })
      .catch(() => {
        navigate('/login', { replace: true })
      })
  }, [setTokens, navigate])

  return (
    <div className="flex items-center justify-center min-h-screen">
      <p className="text-muted-foreground">Completing sign in...</p>
    </div>
  )
}
