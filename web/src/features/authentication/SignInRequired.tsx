import { Card, PageIntro } from '../../components/ui'
import { SignInLink } from './SignInLink'

export function SignInRequired({
  title,
  detail,
  returnTo,
}: {
  title: string
  detail: string
  returnTo?: string
}) {
  return (
    <div className="grid max-w-3xl gap-7 max-sm:gap-5">
      <PageIntro compact title={title} />
      <Card className="p-6 max-sm:p-5">
        <p className="max-w-xl text-sm leading-relaxed text-muted">{detail}</p>
        <SignInLink
          className="mt-5 inline-flex rounded-lg bg-brand-teal px-4 py-2.5 text-sm font-bold text-brand-ink no-underline transition hover:bg-brand-teal-light"
          returnTo={returnTo}
        >
          Sign in
        </SignInLink>
      </Card>
    </div>
  )
}
