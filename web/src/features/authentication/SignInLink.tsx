import type { PropsWithChildren } from 'react'
import { Link, useLocation } from 'react-router-dom'

export function SignInLink({
  children,
  className,
  returnTo,
}: PropsWithChildren<{ className?: string; returnTo?: string }>) {
  const location = useLocation()
  const destination = returnTo ?? `${location.pathname}${location.search}${location.hash}`

  return (
    <Link className={className} to={`/sign-in?returnTo=${encodeURIComponent(destination)}`}>
      {children}
    </Link>
  )
}
