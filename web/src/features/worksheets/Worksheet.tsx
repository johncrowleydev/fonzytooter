import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  getCourseModuleWorksheetDocument,
  useGetCourse,
  useGetCourseModule,
  useGetCourseModuleWorksheet,
} from '../../api/generated/endpoints'
import type { WorksheetResource } from '../../api/generated/schemas/worksheetResource.zod'
import { coursePath, lessonPath, modulePath } from '../../app/routes'
import { Badge, Button, Card, PageIntro } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'
import { downloadPdf, pdfDownloadErrorMessage } from './downloadPdf'
import { WorksheetMarkup } from './WorksheetMarkup'

export function Worksheet() {
  const { courseId, moduleId, worksheetId } = useParams()
  const courseQuery = useGetCourse(courseId ?? '', {
    query: { enabled: Boolean(courseId) },
  })
  const moduleQuery = useGetCourseModule(courseId ?? '', moduleId ?? '', {
    query: { enabled: Boolean(courseId && moduleId) },
  })
  const worksheetQuery = useGetCourseModuleWorksheet(
    courseId ?? '',
    moduleId ?? '',
    worksheetId ?? '',
    { query: { enabled: Boolean(courseId && moduleId && worksheetId) } },
  )
  const { setPageContext } = useTutor()
  const course = courseQuery.data?.data
  const module = moduleQuery.data?.data
  const worksheet = worksheetQuery.data?.data
  const matchingCourse = course?.id === courseId ? course : undefined
  const matchingModule =
    matchingCourse &&
    module?.courseId === matchingCourse.id &&
    module.id === moduleId &&
    matchingCourse.modules.some((item) => item.id === module.id)
      ? module
      : undefined
  const worksheetSummary = matchingModule?.worksheets.find((item) => item.id === worksheetId)
  const matchingWorksheet =
    matchingModule &&
    worksheetSummary &&
    worksheet &&
    worksheet.courseId === matchingCourse?.id &&
    worksheet.moduleId === matchingModule.id &&
    worksheet.id === worksheetId &&
    worksheet.lessonId === worksheetSummary.lessonId
      ? worksheet
      : undefined
  const lesson = matchingWorksheet
    ? matchingModule?.lessons.find((item) => item.id === matchingWorksheet.lessonId)
    : undefined

  useEffect(() => {
    setPageContext({
      type: 'curriculum',
      title:
        matchingWorksheet?.title ?? matchingModule?.title ?? matchingCourse?.title ?? 'Practice',
      courseId,
      courseTitle: matchingCourse?.title,
      lessonId: lesson?.id,
      lessonTitle: lesson?.title,
      moduleId,
      moduleTitle: matchingModule?.title,
      objectiveIds: matchingWorksheet?.objectiveIds,
    })
  }, [
    courseId,
    lesson,
    matchingCourse,
    matchingModule,
    matchingWorksheet,
    moduleId,
    setPageContext,
  ])

  if (!courseId || !moduleId || !worksheetId) {
    return (
      <WorksheetState
        title="Worksheet unavailable"
        detail="This worksheet route is missing its course, module, or worksheet identity."
      />
    )
  }

  if (courseQuery.isPending || moduleQuery.isPending || worksheetQuery.isPending) {
    return (
      <WorksheetState
        courseId={matchingCourse?.id}
        moduleId={matchingModule?.id}
        title="Loading worksheet"
        detail="Fetching the worksheet and its curriculum context…"
      />
    )
  }

  if (courseQuery.isError || moduleQuery.isError || worksheetQuery.isError) {
    return (
      <WorksheetState
        courseId={matchingCourse?.id}
        moduleId={matchingModule?.id}
        title="Worksheet unavailable"
        detail={getErrorMessage(courseQuery.error ?? moduleQuery.error ?? worksheetQuery.error)}
      />
    )
  }

  if (!course || !module || !worksheet) {
    return (
      <WorksheetState
        courseId={matchingCourse?.id}
        moduleId={matchingModule?.id}
        title="Worksheet unavailable"
        detail="No worksheet data was returned for this route."
      />
    )
  }

  if (!matchingCourse || !matchingModule || !matchingWorksheet || !lesson) {
    return (
      <WorksheetState
        courseId={matchingCourse?.id}
        moduleId={matchingModule?.id}
        title="Worksheet unavailable"
        detail="This worksheet does not belong to the requested course and module. Use the module page to choose practice."
      />
    )
  }

  return (
    <WorksheetContent
      courseId={matchingCourse.id}
      courseTitle={matchingCourse.title}
      moduleId={matchingModule.id}
      moduleTitle={matchingModule.title}
      lessonId={lesson.id}
      lessonTitle={lesson.title}
      worksheet={matchingWorksheet}
    />
  )
}

