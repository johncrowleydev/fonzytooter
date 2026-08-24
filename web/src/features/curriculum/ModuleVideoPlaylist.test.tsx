import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it } from 'vitest'
import type { ModuleResource } from '../../api/generated/schemas/moduleResource.zod'
import { ModuleVideoPlaylist } from './ModuleVideoPlaylist'

afterEach(cleanup)

const moduleWithVideos: ModuleResource = {
  courseId: 'course',
  exercises: [],
  id: 'module',
  lessons: [
    { id: 'lesson-one', objectiveIds: ['objective.one'], title: 'A human lesson title' },
    { id: 'lesson-two', objectiveIds: ['objective.two'], title: 'Another lesson' },
  ],
  objectives: [
    {
      description: 'First objective.',
      id: 'objective.one',
      prerequisites: [],
      title: 'Understand the visual model',
    },
    {
      description: 'Second objective.',
      id: 'objective.two',
      prerequisites: ['objective.one'],
      title: 'Apply the visual model',
    },
  ],
  order: 0,
  title: 'Module title',
  videos: [
    {
      channel: 'Second Creator',
      courseId: 'course',
      durationMinutes: 9,
      id: 'second-video',
      lessonIds: ['lesson-two'],
      moduleId: 'module',
      objectiveIds: ['objective.two'],
      order: 1,
      title: 'Second in the playlist',
      youtubeId: '9bZkp7q19f0',
    },
    {
      channel: 'First Creator',
      courseId: 'course',
      durationMinutes: 6,
      id: 'first-video',
      lessonIds: ['lesson-one'],
      moduleId: 'module',
      objectiveIds: ['objective.one'],
      order: 0,
      title: 'First in the playlist',
      youtubeId: 'dQw4w9WgXcQ',
    },
  ],
  worksheets: [],
}

describe('ModuleVideoPlaylist', () => {
  it('renders nothing for a module without videos', () => {
    const { container } = render(
      <MemoryRouter>
        <ModuleVideoPlaylist module={{ ...moduleWithVideos, videos: [] }} />
      </MemoryRouter>,
    )

    expect(container.textContent).toBe('')
  })

  it('presents canonical order, human associations, public playback, and lesson navigation', () => {
    const { container } = render(
      <MemoryRouter>
        <ModuleVideoPlaylist module={moduleWithVideos} />
      </MemoryRouter>,
    )

    const choices = screen.getAllByRole('button')
    expect(choices[0].textContent).toContain('First in the playlist')
    expect(choices[0].getAttribute('aria-pressed')).toBe('true')
    expect(choices[1].textContent).toContain('Second in the playlist')
    expect(screen.getByTitle('First in the playlist by First Creator')).toBeTruthy()
    expect(screen.getByText('A human lesson title →')).toBeTruthy()
    expect(screen.getByText('Understand the visual model')).toBeTruthy()
    expect(container.textContent).not.toContain('objective.one')
    expect(container.textContent).not.toContain('lesson-one')
    expect(screen.getByRole('link', { name: 'A human lesson title →' }).getAttribute('href')).toBe(
      '/courses/course/modules/module/lessons/lesson-one',
    )

    const thumbnail = container.querySelector('img')
    expect(thumbnail?.getAttribute('src')).toBe('https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg')

    fireEvent.click(choices[1])
    expect(screen.getByTitle('Second in the playlist by Second Creator')).toBeTruthy()
    expect(choices[1].getAttribute('aria-pressed')).toBe('true')
    expect(screen.getByText('Another lesson →')).toBeTruthy()
    expect(screen.getByText('Apply the visual model')).toBeTruthy()
  })
})
