import { useEffect, useMemo, useState } from 'react'
import { useGetCourse, useGetCourseProgress } from '../../api/generated/endpoints'
import type { CourseProgressResource } from '../../api/generated/schemas/courseProgressResource.zod'
import type { ObjectiveProgressResource } from '../../api/generated/schemas/objectiveProgressResource.zod'
import { DEFAULT_COURSE_ID } from '../../app/routes'
import { Badge, Card, PageIntro, SectionHeading, StatusDot } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

const summaryToneStyles = {
  teal: 'border-t-accent-teal',
  gold: 'border-t-accent-gold',
  coral: 'border-t-accent-coral',
  violet: 'border-t-accent-violet',
} as const

// Accent rather than brand: these are bare meter bars with no text, so they need to contrast with
// the card behind them. See the note on progressTones in components/ui.tsx.
const dimensionToneStyles = {
  introduced: 'bg-accent-teal',
  pending: 'bg-accent-gold',
  evidence: 'bg-accent-coral',
  'not-assessed': 'bg-accent-slate/20',
} as const

type EvidenceFilter = 'all' | 'needs-review' | 'application' | 'not-introduced'

export function Progress() {
  const { setPageContext } = useTutor()
  const progressQuery = useGetCourseProgress(DEFAULT_COURSE_ID)
  // Objective progress carries a moduleId but no module title, so the names come from the course.
  const courseQuery = useGetCourse(DEFAULT_COURSE_ID)
  const [selectedId, setSelectedId] = useState<string>()
  const progress = progressQuery.data?.data
  const moduleTitles = useMemo(
    () =>
      new Map((courseQuery.data?.data.modules ?? []).map((module) => [module.id, module.title])),
    [courseQuery.data],
  )
  const selected =
    progress?.objectives.find((objective) => objective.id === selectedId) ?? progress?.objectives[0]

  useEffect(() => {
    setPageContext({
      type: 'progress',
      title: 'Progress',
      courseId: DEFAULT_COURSE_ID,
      objectiveId: selected?.id,
      objectiveTitle: selected?.title,
    })
  }, [selected, setPageContext])

  if (progressQuery.isPending) {
    return <ProgressState title="Loading progress" detail="Reading your saved lesson progress…" />
  }
  if (progressQuery.isError || !progress) {
    return (
      <ProgressState
        title="Progress unavailable"
        detail="Your saved progress could not be loaded. Try again."
      />
    )
  }

  return (
    <ProgressView
      progress={progress}
      selectedId={selected?.id}
      onSelect={setSelectedId}
      moduleTitles={moduleTitles}
    />
  )
}

export function ProgressView({
  progress,
  selectedId,
  onSelect,
  moduleTitles,
}: {
  progress: CourseProgressResource
  selectedId?: string
  onSelect: (objectiveId: string) => void
  /** Module id to title. Optional because the course may still be loading. */
  moduleTitles?: ReadonlyMap<string, string>
}) {
  const introducedCount = progress.objectives.filter((objective) => objective.introduced).length
  const [filter, setFilter] = useState<EvidenceFilter>('all')
  const filteredObjectives = progress.objectives.filter((objective) => {
    if (filter === 'needs-review') return objective.recall.dueReviewCount > 0
    if (filter === 'application') return objective.application.attempts > 0
    if (filter === 'not-introduced') return !objective.introduced
    return true
  })
  const selected =
    filteredObjectives.find((objective) => objective.id === selectedId) ?? filteredObjectives[0]
  const reviewedCount = progress.objectives.filter(
    (objective) => objective.recall.reviewsCompleted > 0,
  ).length
  const applicationCount = progress.objectives.filter(
    (objective) => objective.application.attempts > 0,
  ).length

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro
        compact
        title="Progress"
        detail={`${progress.completedLessonCount} of ${progress.totalLessonCount} lessons complete`}
      />
      <section className="grid grid-cols-4 gap-2.5 max-lg:grid-cols-2">
        <SummaryStat
          value={String(introducedCount)}
          label="Introduced"
          note="objectives encountered"
          tone="teal"
        />
        <SummaryStat
          value={String(progress.objectives.length - introducedCount)}
          label="Not introduced"
          note="still ahead"
          tone="gold"
        />
        <SummaryStat
          value={String(reviewedCount)}
          label="Reviewed"
          note="objectives with recall evidence"
          tone="coral"
        />
        <SummaryStat
          value={String(applicationCount)}
          label="Practiced"
          note="objectives with checked attempts"
          tone="violet"
        />
      </section>
      <div className="flex flex-wrap gap-2" aria-label="Objective evidence filters">
        {(
          [
            ['all', 'All'],
            ['needs-review', 'Needs review'],
            ['application', 'Has application evidence'],
            ['not-introduced', 'Not introduced'],
          ] as const
        ).map(([value, label]) => (
          <button
            type="button"
            key={value}
            onClick={() => setFilter(value)}
            className={`rounded-full border px-3 py-2 text-sm font-semibold pointer-coarse:min-h-11 pointer-coarse:px-4 ${filter === value ? 'border-accent-teal bg-accent-teal/10 text-ink' : 'border-line bg-panel text-muted'}`}
          >
            {label}
          </button>
        ))}
      </div>
      {selected ? (
        <div className="grid grid-cols-[1.15fr_0.85fr] gap-3.5 max-lg:grid-cols-1 max-sm:gap-2.5">
          <Card className="min-h-128">
            <SectionHeading
              title="Objectives"
              detail={`${filteredObjectives.length} objectives match this evidence filter.`}
            />
            <div className="grid">
              {filteredObjectives.map((objective) => (
                <ObjectiveBrowserRow
                  key={objective.id}
                  objective={objective}
                  moduleTitle={moduleTitles?.get(objective.moduleId)}
                  selected={selected.id === objective.id}
                  onClick={() => onSelect(objective.id)}
                />
              ))}
            </div>
          </Card>
          <ObjectiveDetail objective={selected} />
        </div>
      ) : (
        <Card muted>
          <h2 className="text-base tracking-tight text-ink">No objectives available</h2>
          <p className="mt-2 text-sm leading-relaxed text-muted">
            This course does not have authored learning objectives yet.
          </p>
        </Card>
      )}
    </div>
  )
}

