import { useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import {
  getGetCourseProgressQueryKey,
  getGetLessonProgressQueryKey,
  getListActivitiesQueryKey,
  useGetCourse,
  useGetCourseLesson,
  useGetCourseModule,
  useGetLessonProgress,
  usePutLessonProgress,
} from '../../api/generated/endpoints'
import type { LessonResource } from '../../api/generated/schemas/lessonResource.zod'
import { coursePath, lessonPath, modulePath, worksheetPath } from '../../app/routes'
import { Badge, Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'
import { externalHost, safeExternalUrl } from '../curriculum/externalLinks'
import { LessonMdx } from './LessonMdx'
import { LessonCompletionControl } from './LessonCompletionControl'
import { useAuth } from '../authentication/AuthContext'

export function Lesson() {
  const auth = useAuth()
  const { courseId, moduleId, lessonId } = useParams()
  const courseQuery = useGetCourse(courseId ?? '', {
    query: { enabled: Boolean(courseId) },
  })
  const moduleQuery = useGetCourseModule(courseId ?? '', moduleId ?? '', {
    query: { enabled: Boolean(courseId && moduleId) },
  })
  const lessonQuery = useGetCourseLesson(courseId ?? '', moduleId ?? '', lessonId ?? '', {
    query: { enabled: Boolean(courseId && moduleId && lessonId) },
  })
  const { setPageContext, openTutorWithContext } = useTutor()
  const [selectedText, setSelectedText] = useState('')
  const course = courseQuery.data?.data
  const module = moduleQuery.data?.data
  const lesson = lessonQuery.data?.data
  const matchingCourse = course?.id === courseId ? course : undefined
  const matchingModule =
    matchingCourse &&
    module?.courseId === matchingCourse.id &&
    module.id === moduleId &&
    matchingCourse.modules.some((item) => item.id === module.id)
      ? module
      : undefined
  const matchingLesson =
    matchingModule &&
    lesson &&
    lesson.courseId === matchingCourse?.id &&
    lesson.moduleId === matchingModule.id &&
    lesson.id === lessonId
      ? lesson
      : undefined
  const progressQuery = useGetLessonProgress(courseId ?? '', moduleId ?? '', lessonId ?? '', {
    query: { enabled: auth.isAuthenticated && Boolean(matchingLesson) },
  })
  const queryClient = useQueryClient()
  const updateProgress = usePutLessonProgress({
    mutation: {
      onSuccess: async (response, variables) => {
        queryClient.setQueryData(
          getGetLessonProgressQueryKey(variables.courseId, variables.moduleId, variables.lessonId),
          response,
        )
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getGetLessonProgressQueryKey(
              variables.courseId,
              variables.moduleId,
              variables.lessonId,
            ),
          }),
          queryClient.invalidateQueries({
            queryKey: getGetCourseProgressQueryKey(variables.courseId),
          }),
          queryClient.invalidateQueries({
            queryKey: getListActivitiesQueryKey({ courseId: variables.courseId, limit: 6 }),
          }),
        ])
      },
    },
  })
  const lessonIndex =
    module && lesson ? module.lessons.findIndex((item) => item.id === lesson.id) : -1

  useEffect(() => {
    setPageContext({
      type: 'lesson',
      title: matchingLesson?.title ?? matchingModule?.title ?? matchingCourse?.title ?? 'Lesson',
      courseId,
      courseTitle: matchingCourse?.title,
      lessonId,
      lessonTitle: matchingLesson?.title,
      moduleId,
      moduleTitle: matchingModule?.title,
      objectiveIds: matchingLesson?.objectiveIds,
    })
  }, [courseId, lessonId, matchingCourse, matchingLesson, matchingModule, moduleId, setPageContext])

  if (!courseId || !moduleId || !lessonId) {
    return (
      <LessonState
        title="Lesson unavailable"
        detail="This lesson route is missing its course, module, or lesson identity."
      />
    )
  }

  if (courseQuery.isPending || moduleQuery.isPending || lessonQuery.isPending) {
    return (
      <LessonState
        courseId={matchingCourse?.id}
        title="Loading lesson"
        detail="Fetching the lesson and its module context…"
      />
    )
  }

  if (courseQuery.isError || moduleQuery.isError || lessonQuery.isError) {
    return (
      <LessonState
        title="Lesson unavailable"
        courseId={matchingCourse?.id}
        detail={getErrorMessage(courseQuery.error ?? moduleQuery.error ?? lessonQuery.error)}
      />
    )
  }

  if (!course || !module || !lesson) {
    return (
      <LessonState
        courseId={matchingCourse?.id}
        title="Lesson unavailable"
        detail="No lesson data was returned for this route."
      />
    )
  }

  if (
    course.id !== courseId ||
    module.courseId !== course.id ||
    module.id !== moduleId ||
    lesson.courseId !== course.id ||
    lesson.moduleId !== module.id ||
    lesson.id !== lessonId ||
    !course.modules.some((item) => item.id === module.id) ||
    lessonIndex < 0
  ) {
    return (
      <LessonState
        courseId={matchingCourse?.id}
        title="Lesson unavailable"
        detail="This lesson does not belong to the requested module. Use the module page to choose a lesson."
      />
    )
  }

  const previousLesson = module.lessons[lessonIndex - 1]
  const nextLesson = module.lessons[lessonIndex + 1]
  const openSelectionTutor = () => {
    openTutorWithContext({
      type: 'lesson',
      title: lesson.title,
      courseId: course.id,
      courseTitle: course.title,
      lessonId: lesson.id,
      lessonTitle: lesson.title,
      moduleId: module.id,
      moduleTitle: module.title,
      selectedText,
      objectiveIds: lesson.objectiveIds,
    })
  }

  return (
    <div className="grid max-w-none gap-7 max-sm:gap-5">
      <div className="flex items-center gap-2.5 text-sm text-faint max-sm:gap-2">
        <Link
          className="text-muted no-underline hover:text-accent-teal"
          to={modulePath(course.id, module.id)}
        >
          {module.title}
        </Link>
        <span>/</span>
        <span>{lesson.title}</span>
        <span className="ml-auto text-muted">
          {lessonIndex + 1} / {module.lessons.length}
        </span>
      </div>

      <article className="relative" onMouseUp={() => handleSelection(setSelectedText)}>
        <PageIntro
          compact
          eyebrow={`Lesson ${String(lessonIndex + 1).padStart(2, '0')} · ${module.title}`}
          title={lesson.title}
          detail={`${lesson.objectiveIds.length} ${lesson.objectiveIds.length === 1 ? 'objective' : 'objectives'}`}
        />

        {selectedText ? (
          <SelectionPopover onAskTutor={openSelectionTutor} onDismiss={() => setSelectedText('')} />
        ) : null}

        <div className="mx-auto mt-12 max-w-3xl text-base leading-loose text-body max-sm:mt-9 max-sm:text-sm max-sm:leading-8">
          <LessonMdx source={lesson.content} />
          <LessonSources sources={lesson.sources} />
          <LessonWorksheets
            courseId={course.id}
            moduleId={module.id}
            worksheets={lesson.worksheets}
          />
          <LessonCompletionControl
            authenticated={auth.isAuthenticated}
            completed={progressQuery.data?.data.completed ?? false}
            pending={auth.isAuthenticated && (progressQuery.isPending || updateProgress.isPending)}
            error={
              progressQuery.isError || updateProgress.isError
                ? 'Lesson progress could not be saved. Try again.'
                : undefined
            }
            onChange={(completed) =>
              updateProgress.mutate({
                courseId: course.id,
                moduleId: module.id,
                lessonId: lesson.id,
                data: { completed },
              })
            }
          />
        </div>
      </article>

      <nav
        className="mx-auto flex w-full max-w-3xl items-center justify-between gap-3 border-t border-line pt-5 max-sm:gap-2"
        aria-label="Lesson navigation"
      >
        {previousLesson ? (
          <Link
            className="inline-flex max-w-[45%] items-center gap-2.5 rounded-lg border border-line-strong bg-accent-slate/10 px-4 py-2.5 text-sm font-bold text-ink no-underline transition hover:bg-accent-slate/20"
            to={lessonPath(course.id, module.id, previousLesson.id)}
          >
            ← <span className="truncate">{previousLesson.title}</span>
          </Link>
        ) : (
          <span className="text-sm text-faint">Beginning of module</span>
        )}
        <span className="text-sm text-faint max-sm:hidden">
          Lesson {lessonIndex + 1} of {module.lessons.length}
        </span>
        {nextLesson ? (
          <Link
            className="inline-flex max-w-[45%] items-center gap-2.5 rounded-lg bg-brand-teal px-4 py-2.5 text-sm font-bold text-brand-ink no-underline transition hover:-translate-y-px hover:bg-brand-teal-light"
            to={lessonPath(course.id, module.id, nextLesson.id)}
          >
            <span className="truncate">{nextLesson.title}</span> →
          </Link>
        ) : (
          <span className="text-right text-sm text-faint">End of module</span>
        )}
      </nav>
    </div>
  )
}

