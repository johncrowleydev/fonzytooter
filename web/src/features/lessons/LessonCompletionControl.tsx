import { Badge } from '../../components/ui'

export function LessonCompletionControl({
  completed,
  pending,
  error,
  onChange,
}: {
  completed: boolean
  pending: boolean
  error?: string
  onChange: (completed: boolean) => void
}) {
  return (
    <section
      aria-label="Lesson progress"
      className="mt-11 flex items-center justify-between gap-5 border-t border-line pt-7 max-sm:items-start"
    >
      <div>
        <Badge tone={completed ? 'teal' : 'neutral'}>
          {completed ? 'Completed' : 'In progress'}
        </Badge>
        <p className="mt-3 text-xs leading-relaxed text-muted">
          {completed
            ? 'This lesson is recorded as complete.'
            : 'Mark this lesson complete when you have finished the material.'}
        </p>
        {error ? <p className="mt-2 text-2xs text-accent-coral">{error}</p> : null}
      </div>
      <button
        type="button"
        disabled={pending}
        onClick={() => onChange(!completed)}
        className="shrink-0 rounded-lg border border-line-strong bg-accent-slate/10 px-4 py-2.5 text-xs font-bold text-ink transition hover:bg-accent-slate/20 disabled:cursor-wait disabled:opacity-60"
      >
        {pending ? 'Saving…' : completed ? 'Mark incomplete' : 'Mark complete'}
      </button>
    </section>
  )
}
