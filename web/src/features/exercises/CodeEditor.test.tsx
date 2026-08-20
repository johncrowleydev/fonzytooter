import { render } from '@testing-library/react'
import { beforeAll, describe, expect, it, vi } from 'vitest'
import { CodeEditor } from './CodeEditor'

describe('CodeEditor', () => {
  beforeAll(() => {
    Object.defineProperty(Range.prototype, 'getClientRects', { value: () => [] })
  })

  it('reconfigures editability without recreating the editor', () => {
    const { container, rerender } = render(
      <CodeEditor onChange={vi.fn()} value={'print("ready")'} />,
    )
    const editor = container.querySelector('.cm-editor')
    const content = container.querySelector('.cm-content')

    expect(editor).not.toBeNull()
    expect(content?.getAttribute('contenteditable')).toBe('true')

    rerender(<CodeEditor disabled onChange={vi.fn()} value={'print("ready")'} />)

    expect(container.querySelector('.cm-editor')).toBe(editor)
    expect(container.querySelector('.cm-content')).toBe(content)
    expect(content?.getAttribute('contenteditable')).toBe('false')
  })
})
