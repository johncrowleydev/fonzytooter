import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useGetLesson, useGetModule } from '../../api/generated/endpoints'
import type { LessonResource } from '../../api/generated/schemas/lessonResource.zod'
import { Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'
import { safeExternalUrl } from '../curriculum/externalLinks'
import { LessonMdx } from './LessonMdx'

export function Lesson() {
  const { moduleId, lessonId } = useParams()
  const moduleQuery = useGetModule(moduleId ?? '', {
    query: { enabled: Boolean(moduleId) },
  })
  const lessonQuery = useGetLesson(moduleId ?? '', lessonId ?? '', {
    query: { enabled: Boolean(moduleId && lessonId) },
  })
  const { setPageContext, openTutorWithContext } = useTutor()
  const [selectedText, setSelectedText] = useState('')
  const module = moduleQuery.data?.data
  const lesson = lessonQuery.data?.data
  const lessonIndex =
    module && lesson ? module.lessons.findIndex((item) => item.id === lesson.id) : -1

  useEffect(() => {
    if (!module || !lesson) return

    setPageContext({
      type: 'lesson',
      title: lesson.title,
      lessonId: lesson.id,
      lessonTitle: lesson.title,
      moduleId: module.id,
      moduleTitle: module.title,
      objectiveIds: lesson.objectiveIds,
    })
  }, [lesson, module, setPageContext])

  if (!moduleId || !lessonId) {
    return (
      <LessonState
        title="Lesson unavailable"
        detail="This lesson route is missing its module or lesson identity."
      />
    )
  }

  if (moduleQuery.isPending || lessonQuery.isPending) {
    return (
      <LessonState title="Loading lesson" detail="Fetching the lesson and its module context…" />
    )
  }

  if (moduleQuery.isError || lessonQuery.isError) {
    return (
      <LessonState
        title="Lesson unavailable"
        detail={getErrorMessage(moduleQuery.error ?? lessonQuery.error)}
      />
    )
  }

  if (!module || !lesson) {
    return (
      <LessonState
        title="Lesson unavailable"
        detail="No lesson data was returned for this route."
      />
    )
  }

  if (lesson.moduleId !== module.id || lessonIndex < 0) {
    return (
      <LessonState
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
      <div className="flex items-center gap-2.5 text-2xs text-faint max-sm:gap-2">
        <Link
          className="text-muted no-underline hover:text-brand-teal"
          to={`/curriculum/${module.id}`}
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
        </div>
      </article>

      <nav
        className="mx-auto flex w-full max-w-3xl items-center justify-between gap-3 border-t border-line pt-5 max-sm:gap-2"
        aria-label="Lesson navigation"
      >
        {previousLesson ? (
          <Link
            className="inline-flex max-w-[45%] items-center gap-2.5 rounded-lg border border-line-strong bg-brand-slate/10 px-4 py-2.5 text-xs font-bold text-ink no-underline transition hover:bg-brand-slate/20"
            to={lessonPath(module.id, previousLesson.id)}
          >
            ← <span className="truncate">{previousLesson.title}</span>
          </Link>
        ) : (
          <span className="text-xs text-faint">Beginning of module</span>
        )}
        <span className="text-2xs text-faint max-sm:hidden">
          Lesson {lessonIndex + 1} of {module.lessons.length}
        </span>
        {nextLesson ? (
          <Link
            className="inline-flex max-w-[45%] items-center gap-2.5 rounded-lg bg-brand-teal px-4 py-2.5 text-xs font-bold text-brand-ink no-underline transition hover:-translate-y-px hover:bg-brand-teal-light"
            to={lessonPath(module.id, nextLesson.id)}
          >
            <span className="truncate">{nextLesson.title}</span> →
          </Link>
        ) : (
          <span className="text-right text-xs text-faint">End of module</span>
        )}
      </nav>
    </div>
  )
}

function handleSelection(setSelectedText: (value: string) => void) {
  const selection = window.getSelection()?.toString().trim() ?? ''
  if (selection.length > 8) setSelectedText(selection)
}

function lessonPath(moduleId: string, lessonId: string) {
  return `/curriculum/${moduleId}/lessons/${lessonId}`
}

function SelectionPopover({
  onAskTutor,
  onDismiss,
}: {
  onAskTutor: () => void
  onDismiss: () => void
}) {
  return (
    <div className="absolute top-40 right-5 z-4 flex items-center gap-2 rounded-lg border border-line-strong bg-slate-800 px-2 py-2 text-2xs shadow-2xl max-sm:top-36 max-sm:left-2.5 max-sm:right-2.5">
      <span className="text-faint">Text selected</span>
      <button
        className="border-0 bg-transparent p-0 text-2xs font-bold text-brand-teal"
        type="button"
        onClick={onAskTutor}
      >
        Ask tutor about this ↗
      </button>
      <button
        className="border-0 bg-transparent p-0 text-2xs text-faint"
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
            return (
              <div
                key={source.id}
                className="grid grid-cols-[27px_minmax(0,1fr)_auto] gap-2.5 border-t border-line py-3"
              >
                <span className="text-2xs text-faint">[{String(index + 1).padStart(2, '0')}]</span>
                <div>
                  {url ? (
                    <a
                      className="text-xs font-bold text-brand-teal no-underline hover:text-ink"
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {source.title} ↗
                    </a>
                  ) : (
                    <strong className="block text-xs">{source.title}</strong>
                  )}
                  <span className="mt-1 block text-2xs text-faint">{source.id}</span>
                </div>
                <span className="text-2xs text-faint">{url ? 'External' : 'Link unavailable'}</span>
              </div>
            )
          })}
        </div>
      ) : (
        <Card muted>
          <p className="text-xs leading-relaxed text-muted">
            No sources are recorded for this lesson.
          </p>
        </Card>
      )}
    </section>
  )
}

function LessonState({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <Link
        className="justify-self-start text-xs font-bold text-muted no-underline hover:text-ink"
        to="/curriculum"
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
    : 'The lesson could not be loaded. Check the lesson link and try again.'
}
