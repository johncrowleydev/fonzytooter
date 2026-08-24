import { evaluate } from '@mdx-js/mdx'
import type { MDXContent } from 'mdx/types'
import { Component, type ReactNode, useEffect, useState } from 'react'
import * as runtime from 'react/jsx-runtime'
import type { VideoResource } from '../../api/generated/schemas/videoResource.zod'
import { LessonVideoCatalogProvider } from './components/YouTubeVideo'
import { lessonMdxComponents } from './mdx/components'

export type LessonMdxProps = {
  source: string
  videos: VideoResource[]
}

type EvaluationState =
  | { source: string; status: 'loading' }
  | { source: string; status: 'ready'; content: MDXContent }
  | { source: string; status: 'error'; error: Error }

const evaluatedSources = new Map<string, Promise<MDXContent>>()

/**
 * Lesson MDX is trusted Git-authored curriculum content. MDX evaluation executes the JavaScript
 * produced by the compiler, so this renderer must never receive user submissions, tutor output,
 * arbitrary remote content, or database-authored untrusted markup.
 */
function evaluateLessonSource(source: string) {
  const cached = evaluatedSources.get(source)
  if (cached) return cached

  const evaluation = evaluate(source, {
    ...runtime,
    useMDXComponents: () => lessonMdxComponents,
  }).then((module) => {
    if (typeof module.default !== 'function') {
      throw new Error('The evaluated lesson did not provide an MDX content component.')
    }
    return module.default
  })

  evaluatedSources.set(source, evaluation)
  return evaluation
}

export function LessonMdx({ source, videos }: LessonMdxProps) {
  const [state, setState] = useState<EvaluationState>({ source, status: 'loading' })

  useEffect(() => {
    let cancelled = false
    setState({ source, status: 'loading' })

    evaluateLessonSource(source).then(
      (content) => {
        if (!cancelled) setState({ source, status: 'ready', content })
      },
      (error: unknown) => {
        if (!cancelled) setState({ source, status: 'error', error: toError(error) })
      },
    )

    return () => {
      cancelled = true
    }
  }, [source])

  const currentState = state.source === source ? state : { source, status: 'loading' as const }

  if (currentState.status === 'loading') {
    return (
      <div className="text-sm text-muted" role="status">
        Loading lesson…
      </div>
    )
  }

  if (currentState.status === 'error') {
    return <LessonMdxError error={currentState.error} />
  }

  return (
    <LessonMdxErrorBoundary key={source}>
      <LessonVideoCatalogProvider videos={videos}>
        <div className="max-w-3xl text-base leading-7 text-body">
          <currentState.content components={lessonMdxComponents} />
        </div>
      </LessonVideoCatalogProvider>
    </LessonMdxErrorBoundary>
  )
}

type LessonMdxErrorBoundaryProps = {
  children: ReactNode
}

type LessonMdxErrorBoundaryState = {
  error: Error | null
}

class LessonMdxErrorBoundary extends Component<
  LessonMdxErrorBoundaryProps,
  LessonMdxErrorBoundaryState
> {
  state: LessonMdxErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: unknown): LessonMdxErrorBoundaryState {
    return { error: toError(error) }
  }

  render() {
    return this.state.error ? <LessonMdxError error={this.state.error} /> : this.props.children
  }
}

function LessonMdxError({ error }: { error: Error }) {
  return (
    <div
      className="rounded-lg border border-accent-coral/40 bg-accent-coral/10 p-4 text-sm text-accent-coral"
      role="alert"
    >
      <strong className="font-semibold">Lesson rendering error</strong>
      <p className="mt-2 break-words leading-6">{error.message}</p>
    </div>
  )
}

function toError(error: unknown) {
  return error instanceof Error ? error : new Error(String(error))
}
