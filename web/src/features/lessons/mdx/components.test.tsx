import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lessonMdxComponents } from './components'

afterEach(cleanup)

/**
 * Covers the wiring rather than the highlighter itself. The renderer depends on MDX handing a
 * fenced block's source over as a single string child; if that ever changes shape, highlighting
 * would quietly stop without any error.
 */
const LessonCode = lessonMdxComponents.code

describe('lesson MDX code rendering', () => {
  it('highlights a fenced Python block', () => {
    const { container } = render(
      <LessonCode className="language-python">{'def f():\n    return 1\n'}</LessonCode>,
    )

    expect(container.querySelectorAll('.tok-keyword').length).toBeGreaterThan(0)
    expect(container.textContent).toBe('def f():\n    return 1\n')
  })

  it('leaves inline code unhighlighted and on the inline style', () => {
    const { container } = render(<LessonCode>{'shape'}</LessonCode>)
    const code = container.querySelector('code')

    expect(code?.querySelectorAll('span').length).toBe(0)
    expect(code?.className).toContain('text-accent-teal')
    expect(code?.textContent).toBe('shape')
  })

  it('renders a fenced block in an unsupported language as plain code', () => {
    const { container } = render(<LessonCode className="language-text">{'output: 3'}</LessonCode>)

    expect(container.querySelectorAll('span').length).toBe(0)
    expect(container.textContent).toBe('output: 3')
  })

  it('keeps the language class on the element so styling can key off it', () => {
    const { container } = render(<LessonCode className="language-python">{'x = 1'}</LessonCode>)

    expect(container.querySelector('code')?.className).toContain('language-python')
  })
})
