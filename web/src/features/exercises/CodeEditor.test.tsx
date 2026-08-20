import { fireEvent, render } from '@testing-library/react'
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

  it('recreates the editor when the exercise identity changes', () => {
    const { container, rerender } = render(
      <CodeEditor key="exercise-a" onChange={vi.fn()} value={'print("exercise a")'} />,
    )
    const firstEditor = container.querySelector('.cm-editor')

    rerender(<CodeEditor key="exercise-b" onChange={vi.fn()} value={'print("exercise b")'} />)

    const secondEditor = container.querySelector('.cm-editor')
    expect(secondEditor).not.toBe(firstEditor)
    expect(secondEditor?.textContent).toContain('print("exercise b")')
    expect(secondEditor?.textContent).not.toContain('print("exercise a")')

    fireEvent.keyDown(container.querySelector('.cm-content')!, { key: 'z', ctrlKey: true })

    expect(secondEditor?.textContent).toContain('print("exercise b")')
    expect(secondEditor?.textContent).not.toContain('print("exercise a")')
  })
})
