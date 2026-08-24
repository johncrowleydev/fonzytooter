import { createContext, useContext, useId, useMemo, type ReactNode } from 'react'
import type { VideoResource } from '../../../api/generated/schemas/videoResource.zod'
import { YouTubePlayer } from '../../videos/YouTubePlayer'
import { youtubeWatchUrl } from '../../videos/youtube'

const LessonVideoCatalogContext = createContext<ReadonlyMap<string, VideoResource> | null>(null)

export function LessonVideoCatalogProvider({
  children,
  videos,
}: {
  children: ReactNode
  videos: VideoResource[]
}) {
  const catalog = useMemo(() => new Map(videos.map((video) => [video.id, video])), [videos])

  return (
    <LessonVideoCatalogContext.Provider value={catalog}>
      {children}
    </LessonVideoCatalogContext.Provider>
  )
}

export function YouTubeVideo({ id, children }: { id: string; children?: ReactNode }) {
  const catalog = useContext(LessonVideoCatalogContext)
  const titleId = useId()
  if (!catalog) {
    throw new Error('YouTubeVideo must render inside a lesson video catalog.')
  }

  const video = catalog.get(id)
  if (!video) {
    throw new Error(`Lesson references unknown curated video ${JSON.stringify(id)}.`)
  }

  return (
    <section
      className="my-8 overflow-hidden rounded-xl border border-line bg-panel shadow-lg"
      aria-labelledby={titleId}
    >
      <div className="border-b border-line bg-panel-soft px-5 py-4">
        <p className="mb-1 text-xs font-semibold tracking-wide text-accent-violet uppercase">
          Curated video
        </p>
        <h3 id={titleId} className="text-lg font-semibold tracking-tight text-ink">
          {video.title}
        </h3>
        <p className="mt-1 text-sm text-muted">
          {video.channel} · About {video.durationMinutes} min
        </p>
      </div>

      <YouTubePlayer video={video} />

      <div className="px-5 py-4">
        {children ? (
          <div className="rounded-lg border border-accent-violet/40 bg-accent-violet/10 px-4 py-3 text-sm leading-6 text-body">
            <strong className="mb-1 block font-semibold text-ink">What to notice</strong>
            {children}
          </div>
        ) : null}
        <p className="mt-4 mb-0 text-sm text-muted">
          If the embedded player is unavailable,{' '}
          <a
            className="font-semibold text-accent-teal underline decoration-accent-teal/50 underline-offset-4 hover:text-ink"
            href={youtubeWatchUrl(video.youtubeId)}
            target="_blank"
            rel="noopener noreferrer"
          >
            open this video on YouTube ↗
          </a>
          .
        </p>
      </div>
    </section>
  )
}
