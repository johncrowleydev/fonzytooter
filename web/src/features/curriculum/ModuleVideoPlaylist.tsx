import { useState } from 'react'
import { Link } from 'react-router-dom'
import type { ModuleResource } from '../../api/generated/schemas/moduleResource.zod'
import type { VideoResource } from '../../api/generated/schemas/videoResource.zod'
import { lessonPath } from '../../app/routes'
import { Badge, SectionHeading } from '../../components/ui'
import { YouTubePlayer } from '../videos/YouTubePlayer'
import { youtubeThumbnailUrl, youtubeWatchUrl } from '../videos/youtube'

export function ModuleVideoPlaylist({ module }: { module: ModuleResource }) {
  const orderedVideos = [...module.videos].sort(
    (left, right) => left.order - right.order || left.id.localeCompare(right.id),
  )
  const [selectedVideoId, setSelectedVideoId] = useState(orderedVideos[0]?.id)
  const selectedVideo =
    orderedVideos.find((video) => video.id === selectedVideoId) ?? orderedVideos[0]

  if (!selectedVideo) return null

  return (
    <section aria-label="Module video playlist">
      <SectionHeading
        eyebrow="Watch"
        title="Video playlist"
        detail={`${orderedVideos.length} curated ${orderedVideos.length === 1 ? 'video' : 'videos'} in learning order.`}
      />

      <div className="grid overflow-hidden rounded-xl border border-line bg-panel shadow-lg lg:grid-cols-[minmax(0,3fr)_minmax(18rem,2fr)]">
        <div className="min-w-0 border-b border-line lg:border-r lg:border-b-0">
          <YouTubePlayer video={selectedVideo} />
          <div className="px-5 py-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold tracking-tight text-ink">
                  {selectedVideo.title}
                </h3>
                <p className="mt-1 text-sm text-muted">
                  {selectedVideo.channel} · About {selectedVideo.durationMinutes} min
                </p>
              </div>
              <a
                className="text-sm font-semibold text-accent-teal no-underline hover:text-ink"
                href={youtubeWatchUrl(selectedVideo.youtubeId)}
                target="_blank"
                rel="noopener noreferrer"
              >
                Open on YouTube ↗
              </a>
            </div>
            <VideoAssociations module={module} video={selectedVideo} />
          </div>
        </div>

        <ol className="grid content-start divide-y divide-line" aria-label="Playlist videos">
          {orderedVideos.map((video, index) => {
            const selected = video.id === selectedVideo.id
            return (
              <li key={video.id}>
                <button
                  className={`grid w-full grid-cols-[7rem_minmax(0,1fr)] gap-3 border-0 px-4 py-4 text-left transition pointer-coarse:min-h-11 ${
                    selected
                      ? 'bg-accent-violet/10 text-ink'
                      : 'bg-transparent text-body hover:bg-raised active:bg-accent-slate/20'
                  }`}
                  type="button"
                  aria-pressed={selected}
                  onClick={() => setSelectedVideoId(video.id)}
                >
                  <span className="relative aspect-video overflow-hidden rounded-md bg-code-surface">
                    <img
                      className="h-full w-full object-cover"
                      src={youtubeThumbnailUrl(video.youtubeId)}
                      alt=""
                      loading="lazy"
                    />
                    <span className="absolute right-1 bottom-1 rounded bg-code-surface/90 px-1.5 py-0.5 text-xs font-semibold text-code-ink">
                      {video.durationMinutes} min
                    </span>
                  </span>
                  <span className="min-w-0">
                    <span className="text-xs font-semibold tracking-wide text-faint uppercase">
                      Video {index + 1}
                    </span>
                    <strong className="mt-1 block text-sm leading-snug text-ink">
                      {video.title}
                    </strong>
                    <span className="mt-1 block text-sm text-muted">{video.channel}</span>
                  </span>
                </button>
              </li>
            )
          })}
        </ol>
      </div>
    </section>
  )
}

function VideoAssociations({ module, video }: { module: ModuleResource; video: VideoResource }) {
  const lessons = video.lessonIds.flatMap((lessonId) => {
    const lesson = module.lessons.find((candidate) => candidate.id === lessonId)
    return lesson ? [lesson] : []
  })
  const objectives = video.objectiveIds.flatMap((objectiveId) => {
    const objective = module.objectives.find((candidate) => candidate.id === objectiveId)
    return objective ? [objective] : []
  })

  return (
    <div className="mt-5 grid gap-4 border-t border-line pt-4 sm:grid-cols-2">
      <div>
        <Badge tone="teal">Lesson context</Badge>
        {lessons.length > 0 ? (
          <ul className="mt-2 grid gap-1 text-sm">
            {lessons.map((lesson) => (
              <li key={lesson.id}>
                <Link
                  className="font-semibold text-accent-teal no-underline hover:text-ink"
                  to={lessonPath(module.courseId, module.id, lesson.id)}
                >
                  {lesson.title} →
                </Link>
              </li>
            ))}
          </ul>
        ) : (
          <p className="mt-2 text-sm text-muted">Module-wide resource</p>
        )}
      </div>
      <div>
        <Badge tone="violet">Supports</Badge>
        <ul className="mt-2 grid gap-1 text-sm text-muted">
          {objectives.map((objective) => (
            <li key={objective.id}>{objective.title}</li>
          ))}
        </ul>
      </div>
    </div>
  )
}
