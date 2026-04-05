import { useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/hooks/use-auth'
import {
  signInWithGoogle,
  signInWithGithub,
  signInWithEmail,
  signUpWithEmail,
  sendPhoneOTP,
  confirmPhoneOTP,
  type ConfirmationResult,
} from '@/lib/firebase'

type EmailMode = 'login' | 'register'
type AuthTab = 'email' | 'phone'

const RECAPTCHA_CONTAINER_ID = 'recaptcha-container'

export default function LoginPage() {
  const { user, isLoading, loginWithFirebase } = useAuth()
  const navigate = useNavigate()

  const [tab, setTab] = useState<AuthTab>('email')
  const [emailMode, setEmailMode] = useState<EmailMode>('login')

  // Email/password state
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  // Phone state
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [confirmation, setConfirmation] = useState<ConfirmationResult | null>(null)

  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState<'google' | 'github' | 'email' | 'phone-send' | 'phone-verify' | null>(null)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (user) return <Navigate to="/products" replace />

  function resetError() { setError(null) }

  async function handleSocialLogin(provider: 'google' | 'github') {
    resetError()
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
    resetError()
    setPending('email')
    try {
      const idToken = emailMode === 'login'
        ? await signInWithEmail(email, password)
        : await signUpWithEmail(email, password)
      await loginWithFirebase(idToken)
      navigate('/products', { replace: true })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : ''
      if (msg.includes('invalid-credential') || msg.includes('wrong-password')) setError('Invalid email or password.')
      else if (msg.includes('email-already-in-use')) setError('Email already in use.')
      else if (msg.includes('weak-password')) setError('Password must be at least 6 characters.')
      else setError('Authentication failed.')
    } finally {
      setPending(null)
    }
  }

  // Step 1: send OTP
  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault()
    resetError()
    setPending('phone-send')
    try {
      const result = await sendPhoneOTP(phone, RECAPTCHA_CONTAINER_ID)
      setConfirmation(result)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : ''
      if (msg.includes('invalid-phone-number')) setError('Invalid phone number. Use international format e.g. +84912345678')
      else if (msg.includes('too-many-requests')) setError('Too many attempts. Please try again later.')
      else setError('Failed to send OTP. Please try again.')
    } finally {
      setPending(null)
    }
  }

  // Step 2: verify OTP
  async function handleVerifyOTP(e: React.FormEvent) {
    e.preventDefault()
    if (!confirmation) return
    resetError()
    setPending('phone-verify')
    try {
      const idToken = await confirmPhoneOTP(confirmation, otp)
      await loginWithFirebase(idToken)
      navigate('/products', { replace: true })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : ''
      if (msg.includes('invalid-verification-code')) setError('Incorrect code. Please try again.')
      else if (msg.includes('code-expired')) setError('Code expired. Please request a new one.')
      else setError('Verification failed.')
    } finally {
      setPending(null)
    }
  }

  const isBusy = pending !== null

  return (
    <div className="flex items-center justify-center min-h-screen bg-background px-4">
      {/* Invisible reCAPTCHA anchor — must be in the DOM */}
      <div id={RECAPTCHA_CONTAINER_ID} />

      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome</CardTitle>
          <p className="text-muted-foreground text-sm mt-1">Sign in to access the store</p>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {error && <p className="text-destructive text-sm text-center">{error}</p>}

          {/* Tab switcher */}
          <div className="flex rounded-md border overflow-hidden text-sm">
            <button
              className={`flex-1 py-1.5 transition-colors ${tab === 'email' ? 'bg-primary text-primary-foreground' : 'bg-background text-muted-foreground hover:bg-muted'}`}
              onClick={() => { setTab('email'); resetError() }}
            >
              Email
            </button>
            <button
              className={`flex-1 py-1.5 transition-colors ${tab === 'phone' ? 'bg-primary text-primary-foreground' : 'bg-background text-muted-foreground hover:bg-muted'}`}
              onClick={() => { setTab('phone'); setConfirmation(null); setOtp(''); resetError() }}
            >
              Phone
            </button>
          </div>

          {/* Email/password form */}
          {tab === 'email' && (
            <>
              <form onSubmit={handleEmailSubmit} className="flex flex-col gap-3">
                <div className="flex flex-col gap-1">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" placeholder="you@example.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
                </div>
                <div className="flex flex-col gap-1">
                  <Label htmlFor="password">Password</Label>
                  <Input id="password" type="password" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} required />
                </div>
                <Button type="submit" className="w-full" disabled={isBusy}>
                  {pending === 'email' ? 'Please wait...' : emailMode === 'login' ? 'Sign in' : 'Create account'}
                </Button>
              </form>
              <p className="text-center text-sm text-muted-foreground">
                {emailMode === 'login' ? "Don't have an account? " : 'Already have an account? '}
                <button className="underline" onClick={() => { setEmailMode(emailMode === 'login' ? 'register' : 'login'); resetError() }}>
                  {emailMode === 'login' ? 'Sign up' : 'Sign in'}
                </button>
              </p>
            </>
          )}

          {/* Phone auth form */}
          {tab === 'phone' && (
            <>
              {!confirmation ? (
                // Step 1: enter phone number
                <form onSubmit={handleSendOTP} className="flex flex-col gap-3">
                  <div className="flex flex-col gap-1">
                    <Label htmlFor="phone">Phone number</Label>
                    <Input
                      id="phone"
                      type="tel"
                      placeholder="+84912345678"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      required
                    />
                    <p className="text-xs text-muted-foreground">Include country code e.g. +84</p>
                  </div>
                  <Button type="submit" className="w-full" disabled={isBusy}>
                    {pending === 'phone-send' ? 'Sending...' : 'Send OTP'}
                  </Button>
                </form>
              ) : (
                // Step 2: enter OTP
                <form onSubmit={handleVerifyOTP} className="flex flex-col gap-3">
                  <div className="flex flex-col gap-1">
                    <Label htmlFor="otp">Verification code</Label>
                    <Input
                      id="otp"
                      type="text"
                      inputMode="numeric"
                      placeholder="123456"
                      maxLength={6}
                      value={otp}
                      onChange={(e) => setOtp(e.target.value)}
                      required
                    />
                    <p className="text-xs text-muted-foreground">Enter the 6-digit code sent to {phone}</p>
                  </div>
                  <Button type="submit" className="w-full" disabled={isBusy}>
                    {pending === 'phone-verify' ? 'Verifying...' : 'Verify'}
                  </Button>
                  <button
                    type="button"
                    className="text-sm text-muted-foreground underline text-center"
                    onClick={() => { setConfirmation(null); setOtp(''); resetError() }}
                  >
                    Change number
                  </button>
                </form>
              )}
            </>
          )}

          {/* Social divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center"><span className="w-full border-t" /></div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">or</span>
            </div>
          </div>

          <Button variant="outline" className="w-full" disabled={isBusy} onClick={() => handleSocialLogin('google')}>
            {pending === 'google' ? 'Signing in...' : 'Continue with Google'}
          </Button>
          <Button variant="outline" className="w-full" disabled={isBusy} onClick={() => handleSocialLogin('github')}>
            {pending === 'github' ? 'Signing in...' : 'Continue with GitHub'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
