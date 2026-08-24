import { Badge } from '../../components/ui'
import { SignInLink } from '../authentication/SignInLink'

export function LessonCompletionControl({
  completed,
  pending,
  error,
  onChange,
  authenticated = true,
}: {
  completed: boolean
  pending: boolean
  error?: string
  onChange: (completed: boolean) => void
  authenticated?: boolean
}) {
  if (!authenticated) {
    return (
      <section
        aria-label="Lesson progress"
        className="mt-11 flex items-center justify-between gap-5 border-t border-line pt-7 max-sm:items-start"
      >
        <div>
          <Badge tone="neutral">Progress not saved</Badge>
          <p className="mt-3 text-sm leading-relaxed text-muted">
            Sign in to mark lessons complete and build your learner history.
          </p>
        </div>
        <SignInLink className="shrink-0 rounded-lg border border-line-strong bg-accent-slate/10 px-4 py-2.5 text-sm font-bold text-ink no-underline transition hover:bg-accent-slate/20">
          Sign in to track progress
        </SignInLink>
      </section>
    )
  }

  return (
    <section
      aria-label="Lesson progress"
      className="mt-11 flex items-center justify-between gap-5 border-t border-line pt-7 max-sm:items-start"
    >
      <div>
        <Badge tone={completed ? 'teal' : 'neutral'}>
          {completed ? 'Completed' : 'In progress'}
        </Badge>
        <p className="mt-3 text-sm leading-relaxed text-muted">
          {completed
            ? 'This lesson is recorded as complete.'
            : 'Mark this lesson complete when you have finished the material.'}
        </p>
        {error ? <p className="mt-2 text-sm text-accent-coral">{error}</p> : null}
      </div>
      <button
        type="button"
        disabled={pending}
        onClick={() => onChange(!completed)}
        className="shrink-0 rounded-lg border border-line-strong bg-accent-slate/10 px-4 py-2.5 text-sm font-bold text-ink transition hover:bg-accent-slate/20 disabled:cursor-wait disabled:opacity-60"
      >
        {pending ? 'Saving…' : completed ? 'Mark incomplete' : 'Mark complete'}
      </button>
    </section>
  )
}
