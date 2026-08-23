import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useGetCourse } from '../../api/generated/endpoints'
import type { CourseResource } from '../../api/generated/schemas/courseResource.zod'
import type { ModuleSummary } from '../../api/generated/schemas/moduleSummary.zod'
import { modulePath } from '../../app/routes'
import { Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

export function Curriculum() {
  const { setPageContext } = useTutor()
  const { courseId } = useParams()
  const [query, setQuery] = useState('')
  const courseQuery = useGetCourse(courseId ?? '', {
    query: { enabled: Boolean(courseId) },
  })
  const course = courseQuery.data?.data

  useEffect(() => {
    const matchingCourse = course?.id === courseId ? course : undefined

    setPageContext({
      type: 'curriculum',
      title: matchingCourse?.title ?? 'Curriculum',
      courseId,
      courseTitle: matchingCourse?.title,
    })
  }, [course, courseId, setPageContext])

  if (!courseId) {
    return (
      <CurriculumState
        title="Course unavailable"
        detail="This course route is missing its course identity."
      />
    )
  }

  if (courseQuery.isPending) {
    return <CurriculumState title="Loading course" detail="Fetching the available modules…" />
  }

  if (courseQuery.isError) {
    return (
      <CurriculumState
        title="Course unavailable"
        detail={getErrorMessage(courseQuery.error)}
        action={<RetryButton onClick={() => void courseQuery.refetch()} />}
      />
    )
  }

  if (!course || course.id !== courseId) {
    return (
      <CurriculumState
        title="Course unavailable"
        detail="No matching course data was returned for this route."
      />
    )
  }

  if (course.modules.length === 0) {
    return (
      <CurriculumState
        title="No modules yet"
        detail={`${course.title} modules will appear here when they are published.`}
      />
    )
  }

  return <CurriculumContent course={course} query={query} onQueryChange={setQuery} />
}

function CurriculumContent({
  course,
  query,
  onQueryChange,
}: {
  course: CourseResource
  query: string
  onQueryChange: (value: string) => void
}) {
  const filteredModules = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase()
    if (!normalizedQuery) return course.modules

    return course.modules.filter((module) =>
      `${module.title} ${module.id}`.toLowerCase().includes(normalizedQuery),
    )
  }, [course.modules, query])

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title={course.title} detail={course.description} />
      <div className="flex items-center justify-between gap-4 max-sm:items-stretch max-sm:flex-col">
        <label className="flex w-full max-w-sm items-center gap-2 rounded-lg border border-line bg-panel-soft px-3 py-2 text-faint max-sm:max-w-none">
          <span aria-hidden="true">⌕</span>
          <span className="sr-only">Search curriculum</span>
          <input
            className="w-full border-0 bg-transparent text-sm text-ink placeholder:text-faint"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search modules"
          />
        </label>
        <span className="text-sm text-muted">
          {course.modules.length} {course.modules.length === 1 ? 'module' : 'modules'}
        </span>
      </div>

      <section>
        <SectionHeading title="Modules" />
        {filteredModules.length > 0 ? (
          <div className="grid gap-2.5">
            {filteredModules.map((module) => (
              <ModuleRow key={module.id} courseId={course.id} module={module} />
            ))}
          </div>
        ) : (
          <Card muted>
            <p className="text-sm leading-relaxed text-muted">No modules match “{query}”.</p>
          </Card>
        )}
      </section>
    </div>
  )
}

function ModuleRow({ courseId, module }: { courseId: string; module: ModuleSummary }) {
  return (
    <Link
      className="flex items-center gap-4 rounded-lg border border-line bg-panel px-4 py-5 text-ink no-underline transition hover:translate-x-0.5 hover:border-line-strong hover:bg-panel-soft max-sm:gap-3 max-sm:px-3.5 max-sm:py-4"
      to={modulePath(courseId, module.id)}
    >
      <span className="w-16 shrink-0 text-xs font-bold uppercase tracking-wide text-faint">
        Module {String(module.order + 1).padStart(2, '0')}
      </span>
      <span className="min-w-0 flex-1">
        <strong className="block text-base tracking-tight">{module.title}</strong>
        <span className="mt-1 block text-sm text-faint">{module.id}</span>
      </span>
      <span className="text-base text-faint" aria-hidden="true">
        →
      </span>
    </Link>
  )
}

function CurriculumState({
  title,
  detail,
  action,
}: {
  title: string
  detail: string
  action?: ReactNode
}) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Curriculum" />
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-muted">{detail}</p>
        {action ? <div className="mt-4">{action}</div> : null}
      </Card>
    </div>
  )
}

function RetryButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      className="rounded-lg border border-line-strong bg-accent-slate/10 px-3 py-2 text-sm font-bold text-ink transition hover:bg-accent-slate/20"
      type="button"
      onClick={onClick}
    >
      Try again
    </button>
  )
}

function getErrorMessage(error: unknown) {
  return error instanceof Error && error.message
    ? error.message
    : 'The curriculum could not be loaded. Try again in a moment.'
}
