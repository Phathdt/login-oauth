import { useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/hooks/use-auth'
import { signInWithGoogle, signInWithGithub, signInWithEmail, signUpWithEmail } from '@/lib/firebase'

type Mode = 'login' | 'register'

export default function LoginPage() {
  const { user, isLoading, loginWithFirebase } = useAuth()
  const navigate = useNavigate()

  const [mode, setMode] = useState<Mode>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState<'google' | 'github' | 'email' | null>(null)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (user) {
    return <Navigate to="/products" replace />
  }

  async function handleSocialLogin(provider: 'google' | 'github') {
    setError(null)
    setPending(provider)
    try {
      const idToken = provider === 'google' ? await signInWithGoogle() : await signInWithGithub()
      await loginWithFirebase(idToken)
      navigate('/products', { replace: true })
    } catch {
      setError('Sign in failed. Please try again.')
    } finally {
      setPending(null)
    }
  }

  async function handleEmailSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setPending('email')
    try {
      const idToken = mode === 'login'
        ? await signInWithEmail(email, password)
        : await signUpWithEmail(email, password)
      await loginWithFirebase(idToken)
      navigate('/products', { replace: true })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Authentication failed.'
      // Surface friendly Firebase error messages
      if (msg.includes('invalid-credential') || msg.includes('wrong-password')) {
        setError('Invalid email or password.')
      } else if (msg.includes('email-already-in-use')) {
        setError('Email already in use.')
      } else if (msg.includes('weak-password')) {
        setError('Password must be at least 6 characters.')
      } else {
        setError(msg)
      }
    } finally {
      setPending(null)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">
            {mode === 'login' ? 'Welcome back' : 'Create account'}
          </CardTitle>
          <p className="text-muted-foreground text-sm mt-1">
            Sign in to access the store
          </p>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {error && (
            <p className="text-destructive text-sm text-center">{error}</p>
          )}

          {/* Email / Password form */}
          <form onSubmit={handleEmailSubmit} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="flex flex-col gap-1">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full" disabled={pending !== null}>
              {pending === 'email'
                ? 'Please wait...'
                : mode === 'login' ? 'Sign in' : 'Create account'}
            </Button>
          </form>

          <p className="text-center text-sm text-muted-foreground">
            {mode === 'login' ? "Don't have an account? " : 'Already have an account? '}
            <button
              className="underline"
              onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setError(null) }}
            >
              {mode === 'login' ? 'Sign up' : 'Sign in'}
            </button>
          </p>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">or</span>
            </div>
          </div>

          <Button
            variant="outline"
            className="w-full"
            disabled={pending !== null}
            onClick={() => handleSocialLogin('google')}
          >
            {pending === 'google' ? 'Signing in...' : 'Continue with Google'}
          </Button>
          <Button
            variant="outline"
            className="w-full"
            disabled={pending !== null}
            onClick={() => handleSocialLogin('github')}
          >
            {pending === 'github' ? 'Signing in...' : 'Continue with GitHub'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
