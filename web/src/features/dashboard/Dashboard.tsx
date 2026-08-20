import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useGetCourseProgress, useListActivities } from '../../api/generated/endpoints'
import type { ActivityResource } from '../../api/generated/schemas/activityResource.zod'
import type { CourseProgressResource } from '../../api/generated/schemas/courseProgressResource.zod'
import { coursePath, DEFAULT_COURSE_ID, lessonPath } from '../../app/routes'
import { Badge, Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'
import { formatDashboardDate, formatDashboardGreeting } from './time'

export function Dashboard() {
  const { setPageContext } = useTutor()
  const now = new Date()
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const progressQuery = useGetCourseProgress(DEFAULT_COURSE_ID)
  const activityQuery = useListActivities({ courseId: DEFAULT_COURSE_ID, limit: 6 })

  useEffect(() => {
    setPageContext({ type: 'dashboard', title: 'Home', courseId: DEFAULT_COURSE_ID })
  }, [setPageContext])

  const intro = (
    <PageIntro
      eyebrow={formatDashboardDate(now, timeZone)}
      title={formatDashboardGreeting(now, timeZone)}
      detail="A focused place to pick up where you left off."
    />
  )

  if (progressQuery.isPending || activityQuery.isPending) {
    return <DashboardState intro={intro} title="Loading your learning state" />
  }
  if (progressQuery.isError || activityQuery.isError || !progressQuery.data) {
    return (
      <DashboardState
        intro={intro}
        title="Learning state unavailable"
        detail="Your saved progress and activity could not be loaded. Try again."
      />
    )
  }

  return (
    <DashboardView
      intro={intro}
      progress={progressQuery.data.data}
      activities={activityQuery.data?.data ?? []}
    />
  )
}

export function DashboardView({
  intro,
  progress,
  activities,
}: {
  intro?: React.ReactNode
  progress: CourseProgressResource
  activities: ActivityResource[]
}) {
  const introducedCount = progress.objectives.filter((objective) => objective.introduced).length

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      {intro}

      <section className="grid grid-cols-[minmax(0,1.58fr)_minmax(260px,0.84fr)] gap-3.5 max-lg:grid-cols-1">
        <Card className="min-h-72 border-brand-coral/30 bg-panel p-6 max-sm:min-h-0 max-sm:p-5">
          <Badge tone="coral">Continue learning</Badge>
          {progress.nextLesson ? (
            <div className="mt-8 max-w-2xl">
              <h2 className="text-4xl font-semibold leading-none tracking-tight max-sm:text-3xl">
                {progress.nextLesson.lessonTitle}
              </h2>
              <p className="mt-4 text-xs font-semibold text-brand-coral">
                {progress.nextLesson.moduleTitle}
              </p>
              <p className="my-4 max-w-lg text-xs leading-relaxed text-muted">
                This is the next incomplete lesson in curriculum order.
              </p>
              <Link
                className="inline-flex items-center justify-center gap-2.5 rounded-lg bg-brand-teal px-4 py-2.5 text-xs font-bold text-brand-ink no-underline transition hover:-translate-y-px hover:bg-brand-teal-light"
                to={lessonPath(
                  progress.nextLesson.courseId,
                  progress.nextLesson.moduleId,
                  progress.nextLesson.lessonId,
                )}
              >
                Continue lesson <span>→</span>
              </Link>
            </div>
          ) : (
            <div className="mt-8 max-w-xl">
              <h2 className="text-3xl font-semibold tracking-tight">
                All current lessons complete
              </h2>
              <p className="my-4 text-xs leading-relaxed text-muted">
                There is no incomplete authored lesson in this course right now.
              </p>
              <Link
                className="text-xs font-bold text-brand-teal no-underline hover:text-ink"
                to={coursePath(progress.courseId)}
              >
                Browse curriculum →
              </Link>
            </div>
          )}
        </Card>

        <div className="grid gap-3.5 max-sm:gap-2">
          <AvailabilityCard
            icon="↺"
            eyebrow="Reviews"
            title="Not available yet"
            detail="Recall remains not assessed until the review workflow is implemented."
          />
          <AvailabilityCard
            icon="◎"
            eyebrow="Mastery checks"
            title="No checks available"
            detail="Lesson completion does not create mastery claims."
          />
        </div>
      </section>

      <section className="grid grid-cols-[1.05fr_0.95fr] gap-3.5 max-lg:grid-cols-1">
        <Card className="min-h-72">
          <SectionHeading eyebrow="Recent activity" title="A quiet trail of progress" />
          <ActivityList activities={activities} />
        </Card>
        <Card className="min-h-72">
          <SectionHeading eyebrow="Projects" title="Tracking not available yet" />
          <p className="text-xs leading-relaxed text-muted">
            Projects remain Git-based learning work. This dashboard does not claim project progress
            before project state is implemented.
          </p>
          <Link
            className="mt-6 inline-flex text-xs font-bold text-brand-teal no-underline hover:text-ink"
            to="/projects"
          >
            Browse project plans →
          </Link>
        </Card>
      </section>

      <section className="flex items-center justify-between gap-5 border-t border-line pt-6 max-sm:block">
        <div className="flex max-w-xl items-start gap-3">
          <span className="text-base text-brand-gold">✦</span>
          <div>
            <strong className="text-xs">What this progress means</strong>
            <p className="mt-1.5 text-xs leading-normal text-muted">
              Completed lessons introduce objectives. Recall, application, and transfer are still
              not assessed.
            </p>
          </div>
        </div>
        <div className="flex gap-7 max-sm:mt-5 max-sm:justify-between">
          <DashboardStat value={progress.completedLessonCount} label="lessons complete" />
          <DashboardStat value={introducedCount} label="objectives introduced" />
        </div>
      </section>
    </div>
  )
}

