import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { CourseProgressResource } from '../../api/generated/schemas/courseProgressResource.zod'
import { ProgressView } from './Progress'

afterEach(cleanup)

const progress: CourseProgressResource = {
  courseId: 'ai-ml',
  completedLessonCount: 1,
  totalLessonCount: 2,
  dueReviewCount: 0,
  objectives: [
    {
      courseId: 'ai-ml',
      moduleId: '01-scientific-python',
      id: 'python.execution-model',
      title: 'Reason about the Python execution model',
      description: 'Trace how names bind to objects.',
      introduced: true,
      linkedLessonCount: 1,
      completedLessonCount: 1,
      recall: { reviewItemCount: 0, reviewsCompleted: 0, dueReviewCount: 0 },
      application: { exerciseCount: 0, attempts: 0, fullyPassedAttempts: 0 },
      transferAssessed: false,
    },
  ],
}

const moduleTitles = new Map([['01-scientific-python', 'Scientific Python foundations']])

describe('objective rows name their module', () => {
  it('shows the module title rather than its identifier', () => {
    render(<ProgressView progress={progress} onSelect={() => {}} moduleTitles={moduleTitles} />)

    expect(screen.getByText('Scientific Python foundations')).toBeDefined()
  })

  /**
   * The row used to de-slugify `moduleId` and capitalise it, producing "01 scientific python" --
   * prose-shaped text that was still an identifier. Asserting on the identifier's own shape rather
   * than one rendering of it keeps this honest if the transform is ever reintroduced differently.
   */
  it('never shows the module identifier, de-slugified or otherwise', () => {
    const { container } = render(
      <ProgressView progress={progress} onSelect={() => {}} moduleTitles={moduleTitles} />,
    )

    expect(container.textContent).not.toContain('01-scientific-python')
    expect(container.textContent).not.toContain('01 scientific python')
    expect(container.textContent).not.toContain('python.execution-model')
  })

  it('omits the module line entirely while the course is still loading', () => {
    const { container } = render(<ProgressView progress={progress} onSelect={() => {}} />)

    // Better to show nothing than to fall back to the identifier.
    expect(container.textContent).not.toContain('01-scientific-python')
    // The objective itself still renders; it appears in both the row and the detail panel.
    expect(screen.getAllByText('Reason about the Python execution model').length).toBeGreaterThan(0)
  })
})