function ObjectiveDetail({ objective }: { objective: ObjectiveProgressResource }) {
  return (
    <Card className="min-h-128 p-6 max-lg:min-h-0 max-sm:p-5">
      <Badge tone={objective.introduced ? 'teal' : 'neutral'}>
        {objective.introduced ? 'Introduced' : 'Not introduced'}
      </Badge>
      <h2 className="my-5 max-w-sm text-3xl leading-tight tracking-tight">{objective.title}</h2>
      <p className="text-sm leading-relaxed text-muted">{objective.description}</p>
      <div className="mt-8 grid gap-0">
        <Dimension
          label="Introduced"
          value={objective.introduced ? 'Yes' : 'Not yet'}
          state={objective.introduced ? 'introduced' : 'pending'}
        />
        <Dimension
          label="Recall"
          value={
            objective.recall.reviewItemCount === 0
              ? 'No authored items'
              : `${objective.recall.reviewsCompleted} reviews · ${objective.recall.dueReviewCount} due`
          }
          state={objective.recall.reviewsCompleted > 0 ? 'evidence' : 'not-assessed'}
        />
        <Dimension
          label="Application"
          value={
            objective.application.exerciseCount === 0
              ? 'No authored exercises'
              : `${objective.application.attempts} attempts · ${objective.application.fullyPassedAttempts} passed`
          }
          state={objective.application.attempts > 0 ? 'evidence' : 'not-assessed'}
        />
        <Dimension label="Transfer" value="Not assessed" state="not-assessed" />
      </div>
      <div className="mt-6 grid gap-2 text-sm text-muted">
        <p>
          {objective.completedLessonCount} of {objective.linkedLessonCount} linked lessons
          completed.
        </p>
        {objective.recall.lastReviewedAt ? (
          <p>Last reviewed {formatEvidenceTime(objective.recall.lastReviewedAt)}.</p>
        ) : null}
        {objective.recall.nextDueAt ? (
          <p>Next due {formatEvidenceTime(objective.recall.nextDueAt)}.</p>
        ) : null}
        {objective.application.lastCheckedAt ? (
          <p>Last exercise check {formatEvidenceTime(objective.application.lastCheckedAt)}.</p>
        ) : null}
      </div>
      <p className="mt-8 rounded-lg border border-line bg-accent-slate/10 p-4 text-sm leading-relaxed text-muted">
        These are counts and timestamps from real learner actions. They are evidence, not a mastery
        score. Transfer remains {objective.transferAssessed ? 'assessed' : 'not assessed'}.
      </p>
    </Card>
  )
}

function SummaryStat({
  value,
  label,
  note,
  tone,
}: {
  value: string
  label: string
  note: string
  tone: keyof typeof summaryToneStyles
}) {
  return (
    <Card className={`grid gap-1 border-t-2 border-t-transparent p-5 ${summaryToneStyles[tone]}`}>
      <span className="text-4xl tracking-tight">{value}</span>
      <strong className="text-sm">{label}</strong>
      <span className="text-sm text-faint">{note}</span>
    </Card>
  )
}

function ObjectiveBrowserRow({
  objective,
  moduleTitle,
  selected,
  onClick,
}: {
  objective: ObjectiveProgressResource
  moduleTitle?: string
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`grid w-full grid-cols-[11px_minmax(0,1fr)_auto_15px] items-center gap-2.5 border-0 border-b border-line bg-transparent py-3 text-left text-ink max-sm:flex max-sm:min-w-0 ${selected ? 'border-l-2 border-accent-teal bg-accent-teal/5 pl-2' : 'hover:bg-accent-teal/5'}`}
    >
      <StatusDot state={objective.introduced ? 'completed' : 'available'} />
      <span className="min-w-0 max-sm:flex-1">
        <strong className="block overflow-wrap-anywhere text-sm font-semibold">
          {objective.title}
        </strong>
        {/*
          This previously de-slugified objective.moduleId and capitalised it, which produced
          prose-shaped text that was still an identifier. The module's real title is used instead,
          and nothing is shown when it is not available yet.
        */}
        {moduleTitle ? (
          <small className="mt-1 block text-sm text-faint">{moduleTitle}</small>
        ) : null}
      </span>
      <Badge tone={objective.introduced ? 'teal' : 'neutral'}>
        {objective.introduced ? 'Introduced' : 'Not introduced'}
      </Badge>
      <span className="text-right text-base text-faint max-sm:ml-auto max-sm:w-4">→</span>
    </button>
  )
}

function Dimension({
  label,
  value,
  state,
}: {
  label: string
  value: string
  state: keyof typeof dimensionToneStyles
}) {
  return (
    <div className="grid grid-cols-[85px_1fr_100px] items-center gap-3 border-b border-line py-3 max-sm:grid-cols-[72px_1fr_84px]">
      <span className="text-sm text-muted">{label}</span>
      <span className={`h-1 rounded-full ${dimensionToneStyles[state]}`} />
      <strong className="text-right text-sm font-semibold">{value}</strong>
    </div>
  )
}

function ProgressState({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Progress" />
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-muted">{detail}</p>
      </Card>
    </div>
  )
}

function formatEvidenceTime(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
    new Date(value),
  )
}
