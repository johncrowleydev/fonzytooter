import { useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import {
  getGetCurrentAuthenticationSessionQueryKey,
  useCreateAuthenticationSession,
} from '../../api/generated/endpoints'
import { Button, Card, PageIntro } from '../../components/ui'
import { useAuth } from './AuthContext'
import { safeReturnTo } from './returnTo'

export function SignIn() {
  const auth = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const signIn = useCreateAuthenticationSession()
  const returnTo = safeReturnTo(searchParams.get('returnTo'))

  if (!auth.isPending && auth.isAuthenticated) {
    return <Navigate replace to={returnTo} />
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    try {
      const session = await signIn.mutateAsync({ data: { username, password } })
      const sessionKey = getGetCurrentAuthenticationSessionQueryKey()
      queryClient.removeQueries({
        predicate: (query) => query.queryKey[0] !== sessionKey[0],
      })
      queryClient.setQueryData(sessionKey, session)
      await navigate(returnTo, { replace: true })
    } catch {
      // The generated mutation exposes the failure state for the inline message below.
    }
  }

  return (
    <div className="mx-auto grid max-w-xl gap-7">
      <PageIntro
        compact
        eyebrow="Learner access"
        title="Sign in"
        detail="Curriculum stays public. Sign in to use your saved progress, reviews, exercises, and tutor."
      />
      <Card className="p-6 max-sm:p-5">
        <form className="grid gap-4" onSubmit={submit}>
          <label className="grid gap-2 text-sm font-semibold">
            Username
            <input
              autoComplete="username"
              className="rounded-lg border border-line-strong bg-panel-soft px-3 py-2.5 text-ink"
              onChange={(event) => setUsername(event.target.value)}
              required
              value={username}
            />
          </label>
          <label className="grid gap-2 text-sm font-semibold">
            Password
            <input
              autoComplete="current-password"
              className="rounded-lg border border-line-strong bg-panel-soft px-3 py-2.5 text-ink"
              onChange={(event) => setPassword(event.target.value)}
              required
              type="password"
              value={password}
            />
          </label>
          {signIn.isError ? (
            <p role="alert" className="text-sm text-accent-coral">
              That username and password were not accepted.
            </p>
          ) : null}
          <Button disabled={signIn.isPending || !username || !password} type="submit">
            {signIn.isPending ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>
      </Card>
    </div>
  )
}
