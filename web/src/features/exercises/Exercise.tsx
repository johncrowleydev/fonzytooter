import { useEffect, useMemo, useRef, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { Link, useParams } from 'react-router-dom'
import {
  useCreateExerciseAttempt,
  useGetCourse,
  useGetCourseLesson,
  useGetCourseModule,
  useGetCourseModuleExercise,
  useGetExerciseCheckDefinition,
  useGetExerciseWorkspace,
  usePutExerciseWorkspace,
} from '../../api/generated/endpoints'
import { lessonPath, modulePath } from '../../app/routes'
import { Badge, Button, Card, PageIntro, SectionHeading } from '../../components/ui'
import { HighlightedCode } from '../../components/HighlightedCode'
import { useTutor } from '../tutor/TutorContext'
import { useAuth } from '../authentication/AuthContext'
import { SignInLink } from '../authentication/SignInLink'
import { CodeEditor } from './CodeEditor'
import { LatestTaskQueue } from './LatestTaskQueue'
import { PyodideRunner } from './runtime/PyodideRunner'
import type { PythonCheckResult, PythonRunResult } from './types'

type SaveState = 'saved' | 'saving' | 'failed' | 'anonymous'

type LocalDraft = {
  code: string
  savedCode: string
}

function localDraftKey(courseId: string, moduleId: string, exerciseId: string) {
  return `helix-academy:exercise:${courseId}:${moduleId}:${exerciseId}`
}

function readLocalDraft(key: string): LocalDraft | undefined {
  try {
    const parsed = JSON.parse(localStorage.getItem(key) ?? 'null') as LocalDraft | null
    return parsed && typeof parsed.code === 'string' && typeof parsed.savedCode === 'string'
      ? parsed
      : undefined
  } catch {
    return undefined
  }
}

export function Exercise() {
  const auth = useAuth()
  const { courseId = '', moduleId = '', exerciseId = '' } = useParams()
  const { setPageContext, openTutorWithContext } = useTutor()
  const runner = useMemo(() => new PyodideRunner(), [])
  const exerciseQuery = useGetCourseModuleExercise(courseId, moduleId, exerciseId)
  const workspaceQuery = useGetExerciseWorkspace(courseId, moduleId, exerciseId, {
    query: { enabled: auth.isAuthenticated },
  })
  const checkDefinitionQuery = useGetExerciseCheckDefinition(courseId, moduleId, exerciseId, {
    query: { enabled: false },
  })
  const courseQuery = useGetCourse(courseId)
  const moduleQuery = useGetCourseModule(courseId, moduleId)
  const lessonId = exerciseQuery.data?.data.lessonId ?? ''
  const lessonQuery = useGetCourseLesson(courseId, moduleId, lessonId, {
    query: { enabled: lessonId.length > 0 },
  })
  const [code, setCode] = useState('')
  const [saveState, setSaveState] = useState<SaveState>('saved')
  const [runResult, setRunResult] = useState<PythonRunResult>()
  const [checkResult, setCheckResult] = useState<PythonCheckResult>()
  const [executionError, setExecutionError] = useState<string>()
  const [activeTab, setActiveTab] = useState<'prompt' | 'tests'>('prompt')
  const [executing, setExecuting] = useState(false)
  const initialized = useRef(false)
  const skipNextSave = useRef(false)
  const savedCode = useRef('')
  const currentCode = useRef(code)
  currentCode.current = code
  const draftKey = localDraftKey(courseId, moduleId, exerciseId)
  const currentDraftKey = useRef(draftKey)
  currentDraftKey.current = draftKey
  const saveQueue = useMemo(() => new LatestTaskQueue(), [])

  const saveWorkspace = usePutExerciseWorkspace()
  const saveWorkspaceRef = useRef(saveWorkspace.mutateAsync)
  saveWorkspaceRef.current = saveWorkspace.mutateAsync
  const createAttempt = useCreateExerciseAttempt()

  function queueWorkspaceSave(codeToSave: string) {
    if (!auth.isAuthenticated) return
    saveQueue.enqueue(async () => {
      try {
        const response = await saveWorkspaceRef.current({
          courseId,
          moduleId,
          exerciseId,
          data: { code: codeToSave },
        })
        if (currentDraftKey.current !== draftKey) return
        savedCode.current = response.data.code
        localStorage.setItem(
          draftKey,
          JSON.stringify({ code: currentCode.current, savedCode: response.data.code }),
        )
        setSaveState(currentCode.current === response.data.code ? 'saved' : 'saving')
      } catch {
        if (currentDraftKey.current === draftKey) setSaveState('failed')
      }
    })
  }

  useEffect(() => {
    initialized.current = false
    setCode('')
    setRunResult(undefined)
    setCheckResult(undefined)
  }, [draftKey])

  useEffect(() => {
    if (auth.isPending || auth.isAuthenticated || initialized.current || !exerciseQuery.data) return
    const starterCode = exerciseQuery.data.data.starterCode
    savedCode.current = starterCode
    skipNextSave.current = true
    setCode(starterCode)
    setSaveState('anonymous')
    initialized.current = true
  }, [auth.isAuthenticated, auth.isPending, exerciseQuery.data])

  useEffect(() => {
    if (!auth.isAuthenticated || initialized.current || !workspaceQuery.data) return
    const serverCode = workspaceQuery.data.data.code
    const local = readLocalDraft(draftKey)
    const recoveredCode = local?.savedCode === serverCode ? local.code : serverCode
    savedCode.current = serverCode
    skipNextSave.current = true
    setCode(recoveredCode)
    setSaveState(recoveredCode === serverCode ? 'saved' : 'saving')
    initialized.current = true
  }, [auth.isAuthenticated, draftKey, workspaceQuery.data])

  useEffect(() => {
    if (
      !auth.isAuthenticated ||
      initialized.current ||
      !workspaceQuery.isError ||
      !exerciseQuery.data
    )
      return
    const local = readLocalDraft(draftKey)
    const fallback = local?.code ?? exerciseQuery.data.data.starterCode
    savedCode.current = local?.savedCode ?? fallback
    skipNextSave.current = true
    setCode(fallback)
    setSaveState('failed')
    initialized.current = true
  }, [auth.isAuthenticated, draftKey, exerciseQuery.data, workspaceQuery.isError])

  useEffect(() => {
    if (skipNextSave.current) {
      skipNextSave.current = false
      return
    }
    if (!auth.isAuthenticated || !initialized.current || code === savedCode.current) return
    setSaveState('saving')
    localStorage.setItem(draftKey, JSON.stringify({ code, savedCode: savedCode.current }))
    const timer = window.setTimeout(() => {
      queueWorkspaceSave(code)
    }, 700)
    return () => window.clearTimeout(timer)
  }, [auth.isAuthenticated, code, courseId, draftKey, exerciseId, moduleId])

  useEffect(() => () => runner.dispose(), [runner])

  const exercise = exerciseQuery.data?.data
  const course = courseQuery.data?.data
  const module = moduleQuery.data?.data
  const lesson = lessonQuery.data?.data
  const objectives =
    module?.objectives.filter((objective) => exercise?.objectiveIds.includes(objective.id)) ?? []
  const execution = useMemo(
    () =>
      checkResult
        ? {
            passed: checkResult.tests.filter((test) => test.status === 'passed').length,
            failed: checkResult.tests.filter((test) => test.status !== 'passed').length,
            summary: checkResult.tests.every((test) => test.status === 'passed')
              ? 'All authored checks passed.'
              : 'One or more authored checks need attention.',
          }
        : undefined,
    [checkResult],
  )

  useEffect(() => {
    setPageContext({
      type: 'exercise',
      title: exercise?.title ?? 'Exercise',
      courseId,
      courseTitle: course?.title,
      moduleId,
      moduleTitle: module?.title,
      lessonId: exercise?.lessonId,
      lessonTitle: lesson?.title,
      exerciseId,
      exerciseTitle: exercise?.title,
      objectiveIds: exercise?.objectiveIds,
      code,
      lastExecution: execution,
    })
  }, [
    code,
    course,
    courseId,
    exercise,
    exerciseId,
    execution,
    lesson,
    module,
    moduleId,
    setPageContext,
  ])

  async function run() {
    setExecuting(true)
    setExecutionError(undefined)
    setCheckResult(undefined)
    try {
      setRunResult(await runner.run({ code }))
    } catch (error) {
      setExecutionError(error instanceof Error ? error.message : String(error))
    } finally {
      setExecuting(false)
    }
  }

  async function check() {
    if (!exercise) return
    setExecuting(true)
    setExecutionError(undefined)
    setRunResult(undefined)
    try {
      const tests = auth.isAuthenticated
        ? (await checkDefinitionQuery.refetch()).data?.data.tests
        : exercise.visibleTests.map((test) => ({ ...test, visibility: 'visible' as const }))
      if (!tests) throw new Error('Exercise checks are unavailable')
      const result = await runner.check({ code, tests })
      setCheckResult(result)
      if (auth.isAuthenticated) {
        await createAttempt.mutateAsync({
          courseId,
          moduleId,
          exerciseId,
          data: {
            codeSnapshot: code,
            durationMs: result.durationMs,
            results: result.tests.map((test) => ({
              testId: test.testId,
              status: test.status,
              message: test.message,
              durationMs: test.durationMs,
            })),
          },
        })
      }
    } catch (error) {
      setExecutionError(error instanceof Error ? error.message : String(error))
    } finally {
      setExecuting(false)
    }
  }

  if (
    exerciseQuery.isLoading ||
    auth.isPending ||
    (auth.isAuthenticated && workspaceQuery.isLoading)
  ) {
    return <Card className="p-8 text-sm text-muted">Loading exercise workspace…</Card>
  }
  if (!exercise) {
    return <Card className="p-8 text-sm text-accent-coral">Exercise not found.</Card>
  }

  const output = checkResult ?? runResult
  const passed = checkResult?.tests.filter((test) => test.status === 'passed').length ?? 0
  const failed = checkResult?.tests.filter((test) => test.status !== 'passed').length ?? 0
  const saveLabels: Record<SaveState, string> = {
    saved: 'Saved',
    saving: 'Saving…',
    failed: 'Save failed · local draft kept',
    anonymous: 'Browser-only · not saved',
  }

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <div className="flex flex-wrap gap-2 text-sm text-muted">
        <Link className="font-bold no-underline hover:text-ink" to={modulePath(courseId, moduleId)}>
          ← {module?.title ?? moduleId}
        </Link>
        {lesson ? (
          <Link
            className="no-underline hover:text-ink"
            to={lessonPath(courseId, moduleId, lesson.id)}
          >
            / {lesson.title}
          </Link>
        ) : null}
      </div>
      <PageIntro
        compact
        eyebrow={`${course?.title ?? courseId} · Exercise`}
        title={exercise.title}
      />
      <div className="grid grid-cols-4 gap-5 max-xl:grid-cols-1">
        <main className="col-span-3 grid gap-3.5 max-xl:col-span-1">
          <Card className="overflow-hidden p-0">
            <div className="flex border-b border-line px-5">
              {(['prompt', 'tests'] as const).map((tab) => (
                <button
                  className={`mr-4 border-0 border-b-2 bg-transparent px-2 py-3 text-sm ${activeTab === tab ? 'border-accent-coral text-ink' : 'border-transparent text-faint'}`}
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  type="button"
                >
                  {tab === 'prompt' ? 'Prompt' : 'Visible tests'}
                </button>
              ))}
            </div>
            <div className="prose prose-sm max-w-none px-6 py-5 text-muted">
              {activeTab === 'prompt' ? (
                <ReactMarkdown>{exercise.prompt}</ReactMarkdown>
              ) : (
                <div className="grid gap-4">
                  {exercise.visibleTests.map((test) => (
                    <div key={test.id}>
                      <p className="mb-2 font-semibold text-ink">{test.title}</p>
                      <pre className="overflow-x-auto overscroll-x-contain rounded-lg bg-code-surface p-3 text-sm text-code-ink">
                        <code>
                          {/* Exercise tests are Python by construction: Pyodide is the only
                              interpreter in the learning app. */}
                          <HighlightedCode code={test.code} language="python" />
                        </code>
                      </pre>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </Card>
          <Card className="overflow-hidden p-0">
            <div className="border-b border-line px-5 py-3.5">
              <div>
                <p className="mb-1 font-mono text-sm text-ink">workspace.py</p>
                <span
                  className={
                    saveState === 'failed' ? 'text-sm text-accent-coral' : 'text-sm text-faint'
                  }
                >
                  {saveLabels[saveState]}
                </span>
              </div>
            </div>
            <CodeEditor key={draftKey} disabled={executing} onChange={setCode} value={code} />
            <div className="flex items-center justify-between gap-4 border-t border-line px-4 py-3 max-sm:flex-col max-sm:items-start">
              <div className="flex gap-2">
                <Button disabled={executing} onClick={run}>
                  Run ▶
                </Button>
                <Button disabled={executing} onClick={check} variant="secondary">
                  {auth.isAuthenticated ? 'Check & save ✓' : 'Check visible tests ✓'}
                </Button>
              </div>
              {!auth.isAuthenticated ? (
                <SignInLink className="text-sm font-bold text-accent-teal no-underline hover:text-ink">
                  Sign in to save attempts
                </SignInLink>
              ) : null}
              <button
                className="border-0 bg-transparent p-0 text-sm text-accent-gold"
                onClick={() =>
                  openTutorWithContext({
                    type: 'exercise',
                    title: exercise.title,
                    courseId,
                    courseTitle: course?.title,
                    moduleId,
                    moduleTitle: module?.title,
                    lessonId: exercise.lessonId,
                    lessonTitle: lesson?.title,
                    exerciseId,
                    exerciseTitle: exercise.title,
                    objectiveIds: exercise.objectiveIds,
                    code,
                    lastExecution: execution,
                  })
                }
                type="button"
              >
                ✦ Ask tutor
              </button>
            </div>
          </Card>
          <Card className="overflow-hidden p-0">
            <div className="flex items-center justify-between gap-3 px-5 pt-4">
              <SectionHeading
                eyebrow="Feedback"
                title="Output"
                action={
                  checkResult ? (
                    <Badge tone={failed ? 'gold' : 'teal'}>
                      {passed} passed · {failed} failed
                    </Badge>
                  ) : null
                }
              />
            </div>
            <div className="min-h-44 px-5 pb-5 pt-3 text-sm leading-relaxed">
              {executionError ? <p className="text-accent-coral">{executionError}</p> : null}
              {output?.stdout ? (
                <pre className="whitespace-pre-wrap font-mono text-muted">{output.stdout}</pre>
              ) : null}
              {output?.stderr ? (
                <pre className="whitespace-pre-wrap font-mono text-accent-coral">
                  {output.stderr}
                </pre>
              ) : null}
              {runResult?.error ? (
                <p className="text-accent-coral">
                  {runResult.error.name}: {runResult.error.message}
                </p>
              ) : null}
              {checkResult ? (
                <div className="grid gap-2">
                  {checkResult.tests.map((test) => (
                    <div
                      className="flex items-start gap-3 rounded-lg border border-line p-3"
                      key={test.testId}
                    >
                      <span
                        className={
                          test.status === 'passed' ? 'text-accent-teal' : 'text-accent-coral'
                        }
                      >
                        {test.status === 'passed' ? '✓' : '×'}
                      </span>
                      <div>
                        <p className="font-semibold text-ink">
                          {test.title}
                          {test.visibility === 'hidden' ? ' · hidden check' : ''}
                        </p>
                        {test.message ? <p className="mt-1 text-muted">{test.message}</p> : null}
                      </div>
                    </div>
                  ))}
                </div>
              ) : !output && !executionError ? (
                <div className="grid min-h-32 place-items-center content-center text-center text-faint">
                  Run or check your code to see output.
                </div>
              ) : null}
            </div>
          </Card>
        </main>
        <aside className="grid content-start gap-3.5 max-xl:grid-cols-2 max-sm:grid-cols-1">
          <Card>
            <p className="text-xs font-bold uppercase tracking-widest text-faint">
              Learning objectives
            </p>
            <ul className="mt-3 grid gap-2 text-sm text-muted">
              {objectives.map((objective) => (
                <li key={objective.id}>{objective.title}</li>
              ))}
            </ul>
          </Card>
        </aside>
      </div>
    </div>
  )
}
