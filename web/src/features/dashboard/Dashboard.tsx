import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useGetCourseProgress, useListActivities } from '../../api/generated/endpoints'
import type { ActivityResource } from '../../api/generated/schemas/activityResource.zod'
import type { CourseProgressResource } from '../../api/generated/schemas/courseProgressResource.zod'
import {
  coursePath,
  DEFAULT_COURSE_ID,
  exercisePath,
  lessonPath,
  modulePath,
} from '../../app/routes'
import { Badge, Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'
import { useAuth } from '../authentication/AuthContext'
import { SignInRequired } from '../authentication/SignInRequired'
import { formatDashboardDate, formatDashboardGreeting } from './time'

export function Dashboard() {
  const auth = useAuth()
  const { setPageContext } = useTutor()
  const now = new Date()
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const progressQuery = useGetCourseProgress(DEFAULT_COURSE_ID, {
    query: { enabled: auth.isAuthenticated },
  })
  const activityQuery = useListActivities(
    { courseId: DEFAULT_COURSE_ID, limit: 6 },
    { query: { enabled: auth.isAuthenticated } },
  )

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

  if (auth.isPending) {
    return <DashboardState intro={intro} title="Checking learner access" />
  }
  if (!auth.isAuthenticated) {
    return (
      <SignInRequired
        title="Your learning dashboard"
        detail="Sign in to see saved progress and activity. You can browse the complete curriculum without an account."
        returnTo="/"
      />
    )
  }

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
        <Card className="min-h-72 border-accent-coral/30 bg-panel p-6 max-sm:min-h-0 max-sm:p-5">
          <Badge tone="coral">Continue learning</Badge>
          {progress.nextLesson ? (
            <div className="mt-8 max-w-2xl">
              <h2 className="text-4xl font-semibold leading-none tracking-tight max-sm:text-3xl">
                {progress.nextLesson.lessonTitle}
              </h2>
              <p className="mt-4 text-sm font-semibold text-accent-coral">
                {progress.nextLesson.moduleTitle}
              </p>
              <p className="my-4 max-w-lg text-sm leading-relaxed text-muted">
                This is the next incomplete lesson in curriculum order.
              </p>
              <Link
                className="inline-flex items-center justify-center gap-2.5 rounded-lg bg-brand-teal px-4 py-2.5 text-sm font-bold text-brand-ink no-underline transition hover:-translate-y-px hover:bg-brand-teal-light"
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
              <p className="my-4 text-sm leading-relaxed text-muted">
                There is no incomplete authored lesson in this course right now.
              </p>
              <Link
                className="text-sm font-bold text-accent-teal no-underline hover:text-ink"
                to={coursePath(progress.courseId)}
              >
                Browse curriculum →
              </Link>
            </div>
          )}
        </Card>

        <div className="grid gap-3.5 max-sm:gap-2">
          <ActionCard
            icon="↺"
            eyebrow="Reviews"
            title={`${progress.dueReviewCount} due now`}
            detail="Review authored recall prompts with server-computed FSRS scheduling."
            to="/review"
          />
          {progress.practiceExercise ? (
            <ActionCard
              icon="◎"
              eyebrow="Practice"
              title={progress.practiceExercise.exerciseTitle}
              detail={`Exercise in ${progress.practiceExercise.moduleTitle}`}
              to={exercisePath(
                progress.practiceExercise.courseId,
                progress.practiceExercise.moduleId,
                progress.practiceExercise.exerciseId,
              )}
            />
          ) : (
            <ActionCard
              icon="◎"
              eyebrow="Practice"
              title="No exercise cue"
              detail="No introduced objective currently needs an authored exercise attempt."
            />
          )}
        </div>
      </section>

      <section className="grid grid-cols-[1.05fr_0.95fr] gap-3.5 max-lg:grid-cols-1">
        <Card className="min-h-72">
          <SectionHeading eyebrow="Recent activity" title="A quiet trail of progress" />
          <ActivityList activities={activities} />
        </Card>
        <Card className="min-h-72">
          <SectionHeading eyebrow="Evidence" title="What has been recorded" />
          <div className="grid gap-4 text-sm text-muted">
            <p>
              {progress.completedLessonCount} completed lessons introduce their linked objectives.
            </p>
            <p>{progress.dueReviewCount} authored recall prompts are currently due.</p>
            <Link className="font-bold text-accent-teal no-underline hover:text-ink" to="/progress">
              Inspect objective evidence →
            </Link>
          </div>
        </Card>
      </section>

      <section className="flex items-center justify-between gap-5 border-t border-line pt-6 max-sm:block">
        <div className="flex max-w-xl items-start gap-3">
          <span className="text-base text-accent-gold">✦</span>
          <div>
            <strong className="text-sm">What this progress means</strong>
            <p className="mt-1.5 text-sm leading-normal text-muted">
              Lesson completion, review history, and checked exercise attempts are shown as separate
              evidence. Transfer stays unassessed until a real transfer workflow exists.
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
        <strong className="text-sm text-ink">No activity yet</strong>
        <p className="mt-2 text-sm leading-relaxed text-muted">
          Complete a lesson or check an exercise to start a learner activity history.
        </p>
      </div>
    )
  }

  return (
    <div className="grid">
      {activities.map((activity) => {
        const exerciseChecked = activity.kind === 'exercise_checked' && Boolean(activity.exerciseId)
        const reviewCompleted = activity.kind === 'review_completed'
        const videoCompleted = activity.kind === 'video_completed' && Boolean(activity.videoId)
        const content = (
          <>
            <span className="grid size-6 place-items-center rounded-lg bg-accent-gold/10 text-sm text-accent-gold">
              ✓
            </span>
            <div>
              <strong className="block text-sm font-semibold">
                {reviewCompleted
                  ? `Reviewed ${activity.reviewItemId ?? 'recall prompt'}`
                  : videoCompleted
                    ? `Watched ${activity.videoTitle ?? activity.videoId}`
                    : exerciseChecked
                      ? `Checked ${activity.exerciseTitle ?? activity.exerciseId}`
                      : `Completed ${activity.lessonTitle ?? 'lesson'}`}
              </strong>
              <span className="mt-1 block text-sm text-faint">
                {activity.moduleTitle ?? activity.courseTitle}
              </span>
            </div>
            <time className="text-sm text-faint" dateTime={activity.occurredAt}>
              {formatActivityTime(activity.occurredAt)}
            </time>
          </>
        )
        const className =
          'grid grid-cols-[24px_1fr_auto] items-center gap-2.5 border-t border-line py-2.5 text-ink no-underline'
        return reviewCompleted ? (
          <Link className={`${className} hover:text-accent-teal`} key={activity.id} to="/review">
            {content}
          </Link>
        ) : videoCompleted && activity.moduleId ? (
          <Link
            className={`${className} hover:text-accent-teal`}
            key={activity.id}
            to={modulePath(activity.courseId, activity.moduleId)}
          >
            {content}
          </Link>
        ) : exerciseChecked && activity.moduleId && activity.exerciseId ? (
          <Link
            className={`${className} hover:text-accent-teal`}
            key={activity.id}
            to={exercisePath(activity.courseId, activity.moduleId, activity.exerciseId)}
          >
            {content}
          </Link>
        ) : activity.moduleId && activity.lessonId ? (
          <Link
            className={`${className} hover:text-accent-teal`}
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

function ActionCard({
  icon,
  eyebrow,
  title,
  detail,
  to,
}: {
  icon: string
  eyebrow: string
  title: string
  detail: string
  to?: string
}) {
  return (
    <Card className="flex min-h-32 items-start gap-4 p-5 max-sm:min-h-28">
      <div className="grid size-9 place-items-center rounded-lg bg-accent-slate/20 text-xl text-muted">
        {icon}
      </div>
      <div>
        <p className="mt-px text-xs font-bold uppercase tracking-widest text-faint">{eyebrow}</p>
        <h3 className="my-2 text-lg tracking-tight">{title}</h3>
        <p className="text-sm leading-relaxed text-faint">{detail}</p>
        {to ? (
          <Link
            className="mt-3 inline-flex text-sm font-bold text-accent-teal no-underline"
            to={to}
          >
            Open →
          </Link>
        ) : null}
      </div>
    </Card>
  )
}

function DashboardStat({ value, label }: { value: number; label: string }) {
  return (
    <div className="grid gap-1">
      <strong className="text-2xl tracking-tight">{value}</strong>
      <span className="text-xs uppercase tracking-wide text-faint">{label}</span>
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
        <p className="mt-2 text-sm leading-relaxed text-muted">{detail}</p>
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
