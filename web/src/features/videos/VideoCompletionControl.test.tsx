import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { VideoResource } from '../../api/generated/schemas/videoResource.zod'
import { VideoCompletionControl } from './VideoCompletionControl'

const mocks = vi.hoisted(() => ({
  mutate: vi.fn(),
  progress: vi.fn(),
}))

vi.mock('../../api/generated/endpoints', () => ({
  getGetVideoProgressQueryKey: (courseId: string, moduleId: string, videoId: string) => [
    'video-progress',
    courseId,
    moduleId,
    videoId,
  ],
  getListActivitiesQueryKey: () => ['activities'],
  getListVideoRecommendationsQueryKey: () => ['video-recommendations'],
  useGetVideoProgress: (...args: unknown[]) => mocks.progress(...args),
  usePutVideoProgress: () => ({ isError: false, isPending: false, mutate: mocks.mutate }),
}))

const video: VideoResource = {
  channel: 'Creator',
  courseId: 'course',
  durationMinutes: 5,
  id: 'video',
  lessonIds: ['lesson'],
  moduleId: 'module',
  objectiveIds: ['objective'],
  order: 0,
  title: 'Video title',
  youtubeId: 'dQw4w9WgXcQ',
}

function renderControl(authenticated: boolean) {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <VideoCompletionControl authenticated={authenticated} video={video} />
    </QueryClientProvider>,
  )
}

describe('VideoCompletionControl', () => {
  beforeEach(() => {
    mocks.mutate.mockReset()
    mocks.progress.mockReset().mockReturnValue({
      data: { data: { completed: false } },
      isError: false,
      isPending: false,
    })
  })
  afterEach(cleanup)

  it('does not read or expose private learner state for a public viewer', () => {
    const { container } = renderControl(false)
    expect(container.textContent).toBe('')
    expect(mocks.progress).not.toHaveBeenCalled()
  })

  it('uses an explicit manual completion action without claiming mastery', () => {
    renderControl(true)
    expect(screen.getByText('Not watched')).toBeTruthy()
    expect(screen.getByText(/does not award mastery/i)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Mark watched' }))
    expect(mocks.mutate).toHaveBeenCalledWith({
      courseId: 'course',
      data: { completed: true },
      moduleId: 'module',
      videoId: 'video',
    })
  })

  it('shows the same completed state as watched and prevents duplicate writes', () => {
    mocks.progress.mockReturnValue({
      data: { data: { completed: true } },
      isError: false,
      isPending: false,
    })
    renderControl(true)
    const watched = screen.getByRole('button', { name: 'Watched' })
    expect(watched.hasAttribute('disabled')).toBe(true)
    fireEvent.click(watched)
    expect(mocks.mutate).not.toHaveBeenCalled()
  })
})
