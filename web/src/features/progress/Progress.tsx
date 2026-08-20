import { useEffect, useState } from 'react'
import { useGetCourseProgress } from '../../api/generated/endpoints'
import type { CourseProgressResource } from '../../api/generated/schemas/courseProgressResource.zod'
import type { ObjectiveProgressResource } from '../../api/generated/schemas/objectiveProgressResource.zod'
import { DEFAULT_COURSE_ID } from '../../app/routes'
import { Badge, Card, PageIntro, SectionHeading, StatusDot } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

const summaryToneStyles = {
  teal: 'border-t-brand-teal',
  gold: 'border-t-brand-gold',
  coral: 'border-t-brand-coral',
  violet: 'border-t-brand-violet',
} as const

const dimensionToneStyles = {
  introduced: 'bg-brand-teal',
  pending: 'bg-brand-gold',
  'not-assessed': 'bg-brand-slate/20',
} as const

export function Progress() {
  const { setPageContext } = useTutor()
  const progressQuery = useGetCourseProgress(DEFAULT_COURSE_ID)
  const [selectedId, setSelectedId] = useState<string>()
  const progress = progressQuery.data?.data
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

  return <ProgressView progress={progress} selectedId={selected?.id} onSelect={setSelectedId} />
}

export function ProgressView({
  progress,
  selectedId,
  onSelect,
}: {
  progress: CourseProgressResource
  selectedId?: string
  onSelect: (objectiveId: string) => void
}) {
  const introducedCount = progress.objectives.filter((objective) => objective.introduced).length
  const selected =
    progress.objectives.find((objective) => objective.id === selectedId) ?? progress.objectives[0]

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro
        compact
        title="Progress"
        detail={`${progress.completedLessonCount} of ${progress.totalLessonCount} lessons complete`}
      />
      <section className="grid grid-cols-4 gap-2.5 max-lg:grid-cols-2 max-sm:grid-cols-2">
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
        <SummaryStat value="—" label="Recall" note="not assessed" tone="coral" />
        <SummaryStat value="—" label="Application / transfer" note="not assessed" tone="violet" />
      </section>
      {selected ? (
        <div className="grid grid-cols-[1.15fr_0.85fr] gap-3.5 max-lg:grid-cols-1 max-sm:gap-2.5">
          <Card className="min-h-128">
            <SectionHeading
              title="Objectives"
              detail="Lesson completion introduces objectives; other evidence arrives from later practice."
            />
            <div className="grid">
              {progress.objectives.map((objective) => (
                <ObjectiveBrowserRow
                  key={objective.id}
                  objective={objective}
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
          <p className="mt-2 text-xs leading-relaxed text-muted">
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
      <p className="text-xs leading-relaxed text-muted">{objective.description}</p>
      <div className="mt-8 grid gap-0">
        <Dimension
          label="Introduced"
          value={objective.introduced ? 'Yes' : 'Not yet'}
          state={objective.introduced ? 'introduced' : 'pending'}
        />
        <Dimension label="Recall" value="Not assessed" state="not-assessed" />
        <Dimension label="Application" value="Not assessed" state="not-assessed" />
        <Dimension label="Transfer" value="Not assessed" state="not-assessed" />
      </div>
      <p className="mt-8 rounded-lg border border-line bg-brand-slate/10 p-4 text-2xs leading-relaxed text-muted">
        Completing a lesson records that you encountered this objective. It does not claim recall,
        application, transfer, or mastery.
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
      <strong className="text-xs">{label}</strong>
      <span className="text-2xs text-faint">{note}</span>
    </Card>
  )
}

function ObjectiveBrowserRow({
  objective,
  selected,
  onClick,
}: {
  objective: ObjectiveProgressResource
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`grid w-full grid-cols-[11px_minmax(0,1fr)_auto_15px] items-center gap-2.5 border-0 border-b border-line bg-transparent py-3 text-left text-ink max-sm:flex max-sm:min-w-0 ${selected ? 'border-l-2 border-brand-teal bg-brand-teal/5 pl-2' : 'hover:bg-brand-teal/5'}`}
    >
      <StatusDot state={objective.introduced ? 'completed' : 'available'} />
      <span className="min-w-0 max-sm:flex-1">
        <strong className="block overflow-wrap-anywhere text-xs font-semibold">
          {objective.title}
        </strong>
        <small className="mt-1 block text-2xs capitalize text-faint">
          {objective.moduleId.replaceAll('-', ' ')}
        </small>
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
      <span className="text-2xs text-muted">{label}</span>
      <span className={`h-1 rounded-full ${dimensionToneStyles[state]}`} />
      <strong className="text-right text-2xs font-semibold">{value}</strong>
    </div>
  )
}

function ProgressState({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Progress" />
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-xs leading-relaxed text-muted">{detail}</p>
      </Card>
    </div>
  )
}