function WorksheetContent({
  courseId,
  courseTitle,
  moduleId,
  moduleTitle,
  lessonId,
  lessonTitle,
  worksheet,
}: {
  courseId: string
  courseTitle: string
  moduleId: string
  moduleTitle: string
  lessonId: string
  lessonTitle: string
  worksheet: WorksheetResource
}) {
  const [downloading, setDownloading] = useState<'student' | 'solutions' | null>(null)
  const [downloadError, setDownloadError] = useState<string | null>(null)

  async function download(documentId: 'student' | 'solutions') {
    setDownloading(documentId)
    setDownloadError(null)
    try {
      const response = await getCourseModuleWorksheetDocument(
        courseId,
        moduleId,
        worksheet.id,
        documentId,
      )
      const fallbackFilename =
        documentId === 'student' ? `${worksheet.id}.pdf` : `${worksheet.id}-solutions.pdf`
      downloadPdf(response.data, response.headers['content-disposition'], fallbackFilename)
    } catch (error) {
      setDownloadError(pdfDownloadErrorMessage(error))
    } finally {
      setDownloading(null)
    }
  }

  return (
    <div className="grid max-w-5xl gap-7 max-sm:gap-5">
      <nav
        className="flex flex-wrap items-center gap-2 text-2xs text-faint"
        aria-label="Worksheet breadcrumbs"
      >
        <Link className="text-muted no-underline hover:text-brand-teal" to={coursePath(courseId)}>
          {courseTitle}
        </Link>
        <span>/</span>
        <Link
          className="text-muted no-underline hover:text-brand-teal"
          to={modulePath(courseId, moduleId)}
        >
          {moduleTitle}
        </Link>
        <span>/</span>
        <Link
          className="text-muted no-underline hover:text-brand-teal"
          to={lessonPath(courseId, moduleId, lessonId)}
        >
          {lessonTitle}
        </Link>
        <span>/</span>
        <span>{worksheet.title}</span>
      </nav>

      <PageIntro
        compact
        eyebrow="Worksheet · Read-only practice"
        title={worksheet.title}
        detail={`${worksheet.problems.length} ${worksheet.problems.length === 1 ? 'problem' : 'problems'} · ${lessonTitle}`}
      >
        <div className="mt-5 flex flex-wrap gap-2">
          <Badge tone="teal">Practice</Badge>
          <Badge>{moduleTitle}</Badge>
        </div>
      </PageIntro>

      <Card className="max-w-3xl">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-sm font-semibold text-ink">Printable practice</p>
            <p className="mt-1 text-xs leading-relaxed text-muted">
              Download a blank worksheet or a copy with authored solutions.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button disabled={downloading !== null} onClick={() => void download('student')}>
              {downloading === 'student' ? 'Preparing worksheet…' : 'Download worksheet PDF'}
            </Button>
            <Button
              disabled={downloading !== null}
              onClick={() => void download('solutions')}
              variant="outline"
            >
              {downloading === 'solutions' ? 'Preparing solutions…' : 'Download solutions PDF'}
            </Button>
          </div>
        </div>
        {downloadError ? (
          <p className="mt-3 text-xs text-brand-coral" role="alert">
            {downloadError}
          </p>
        ) : null}
      </Card>

      <Card className="max-w-3xl">
        <p className="mb-3 text-2xs font-bold uppercase tracking-widest text-faint">Instructions</p>
        <div className="text-sm">
          <WorksheetMarkup source={worksheet.instructions} />
        </div>
      </Card>

      <section className="grid max-w-3xl gap-5" aria-label="Worksheet problems">
        {worksheet.problems.map((problem, index) => (
          <article
            key={problem.id}
            className="rounded-xl border border-line bg-panel px-6 py-6 shadow-lg max-sm:px-4"
            aria-labelledby={`worksheet-problem-${problem.id}`}
          >
            <div className="mb-5 flex items-start justify-between gap-4 border-b border-line pb-4">
              <h2
                id={`worksheet-problem-${problem.id}`}
                className="text-lg font-semibold tracking-tight text-ink"
              >
                Problem {index + 1}
              </h2>
              {problem.requiresWork ? <Badge tone="gold">Show your work</Badge> : null}
            </div>
            <div className="text-sm">
              <WorksheetMarkup source={problem.prompt} />
            </div>
          </article>
        ))}
      </section>

      <Link
        className="w-max text-xs font-bold text-muted no-underline hover:text-ink"
        to={lessonPath(courseId, moduleId, lessonId)}
      >
        ← Back to {lessonTitle}
      </Link>
    </div>
  )
}

function WorksheetState({
  courseId,
  moduleId,
  title,
  detail,
}: {
  courseId?: string
  moduleId?: string
  title: string
  detail: string
}) {
  const destination =
    courseId && moduleId
      ? modulePath(courseId, moduleId)
      : courseId
        ? coursePath(courseId)
        : '/curriculum'

  return (
    <div className="grid max-w-5xl gap-7 max-sm:gap-5">
      <Link
        className="justify-self-start text-xs font-bold text-muted no-underline hover:text-ink"
        to={destination}
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
    : 'The worksheet could not be loaded. Check the worksheet link and try again.'
}
