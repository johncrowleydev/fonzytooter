import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { StatusDot } from './ui'

afterEach(cleanup)

describe('StatusDot', () => {
  /**
   * The dot conveys state entirely through color, so its label is the only version a screen reader
   * gets. It used to pass the state key straight through, announcing "in-progress" and
   * "not-assessed" -- internal vocabulary that happens to look like words.
   */
  it.each([
    ['locked', 'Locked'],
    ['todo', 'Not started'],
    ['available', 'Available'],
    ['in-progress', 'In progress'],
    ['working', 'In progress'],
    ['completed', 'Completed'],
    ['done', 'Completed'],
  ] as const)('announces %s as %s', (state, label) => {
    render(<StatusDot state={state} />)

    expect(screen.getByRole('img', { name: label })).toBeDefined()
  })

  it('never announces a raw state key', () => {
    render(
      <>
        <StatusDot state="in-progress" />
        <StatusDot state="todo" />
      </>,
    )

    expect(screen.queryByRole('img', { name: 'in-progress' })).toBeNull()
    expect(screen.queryByRole('img', { name: 'todo' })).toBeNull()
  })
})
