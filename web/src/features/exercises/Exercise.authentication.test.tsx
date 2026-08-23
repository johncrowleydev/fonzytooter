import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { TutorProvider } from '../tutor/TutorContext'
import { Exercise } from './Exercise'

const mocks = vi.hoisted(() => ({
  workspaceOptions: vi.fn(),
  checkDefinition: vi.fn(),
  createAttempt: vi.fn(),
  exerciseQuery: {
    data: {
      data: {
        courseId: 'ai-ml',
        moduleId: 'python',
        lessonId: 'intro',
        id: 'double',
        title: 'Double a number',
        prompt: 'Write a doubling function.',
        starterCode: 'def double(value):\n    return value * 2',
        objectiveIds: [],
        order: 1,
        visibleTests: [
          { id: 'visible-double', title: 'Doubles two', code: 'assert double(2) == 4' },
        ],
      },
    },
    isLoading: false,
  },
  courseQuery: { data: { data: { id: 'ai-ml', title: 'AI & Machine Learning' } } },
  moduleQuery: { data: { data: { id: 'python', title: 'Python', objectives: [] } } },
  lessonQuery: { data: { data: { id: 'intro', title: 'Introduction' } } },
}))

vi.mock('../authentication/AuthContext', () => ({
  useAuth: () => ({ isPending: false, isAuthenticated: false }),
}))

vi.mock('../../api/generated/endpoints', () => ({
  useGetCourseModuleExercise: () => mocks.exerciseQuery,
  useGetExerciseWorkspace: (...args: unknown[]) => {
    mocks.workspaceOptions(args.at(-1))
    return { isLoading: false, isError: false }
  },
  useGetExerciseCheckDefinition: () => ({ refetch: mocks.checkDefinition }),
  useGetCourse: () => mocks.courseQuery,
  useGetCourseModule: () => mocks.moduleQuery,
  useGetCourseLesson: () => mocks.lessonQuery,
  usePutExerciseWorkspace: () => ({ mutateAsync: vi.fn() }),
  useCreateExerciseAttempt: () => ({ mutateAsync: mocks.createAttempt }),
}))

vi.mock('./runtime/PyodideRunner', () => ({
  PyodideRunner: class {
    async run() {
      return { stdout: 'browser run complete', stderr: '', durationMs: 1 }
    }

    async check({ tests }: { tests: Array<{ id: string; title: string }> }) {
      return {
        stdout: '',
        stderr: '',
        durationMs: 1,
        tests: tests.map((test) => ({
          testId: test.id,
          title: test.title,
          visibility: 'visible' as const,
          status: 'passed' as const,
          message: '',
          durationMs: 1,
        })),
      }
    }

    dispose() {}
  },
}))

vi.mock('./CodeEditor', () => ({
  CodeEditor: ({ value }: { value: string }) => (
    <pre aria-label="Python exercise editor">{value}</pre>
  ),
}))

afterEach(cleanup)
beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
})

describe('anonymous browser-only exercises', () => {
  it('runs and checks visible tests without private calls or persistence', async () => {
    const path = '/courses/ai-ml/modules/python/exercises/double'
    render(
      <MemoryRouter initialEntries={[path]}>
        <TutorProvider>
          <Routes>
            <Route
              path="/courses/:courseId/modules/:moduleId/exercises/:exerciseId"
              element={<Exercise />}
            />
          </Routes>
        </TutorProvider>
      </MemoryRouter>,
    )

    expect(await screen.findByText('Browser-only · not saved')).toBeDefined()
    expect(mocks.workspaceOptions).toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ enabled: false }) }),
    )

    await userEvent.click(screen.getByRole('button', { name: 'Run ▶' }))
    expect(await screen.findByText('browser run complete')).toBeDefined()

    await userEvent.click(screen.getByRole('button', { name: 'Check visible tests ✓' }))
    await waitFor(() => expect(screen.getByText('1 passed · 0 failed')).toBeDefined())
    expect(mocks.checkDefinition).not.toHaveBeenCalled()
    expect(mocks.createAttempt).not.toHaveBeenCalled()
    expect(localStorage.length).toBe(0)
    expect(screen.getByRole('link', { name: 'Sign in to save attempts' })).toBeDefined()
  })
})
