import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { VideoResource } from '../../../api/generated/schemas/videoResource.zod'
import { LessonVideoCatalogProvider, YouTubeVideo } from './YouTubeVideo'

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

const video: VideoResource = {
  channel: 'Visual Teacher',
  courseId: 'course',
  durationMinutes: 12,
  id: 'shape-lesson',
  lessonIds: ['lesson'],
  moduleId: 'module',
  objectiveIds: ['objective'],
  order: 0,
  title: 'See shapes expand',
  youtubeId: 'dQw4w9WgXcQ',
}

describe('YouTubeVideo', () => {
  it('renders canonical metadata, contextual guidance, and a privacy-enhanced player', () => {
    render(
      <LessonVideoCatalogProvider videos={[video]}>
        <YouTubeVideo id="shape-lesson">
          <p>Notice the singleton dimension.</p>
        </YouTubeVideo>
      </LessonVideoCatalogProvider>,
    )

    expect(screen.getByRole('heading', { name: video.title })).toBeTruthy()
    expect(screen.getByText('Visual Teacher · About 12 min')).toBeTruthy()
    expect(screen.getByText('Notice the singleton dimension.')).toBeTruthy()

    const frame = screen.getByTitle('See shapes expand by Visual Teacher') as HTMLIFrameElement
    const embedUrl = new URL(frame.src)
    expect(embedUrl.hostname).toBe('www.youtube-nocookie.com')
    expect(embedUrl.pathname).toBe('/embed/dQw4w9WgXcQ')
    expect(embedUrl.searchParams.has('autoplay')).toBe(false)
    expect(frame.getAttribute('loading')).toBe('lazy')

    expect(
      screen.getByRole('link', { name: /open this video on youtube/i }).getAttribute('href'),
    ).toBe('https://www.youtube.com/watch?v=dQw4w9WgXcQ')
    expect(screen.queryByRole('button', { name: 'Mark watched' })).toBeNull()
  })

  it('shows authenticated watched state beside the same lesson embed', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            completed: false,
            courseId: 'course',
            moduleId: 'module',
            videoId: 'shape-lesson',
          }),
        ),
      ),
    )
    render(
      <QueryClientProvider client={new QueryClient()}>
        <LessonVideoCatalogProvider videos={[video]} authenticated>
          <YouTubeVideo id="shape-lesson" />
        </LessonVideoCatalogProvider>
      </QueryClientProvider>,
    )

    await waitFor(() => expect(screen.getByRole('button', { name: 'Mark watched' })).toBeTruthy())
  })

  it('fails clearly when authored content references an unknown video', () => {
    expect(() =>
      render(
        <LessonVideoCatalogProvider videos={[video]}>
          <YouTubeVideo id="missing" />
        </LessonVideoCatalogProvider>,
      ),
    ).toThrow('Lesson references unknown curated video "missing".')
  })
})
