import type { VideoResource } from '../../api/generated/schemas/videoResource.zod'
import { youtubeEmbedUrl } from './youtube'

export function YouTubePlayer({ video }: { video: VideoResource }) {
  return (
    <div className="aspect-video w-full bg-code-surface">
      <iframe
        className="h-full w-full border-0"
        src={youtubeEmbedUrl(video.youtubeId)}
        title={`${video.title} by ${video.channel}`}
        loading="lazy"
        referrerPolicy="strict-origin-when-cross-origin"
        allow="encrypted-media; picture-in-picture; web-share"
        allowFullScreen
      />
    </div>
  )
}
