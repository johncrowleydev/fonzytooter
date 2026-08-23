import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { Dashboard } from '../dashboard/Dashboard'
import { Lesson } from '../lessons/Lesson'
import { TutorProvider } from '../tutor/TutorContext'

const mocks = vi.hoisted(() => ({
  lessonProgressOptions: vi.fn(),
  courseProgressOptions: vi.fn(),
  activityOptions: vi.fn(),
  courseQuery: {
    data: {
      data: {
        id: 'ai-ml',
        title: 'AI & Machine Learning',
        description: 'A public course.',
        order: 1,
        modules: [{ id: 'python', title: 'Python', order: 1 }],
      },
    },
  },
  moduleQuery: {
    data: {
      data: {
        courseId: 'ai-ml',
        id: 'python',
        title: 'Python',
        order: 1,
        lessons: [{ id: 'public-lesson', title: 'Public lesson', order: 1 }],
        objectives: [],
        exercises: [],
        videos: [],
        worksheets: [],
      },
    },
  },
  lessonQuery: {
    data: {
      data: {
        courseId: 'ai-ml',
        moduleId: 'python',
        id: 'public-lesson',
        title: 'Public lesson',
        content: '# Read this lesson\n\nComplete public curriculum content.',
        objectiveIds: [],
        sources: [],
        exercises: [],
        worksheets: [],
      },
    },
  },
}))

vi.mock('./AuthContext', () => ({
  useAuth: () => ({ isPending: false, isAuthenticated: false }),
}))

vi.mock('../../api/generated/endpoints', () => ({
  getGetCourseProgressQueryKey: vi.fn(),
  getGetLessonProgressQueryKey: vi.fn(),
  getListActivitiesQueryKey: vi.fn(),
  useGetCourse: () => mocks.courseQuery,
  useGetCourseModule: () => mocks.moduleQuery,
  useGetCourseLesson: () => mocks.lessonQuery,
  useGetLessonProgress: (...args: unknown[]) => {
    mocks.lessonProgressOptions(args.at(-1))
    return { isPending: false }
  },
  usePutLessonProgress: () => ({ isPending: false, mutate: vi.fn() }),
  useGetCourseProgress: (...args: unknown[]) => {
    mocks.courseProgressOptions(args.at(-1))
    return { isPending: false }
  },
  useListActivities: (...args: unknown[]) => {
    mocks.activityOptions(args.at(-1))
    return { isPending: false }
  },
}))

function renderRoute(path: string, element: React.ReactNode) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const routePath = path.includes('/lessons/')
    ? '/courses/:courseId/modules/:moduleId/lessons/:lessonId'
    : path
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[path]}>
        <TutorProvider>
          <Routes>
            <Route path={routePath} element={element} />
          </Routes>
        </TutorProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(cleanup)
beforeEach(() => vi.clearAllMocks())

describe('anonymous curriculum boundary', () => {
  it('renders a directly navigated lesson without requesting private progress', async () => {
    const path = '/courses/ai-ml/modules/python/lessons/public-lesson'
    renderRoute(path, <Lesson />)

    expect(await screen.findByRole('heading', { name: 'Read this lesson' })).toBeDefined()
    expect(screen.getByText('Complete public curriculum content.')).toBeDefined()
    expect(screen.getByRole('link', { name: 'Sign in to track progress' })).toBeDefined()
    expect(mocks.lessonProgressOptions).toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ enabled: false }) }),
    )
  })

  it('shows a learner gate on the dashboard without firing private queries', async () => {
    renderRoute('/', <Dashboard />)

    expect(screen.getByRole('heading', { name: 'Your learning dashboard' })).toBeDefined()
    expect(screen.queryByText('No activity yet')).toBeNull()
    expect(mocks.courseProgressOptions).toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ enabled: false }) }),
    )
    expect(mocks.activityOptions).toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ enabled: false }) }),
    )
    await waitFor(() => expect(screen.getByRole('link', { name: 'Sign in' })).toBeDefined())
  })
})
