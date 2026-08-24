import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { VideoResource } from '../../../api/generated/schemas/videoResource.zod'
import { LessonVideoCatalogProvider, YouTubeVideo } from './YouTubeVideo'

afterEach(cleanup)

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
