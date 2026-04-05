import { Navigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/hooks/use-auth'

const GOOGLE_LOGIN_URL = `${import.meta.env.VITE_API_URL ?? 'http://localhost:3000'}/auth/google/login`

export default function LoginPage() {
  const { user, isLoading } = useAuth()

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

  return (
    <div className="flex items-center justify-center min-h-screen bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome</CardTitle>
          <p className="text-muted-foreground text-sm mt-1">
            Sign in to access the store
          </p>
        </CardHeader>
        <CardContent>
          <Button
            className="w-full"
            onClick={() => {
              window.location.href = GOOGLE_LOGIN_URL
            }}
          >
            Sign in with Google
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
