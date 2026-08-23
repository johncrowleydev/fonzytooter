import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Button, StatusDot } from './ui'

afterEach(cleanup)

describe('Button', () => {
  /**
   * A disabled button that looks enabled is worse than no disabled state: the learner clicks it and
   * nothing happens. The shared button had no disabled styling at all, while two lesson components
   * had rolled their own, so the shared primitive was the weakest of the two standards.
   */
  it('looks disabled when it is disabled, and ignores clicks', async () => {
    const onClick = vi.fn()
    render(
      <Button disabled onClick={onClick}>
        Locked
      </Button>,
    )
    const button = screen.getByRole<HTMLButtonElement>('button', { name: 'Locked' })

    expect(button.disabled).toBe(true)
    expect(button.className).toContain('disabled:opacity-50')
    expect(button.className).toContain('disabled:cursor-not-allowed')

    await userEvent.click(button)
    expect(onClick).not.toHaveBeenCalled()
  })

  it('exposes toggle state so selection is not colour-only', () => {
    render(
      <>
        <Button pressed variant="primary">
          Active
        </Button>
        <Button pressed={false} variant="outline">
          Inactive
        </Button>
      </>,
    )

    expect(screen.getByRole('button', { name: 'Active' }).getAttribute('aria-pressed')).toBe('true')
    expect(screen.getByRole('button', { name: 'Inactive' }).getAttribute('aria-pressed')).toBe(
      'false',
    )
  })

  it('omits aria-pressed entirely for a plain action button', () => {
    render(<Button>Run</Button>)

    // A plain button must not claim to be a toggle.
    expect(screen.getByRole('button', { name: 'Run' }).hasAttribute('aria-pressed')).toBe(false)
  })
})

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
