import { useQueryClient } from '@tanstack/react-query'
import {
  getGetVideoProgressQueryKey,
  getListActivitiesQueryKey,
  useGetVideoProgress,
  usePutVideoProgress,
} from '../../api/generated/endpoints'
import type { VideoResource } from '../../api/generated/schemas/videoResource.zod'
import { Badge } from '../../components/ui'

export function VideoCompletionControl({
  authenticated,
  video,
}: {
  authenticated: boolean
  video: VideoResource
}) {
  if (!authenticated) return null
  return <AuthenticatedVideoCompletionControl video={video} />
}

function AuthenticatedVideoCompletionControl({ video }: { video: VideoResource }) {
  const progressQuery = useGetVideoProgress(video.courseId, video.moduleId, video.id, {
    query: { enabled: true },
  })
  const queryClient = useQueryClient()
  const updateProgress = usePutVideoProgress({
    mutation: {
      onSuccess: async (response, variables) => {
        queryClient.setQueryData(
          getGetVideoProgressQueryKey(variables.courseId, variables.moduleId, variables.videoId),
          response,
        )
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getGetVideoProgressQueryKey(
              variables.courseId,
              variables.moduleId,
              variables.videoId,
            ),
          }),
          queryClient.invalidateQueries({
            queryKey: getListActivitiesQueryKey({ courseId: variables.courseId, limit: 6 }),
          }),
        ])
      },
    },
  })

  const completed = progressQuery.data?.data.completed ?? false
  const pending = progressQuery.isPending || updateProgress.isPending
  const failed = progressQuery.isError || updateProgress.isError

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4">
      <div>
        <Badge tone={completed ? 'teal' : 'neutral'}>{completed ? 'Watched' : 'Not watched'}</Badge>
        {failed ? (
          <p className="mt-2 text-sm text-accent-coral" role="alert">
            Video status could not be saved. Try again.
          </p>
        ) : (
          <p className="mt-2 text-sm text-muted">
            This records that you watched the video. It does not award mastery.
          </p>
        )}
      </div>
      <button
        type="button"
        disabled={pending || completed}
        onClick={() =>
          updateProgress.mutate({
            courseId: video.courseId,
            moduleId: video.moduleId,
            videoId: video.id,
            data: { completed: true },
          })
        }
        className="shrink-0 rounded-lg border border-line-strong bg-accent-slate/10 px-4 py-2.5 text-sm font-bold text-ink transition hover:bg-accent-slate/20 disabled:cursor-default disabled:opacity-60"
      >
        {pending ? 'Saving…' : completed ? 'Watched' : 'Mark watched'}
      </button>
    </div>
  )
}
