import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { SignIn } from './SignIn'

const mocks = vi.hoisted(() => ({ signIn: vi.fn() }))

vi.mock('./AuthContext', () => ({
  useAuth: () => ({ isPending: false, isAuthenticated: false }),
}))

vi.mock('../../api/generated/endpoints', () => ({
  getGetCurrentAuthenticationSessionQueryKey: () => ['/api/authentication-sessions/current'],
  useCreateAuthenticationSession: () => ({
    isPending: false,
    isError: false,
    mutateAsync: mocks.signIn,
  }),
}))

function renderSignIn(entry: string) {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter initialEntries={[entry]}>
        <Routes>
          <Route path="/sign-in" element={<SignIn />} />
          <Route path="/courses/ai-ml" element={<h1>Returned to course</h1>} />
          <Route path="/" element={<h1>Safe home</h1>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

async function completeSignIn() {
  await userEvent.type(screen.getByLabelText('Username'), 'learner')
  await userEvent.type(screen.getByLabelText('Password'), 'secret')
  await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
}

afterEach(cleanup)
beforeEach(() => {
  vi.clearAllMocks()
  mocks.signIn.mockResolvedValue({
    data: {
      authenticated: true,
      user: { id: 'user-1', username: 'learner', displayName: 'Learner' },
    },
    status: 201,
    headers: {},
  })
})

describe('sign-in return destination', () => {
  it('returns to the intended local curriculum route', async () => {
    renderSignIn('/sign-in?returnTo=%2Fcourses%2Fai-ml')
    await completeSignIn()
    expect(await screen.findByRole('heading', { name: 'Returned to course' })).toBeDefined()
  })

  it('rejects an external return URL', async () => {
    renderSignIn('/sign-in?returnTo=https%3A%2F%2Fattacker.example%2Fsteal')
    await completeSignIn()
    expect(await screen.findByRole('heading', { name: 'Safe home' })).toBeDefined()
  })
})
