import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'
import type { VideoResource } from '../../api/generated/schemas/videoResource.zod'
import { LessonMdx } from './LessonMdx'

afterEach(cleanup)

const videos: VideoResource[] = [
  {
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
  },
]

it('resolves a trusted MDX YouTubeVideo without authenticated state', async () => {
  render(
    <LessonMdx
      source={'<YouTubeVideo id="shape-lesson">Watch the singleton axis.</YouTubeVideo>'}
      videos={videos}
    />,
  )

  expect(await screen.findByTitle('See shapes expand by Visual Teacher')).toBeTruthy()
  expect(screen.getByText('Watch the singleton axis.')).toBeTruthy()
})

it('surfaces an unresolved MDX video ID as a clear lesson error', async () => {
  render(<LessonMdx source={'<YouTubeVideo id="missing" />'} videos={videos} />)

  expect((await screen.findByRole('alert')).textContent).toContain(
    'Lesson references unknown curated video "missing".',
  )
})
