import { createContext, useContext, type PropsWithChildren } from 'react'
import { useGetCurrentAuthenticationSession } from '../../api/generated/endpoints'
import type { UserResource } from '../../api/generated/schemas/userResource.zod'

type AuthContextValue = {
  isPending: boolean
  isAuthenticated: boolean
  user?: UserResource
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: PropsWithChildren) {
  const session = useGetCurrentAuthenticationSession({
    query: { staleTime: 60_000 },
  })

  return (
    <AuthContext.Provider
      value={{
        isPending: session.isPending,
        isAuthenticated: session.data?.data.authenticated === true,
        user: session.data?.data.user,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth must be used inside AuthProvider')
  return value
}
