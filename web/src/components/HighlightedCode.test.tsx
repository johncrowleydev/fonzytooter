import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { HighlightedCode, languageFromClassName, tokenizeCode } from './HighlightedCode'

afterEach(cleanup)

const PYTHON_SAMPLE = `import math


def area(radius, cache={}):
    """Docstring."""
    # A comment with punctuation: (x), [y].
    if radius <= 0:
        return None
    return math.pi * radius**2  # tail comment
`

describe('tokenizeCode', () => {
  /**
   * The important property. highlightTree only reports styled ranges, so an implementation that
   * forgets to fill the gaps silently drops indentation and blank lines -- which would corrupt
   * displayed Python rather than merely under-styling it.
   */
  it('reproduces the source exactly, including indentation and blank lines', () => {
    const rebuilt = tokenizeCode(PYTHON_SAMPLE, 'python')
      .map((token) => token.text)
      .join('')

    expect(rebuilt).toBe(PYTHON_SAMPLE)
  })

  it('reproduces the source exactly for an unknown language', () => {
    const rebuilt = tokenizeCode(PYTHON_SAMPLE, 'brainfuck')
      .map((token) => token.text)
      .join('')

    expect(rebuilt).toBe(PYTHON_SAMPLE)
  })

  it('tags keywords, strings, numbers, and comments', () => {
    const tokens = tokenizeCode(PYTHON_SAMPLE, 'python')
    const classFor = (text: string) => tokens.find((token) => token.text === text)?.className

    expect(classFor('def')).toContain('tok-keyword')
    expect(classFor('return')).toContain('tok-keyword')
    expect(classFor('# tail comment')).toContain('tok-comment')
    expect(classFor('2')).toContain('tok-number')
    expect(classFor('area')).toContain('tok-definition')
  })

  it('leaves an unknown language as a single untagged token', () => {
    const tokens = tokenizeCode('SELECT 1', 'sql')

    expect(tokens).toEqual([{ text: 'SELECT 1' }])
  })

  it('treats a missing language as plain text rather than guessing', () => {
    expect(tokenizeCode('def f(): pass')).toEqual([{ text: 'def f(): pass' }])
  })

  it('handles an empty string', () => {
    expect(tokenizeCode('', 'python')).toEqual([])
  })
})

describe('languageFromClassName', () => {
  it('reads the language off a fenced code element', () => {
    expect(languageFromClassName('language-python')).toBe('python')
    expect(languageFromClassName('block font-mono language-Python')).toBe('python')
  })

  it('returns undefined for inline code and unrelated classes', () => {
    expect(languageFromClassName(undefined)).toBeUndefined()
    expect(languageFromClassName('rounded bg-panel-soft font-mono')).toBeUndefined()
    // Guards against matching a class that merely ends in something language-like.
    expect(languageFromClassName('my-language-thing')).toBeUndefined()
  })
})

describe('HighlightedCode', () => {
  it('renders the code verbatim as text content', () => {
    const { container } = render(<HighlightedCode code={PYTHON_SAMPLE} language="python" />)

    expect(container.textContent).toBe(PYTHON_SAMPLE)
  })

  it('emits themed token spans for Python', () => {
    const { container } = render(<HighlightedCode code="def f(): return 1" language="python" />)

    expect(container.querySelectorAll('.tok-keyword').length).toBeGreaterThan(0)
    expect(container.querySelector('.tok-number')?.textContent).toBe('1')
  })

  it('emits no spans when the language is not supported', () => {
    const { container } = render(<HighlightedCode code="echo hi" language="bash" />)

    expect(container.querySelectorAll('span').length).toBe(0)
    expect(container.textContent).toBe('echo hi')
  })
})