function LessonWorksheets({
  courseId,
  moduleId,
  worksheets,
}: {
  courseId: string
  moduleId: string
  worksheets: LessonResource['worksheets']
}) {
  if (worksheets.length === 0) return null

  return (
    <section className="mt-11 border-t border-line pt-7">
      <SectionHeading
        eyebrow="Practice"
        title="Worksheets"
        detail="Reinforce this lesson with focused written practice."
      />
      <div className="grid gap-3">
        {worksheets.map((worksheet) => (
          <Link
            key={worksheet.id}
            className="flex items-center gap-4 rounded-xl border border-line bg-panel px-5 py-4 text-ink no-underline shadow-lg transition hover:border-line-strong hover:text-accent-teal"
            to={worksheetPath(courseId, moduleId, worksheet.id)}
          >
            <Badge tone="teal">Worksheet</Badge>
            <span className="min-w-0 flex-1">
              <strong className="block text-sm">{worksheet.title}</strong>
              <small className="mt-1 block text-sm text-faint">
                {worksheet.problemCount} {worksheet.problemCount === 1 ? 'problem' : 'problems'}
              </small>
            </span>
            <span className="text-base text-faint" aria-hidden="true">
              →
            </span>
          </Link>
        ))}
      </div>
    </section>
  )
}