export function ActivityList({ activities }: { activities: ActivityResource[] }) {
  if (activities.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-line p-5">
        <strong className="text-xs text-ink">No activity yet</strong>
        <p className="mt-2 text-2xs leading-relaxed text-muted">
          Mark a lesson complete to start a learner activity history.
        </p>
      </div>
    )
  }

  return (
    <div className="grid">
      {activities.map((activity) => {
        const content = (
          <>
            <span className="grid size-6 place-items-center rounded-lg bg-brand-gold/10 text-xs text-brand-gold">
              ✓
            </span>
            <div>
              <strong className="block text-xs font-semibold">
                Completed {activity.lessonTitle ?? 'lesson'}
              </strong>
              <span className="mt-1 block text-2xs text-faint">
                {activity.moduleTitle ?? activity.courseTitle}
              </span>
            </div>
            <time className="text-2xs text-faint" dateTime={activity.occurredAt}>
              {formatActivityTime(activity.occurredAt)}
            </time>
          </>
        )
        const className =
          'grid grid-cols-[24px_1fr_auto] items-center gap-2.5 border-t border-line py-2.5 text-ink no-underline'
        return activity.moduleId && activity.lessonId ? (
          <Link
            className={`${className} hover:text-brand-teal`}
            key={activity.id}
            to={lessonPath(activity.courseId, activity.moduleId, activity.lessonId)}
          >
            {content}
          </Link>
        ) : (
          <div className={className} key={activity.id}>
            {content}
          </div>
        )
      })}
    </div>
  )
}

function AvailabilityCard({
  icon,
  eyebrow,
  title,
  detail,
}: {
  icon: string
  eyebrow: string
  title: string
  detail: string
}) {
  return (
    <Card className="flex min-h-32 items-start gap-4 p-5 max-sm:min-h-28">
      <div className="grid size-9 place-items-center rounded-lg bg-brand-slate/20 text-xl text-muted">
        {icon}
      </div>
      <div>
        <p className="mt-px text-2xs font-bold uppercase tracking-widest text-faint">{eyebrow}</p>
        <h3 className="my-2 text-lg tracking-tight">{title}</h3>
        <p className="text-2xs leading-relaxed text-faint">{detail}</p>
      </div>
    </Card>
  )
}

function DashboardStat({ value, label }: { value: number; label: string }) {
  return (
    <div className="grid gap-1">
      <strong className="text-2xl tracking-tight">{value}</strong>
      <span className="text-2xs uppercase tracking-wide text-faint">{label}</span>
    </div>
  )
}

function DashboardState({
  intro,
  title,
  detail = 'Reading your saved progress and recent activity…',
}: {
  intro: React.ReactNode
  title: string
  detail?: string
}) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      {intro}
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-xs leading-relaxed text-muted">{detail}</p>
      </Card>
    </div>
  )
}

function formatActivityTime(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(new Date(value))
}
