import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  getCourseModuleWorkbook,
  useGetCourse,
  useGetCourseModule,
} from '../../api/generated/endpoints'
import type { CourseResource } from '../../api/generated/schemas/courseResource.zod'
import type { ModuleResource } from '../../api/generated/schemas/moduleResource.zod'
import { coursePath, exercisePath, lessonPath, worksheetPath } from '../../app/routes'
import { Badge, Button, Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'
import { downloadPdf, pdfDownloadErrorMessage } from '../worksheets/downloadPdf'
import { hasWorkbook } from '../worksheets/workbookAvailability'
import { safeExternalUrl } from './externalLinks'

export function ModuleDetail() {
  const { courseId, moduleId } = useParams()
  const courseQuery = useGetCourse(courseId ?? '', {
    query: { enabled: Boolean(courseId) },
  })
  const moduleQuery = useGetCourseModule(courseId ?? '', moduleId ?? '', {
    query: { enabled: Boolean(courseId && moduleId) },
  })
  const { setPageContext } = useTutor()
  const course = courseQuery.data?.data
  const module = moduleQuery.data?.data
  const matchingCourse = course?.id === courseId ? course : undefined
  const matchingModule =
    matchingCourse &&
    module?.courseId === matchingCourse.id &&
    module.id === moduleId &&
    matchingCourse.modules.some((item) => item.id === module.id)
      ? module
      : undefined

  useEffect(() => {
    setPageContext({
      type: 'curriculum',
      title: matchingModule?.title ?? matchingCourse?.title ?? 'Curriculum',
      courseId,
      courseTitle: matchingCourse?.title,
      moduleId,
      moduleTitle: matchingModule?.title,
      objectiveIds: matchingModule?.objectives.map((objective) => objective.id),
    })
  }, [courseId, matchingCourse, matchingModule, moduleId, setPageContext])

  if (!courseId || !moduleId) {
    return (
      <ModuleState
        title="Module unavailable"
        detail="This module route is missing its course or module identity."
      />
    )
  }

  if (courseQuery.isPending || moduleQuery.isPending) {
    return (
      <ModuleState
        courseId={matchingCourse?.id}
        title="Loading module"
        detail="Fetching the module details…"
      />
    )
  }

  if (courseQuery.isError || moduleQuery.isError) {
    return (
      <ModuleState
        courseId={matchingCourse?.id}
        title="Module unavailable"
        detail={getErrorMessage(courseQuery.error ?? moduleQuery.error)}
      />
    )
  }

  if (
    !course ||
    !module ||
    course.id !== courseId ||
    module.courseId !== course.id ||
    module.id !== moduleId ||
    !course.modules.some((item) => item.id === module.id)
  ) {
    return (
      <ModuleState
        courseId={matchingCourse?.id}
        title="Module unavailable"
        detail="No matching module data was returned for this course route."
      />
    )
  }

  return <ModuleContent course={course} module={module} />
}

function ModuleContent({ course, module }: { course: CourseResource; module: ModuleResource }) {
  const [downloadingWorkbook, setDownloadingWorkbook] = useState<'student' | 'solutions' | null>(
    null,
  )
  const [workbookError, setWorkbookError] = useState<string | null>(null)
  const objectiveTitles = new Map(
    module.objectives.map((objective) => [objective.id, objective.title]),
  )

  async function downloadWorkbook(workbookId: 'student' | 'solutions') {
    setDownloadingWorkbook(workbookId)
    setWorkbookError(null)
    try {
      const response = await getCourseModuleWorkbook(course.id, module.id, workbookId)
      const fallbackFilename =
        workbookId === 'student'
          ? `${module.id}-workbook.pdf`
          : `${module.id}-workbook-solutions.pdf`
      downloadPdf(response.data, response.headers['content-disposition'], fallbackFilename)
    } catch (error) {
      setWorkbookError(pdfDownloadErrorMessage(error))
    } finally {
      setDownloadingWorkbook(null)
    }
  }

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <Link
        className="justify-self-start text-xs font-bold text-muted no-underline hover:text-ink"
        to={coursePath(course.id)}
      >
        ← {course.title}
      </Link>
      <PageIntro
        compact
        eyebrow={`Module ${String(module.order + 1).padStart(2, '0')}`}
        title={module.title}
        detail={`${module.lessons.length} ${module.lessons.length === 1 ? 'lesson' : 'lessons'}`}
      />

      <section>
        <SectionHeading title="Objectives" />
        {module.objectives.length > 0 ? (
          <div className="grid gap-2">
            {module.objectives.map((objective) => (
              <div key={objective.id} className="border-t border-line py-3">
                <strong className="block text-xs">{objective.title}</strong>
                <span className="mt-1 block text-2xs leading-normal text-muted">
                  {objective.description}
                </span>
                <span className="mt-2 block text-2xs text-faint">{objective.id}</span>
                {objective.prerequisites.length > 0 ? (
                  <div className="mt-3 border-t border-line pt-2">
                    <span className="text-2xs font-bold uppercase tracking-wide text-faint">
                      Prerequisites
                    </span>
                    <ul className="mt-1 grid gap-1 pl-4 text-2xs text-muted">
                      {objective.prerequisites.map((prerequisite) => (
                        <li key={prerequisite}>
                          {objectiveTitles.get(prerequisite) ?? prerequisite}
                          {objectiveTitles.has(prerequisite) ? (
                            <span className="ml-1 text-faint">({prerequisite})</span>
                          ) : null}
                        </li>
                      ))}
                    </ul>
                  </div>
                ) : (
                  <span className="mt-2 block text-2xs text-faint">No prerequisites</span>
                )}
              </div>
            ))}
          </div>
        ) : (
          <Card muted>
            <p className="text-xs leading-relaxed text-muted">
              No objectives are recorded for this module.
            </p>
          </Card>
        )}
      </section>

      <section>
        <SectionHeading title="Lessons" />
        {module.lessons.length > 0 ? (
          <div className="grid">
            {module.lessons.map((lesson, index) => (
              <Link
                key={lesson.id}
                className="grid grid-cols-[35px_minmax(0,1fr)_17px] items-center gap-3 border-t border-line py-3 text-left text-ink no-underline hover:text-brand-teal"
                to={lessonPath(course.id, module.id, lesson.id)}
              >
                <span className="text-2xs text-faint">{String(index + 1).padStart(2, '0')}</span>
                <span>
                  <strong className="block text-xs">{lesson.title}</strong>
                  <small className="mt-1 block text-2xs text-faint">{lesson.id}</small>
                </span>
                <span className="text-right text-base text-faint" aria-hidden="true">
                  →
                </span>
              </Link>
            ))}
          </div>
        ) : (
          <Card muted>
            <p className="text-xs leading-relaxed text-muted">
              No lessons are published for this module yet.
            </p>
          </Card>
        )}
      </section>

      {hasWorkbook(module.worksheets.length) ? (
        <section>
          <SectionHeading
            eyebrow="Practice"
            title="Worksheets"
            detail="Written practice grouped in lesson order."
          />
          <div className="mb-4 flex flex-wrap items-center justify-between gap-4 rounded-xl border border-line bg-panel px-5 py-4 shadow-lg">
            <div>
              <strong className="block text-sm text-ink">Module workbook</strong>
              <span className="mt-1 block text-2xs leading-relaxed text-muted">
                Download every worksheet in this module as one printable PDF.
              </span>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                disabled={downloadingWorkbook !== null}
                onClick={() => void downloadWorkbook('student')}
              >
                {downloadingWorkbook === 'student' ? 'Preparing workbook…' : 'Download workbook'}
              </Button>
              <Button
                disabled={downloadingWorkbook !== null}
                onClick={() => void downloadWorkbook('solutions')}
                variant="outline"
              >
                {downloadingWorkbook === 'solutions'
                  ? 'Preparing solutions…'
                  : 'Download solutions'}
              </Button>
            </div>
            {workbookError ? (
              <p className="w-full text-xs text-brand-coral" role="alert">
                {workbookError}
              </p>
            ) : null}
          </div>
          <div className="grid gap-2">
            {module.worksheets.map((worksheet) => {
              const lesson = module.lessons.find((item) => item.id === worksheet.lessonId)
              return (
                <Link
                  key={worksheet.id}
                  className="flex items-center gap-3 rounded-lg border border-line bg-panel px-4 py-4 text-ink no-underline transition hover:border-line-strong hover:text-brand-teal"
                  to={worksheetPath(course.id, module.id, worksheet.id)}
                >
                  <Badge tone="teal">Worksheet</Badge>
                  <span className="min-w-0 flex-1">
                    <strong className="block text-xs">{worksheet.title}</strong>
                    <small className="mt-1 block text-2xs text-faint">
                      {lesson?.title ?? worksheet.lessonId} · {worksheet.problemCount}{' '}
                      {worksheet.problemCount === 1 ? 'problem' : 'problems'}
                    </small>
                  </span>
                  <span className="text-base text-faint" aria-hidden="true">
                    →
                  </span>
                </Link>
              )
            })}
          </div>
        </section>
      ) : null}

      {module.exercises.length > 0 ? (
        <section>
          <SectionHeading
            eyebrow="Practice"
            title="Coding exercises"
            detail="Small Python exercises run in your browser."
          />
          <div className="grid gap-2">
            {module.exercises.map((exercise) => {
              const lesson = module.lessons.find((item) => item.id === exercise.lessonId)
              return (
                <Link
                  key={exercise.id}
                  className="flex items-center gap-3 rounded-lg border border-line bg-panel px-4 py-4 text-ink no-underline transition hover:border-line-strong hover:text-brand-teal"
                  to={exercisePath(course.id, module.id, exercise.id)}
                >
                  <Badge tone="gold">Python</Badge>
                  <span className="min-w-0 flex-1">
                    <strong className="block text-xs">{exercise.title}</strong>
                    <small className="mt-1 block text-2xs text-faint">
                      {lesson?.title ?? exercise.lessonId}
                    </small>
                  </span>
                  <span className="text-base text-faint" aria-hidden="true">
                    →
                  </span>
                </Link>
              )
            })}
          </div>
        </section>
      ) : null}

      {module.videos.length > 0 ? (
        <section>
          <SectionHeading title="Videos" />
          <div className="grid gap-2">
            {module.videos.map((video) => {
              const url = safeExternalUrl(video.url)
              return (
                <div
                  key={video.id}
                  className="flex items-center gap-3 rounded-lg border border-line bg-panel px-3 py-3"
                >
                  <Badge tone="violet">Video</Badge>
                  <span className="min-w-0 flex-1">
                    <strong className="block text-xs">{video.title}</strong>
                    <small className="mt-1 block text-2xs text-faint">{video.id}</small>
                  </span>
                  {url ? (
                    <a
                      className="text-xs font-bold text-brand-teal no-underline hover:text-ink"
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      Open ↗
                    </a>
                  ) : (
                    <span className="text-2xs text-faint">Link unavailable</span>
                  )}
                </div>
              )
            })}
          </div>
        </section>
      ) : null}
    </div>
  )
}

function ModuleState({
  courseId,
  title,
  detail,
}: {
  courseId?: string
  title: string
  detail: string
}) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <Link
        className="justify-self-start text-xs font-bold text-muted no-underline hover:text-ink"
        to={courseId ? coursePath(courseId) : '/curriculum'}
      >
        ← Curriculum
      </Link>
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-xs leading-relaxed text-muted">{detail}</p>
      </Card>
    </div>
  )
}

function getErrorMessage(error: unknown) {
  return error instanceof Error && error.message
    ? error.message
    : 'The module could not be loaded. Check the module link and try again.'
}