function handleSelection(setSelectedText: (value: string) => void) {
  const selection = window.getSelection()?.toString().trim() ?? ''
  if (selection.length > 8) setSelectedText(selection)
}

function SelectionPopover({
  onAskTutor,
  onDismiss,
}: {
  onAskTutor: () => void
  onDismiss: () => void
}) {
  return (
    <div className="absolute top-40 right-5 z-4 flex items-center gap-2 rounded-lg border border-line-strong bg-panel px-2 py-2 text-sm shadow-2xl max-sm:top-36 max-sm:left-2.5 max-sm:right-2.5">
      <span className="text-faint">Text selected</span>
      <button
        className="border-0 bg-transparent p-0 text-sm font-bold text-accent-teal"
        type="button"
        onClick={onAskTutor}
      >
        Ask tutor about this ↗
      </button>
      <button
        className="border-0 bg-transparent p-0 text-sm text-faint"
        type="button"
        onClick={onDismiss}
        aria-label="Dismiss selection"
      >
        ×
      </button>
    </div>
  )
}

function LessonSources({ sources }: { sources: LessonResource['sources'] }) {
  return (
    <section className="mt-11 border-t border-line pt-7">
      <SectionHeading title="Sources" />
      {sources.length > 0 ? (
        <div className="grid">
          {sources.map((source, index) => {
            const url = safeExternalUrl(source.url)
            const host = externalHost(source.url)
            return (
              <div
                key={source.id}
                className="grid grid-cols-[27px_minmax(0,1fr)_auto] gap-2.5 border-t border-line py-3"
              >
                <span className="text-sm text-faint">[{String(index + 1).padStart(2, '0')}]</span>
                <div>
                  {url ? (
                    <a
                      className="text-sm font-bold text-accent-teal no-underline hover:text-ink"
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {source.title} ↗
                    </a>
                  ) : (
                    <strong className="block text-sm">{source.title}</strong>
                  )}
                  {host ? <span className="mt-1 block text-sm text-faint">{host}</span> : null}
                </div>
                <span className="text-sm text-faint">{url ? 'External' : 'Link unavailable'}</span>
              </div>
            )
          })}
        </div>
      ) : (
        <Card muted>
          <p className="text-sm leading-relaxed text-muted">
            No sources are recorded for this lesson.
          </p>
        </Card>
      )}
    </section>
  )
}

function LessonState({
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
        className="justify-self-start text-sm font-bold text-muted no-underline hover:text-ink"
        to={courseId ? coursePath(courseId) : '/curriculum'}
      >
        ← Curriculum
      </Link>
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-muted">{detail}</p>
      </Card>
    </div>
  )
}

function getErrorMessage(error: unknown) {
  return error instanceof Error && error.message
    ? error.message
    : 'The lesson could not be loaded. Check the lesson link and try again.'
}
