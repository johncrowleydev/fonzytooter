import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { TutorOverlay } from './TutorOverlay'
import { TutorProvider, useTutor } from './TutorContext'

const authState = vi.hoisted(() => ({
  isAuthenticated: true,
  accessStatus: 'allowed',
  accessOptions: vi.fn(),
}))

vi.mock('../authentication/AuthContext', () => ({
  useAuth: () => ({ isPending: false, isAuthenticated: authState.isAuthenticated }),
}))

vi.mock('../../api/generated/endpoints', () => ({
  useGetTutorAccess: (options: unknown) => {
    authState.accessOptions(options)
    return {
      isPending: false,
      isError: false,
      data: {
        data: {
          status: authState.accessStatus,
          monthlyTurnLimit: 10,
          usedTurns: 1,
          remainingTurns: 9,
          windowEndsAt: '2026-09-01T00:00:00Z',
        },
      },
    }
  },
}))

afterEach(cleanup)

/**
 * The overlay is the app's only modal, so it is the one place where keyboard handling is load
 * bearing: without a trap, Tab walks behind the scrim into a page the user cannot see.
 */
function TutorHarness() {
  const { openTutor } = useTutor()

  return (
    <div>
      <button type="button" onClick={openTutor}>
        Open tutor
      </button>
      <button type="button">Behind the scrim</button>
      <TutorOverlay />
    </div>
  )
}

function renderHarness() {
  return render(
    <MemoryRouter>
      <TutorProvider>
        <TutorHarness />
      </TutorProvider>
    </MemoryRouter>,
  )
}

const openTutor = async () => {
  await userEvent.click(screen.getByRole('button', { name: 'Open tutor' }))

  return screen.getByRole('dialog')
}

describe('TutorOverlay keyboard handling', () => {
  beforeEach(() => {
    authState.isAuthenticated = true
    authState.accessStatus = 'allowed'
    renderHarness()
  })

  it('names the dialog by its heading rather than leaving it anonymous', async () => {
    const dialog = await openTutor()

    expect(dialog.getAttribute('aria-modal')).toBe('true')
    // The role belongs on the panel; the scrim is not the dialog.
    expect(dialog.tagName).toBe('SECTION')
    expect(dialog.getAttribute('aria-labelledby')).toBe('tutor-dialog-title')
    expect(document.getElementById('tutor-dialog-title')?.textContent).toContain('Tutor')
  })

  it('focuses the prompt on open so typing works immediately', async () => {
    await openTutor()

    expect(document.activeElement).toBe(
      screen.getByPlaceholderText('Ask anything about this screen…'),
    )
  })

  it('closes on Escape', async () => {
    await openTutor()
    await userEvent.keyboard('{Escape}')

    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('returns focus to the control that opened it', async () => {
    const opener = screen.getByRole('button', { name: 'Open tutor' })
    await userEvent.click(opener)
    await userEvent.keyboard('{Escape}')

    expect(document.activeElement).toBe(opener)
  })

  it('keeps Tab inside the dialog instead of reaching the page behind it', async () => {
    const dialog = await openTutor()
    const behind = screen.getByRole('button', { name: 'Behind the scrim' })

    // Walk forwards well past the end of the dialog's own controls.
    for (let step = 0; step < 12; step += 1) {
      await userEvent.tab()
      expect(document.activeElement).not.toBe(behind)
      expect(dialog.contains(document.activeElement)).toBe(true)
    }
  })

  it('wraps backwards from the first control to the last', async () => {
    const dialog = await openTutor()
    const focusable = Array.from(dialog.querySelectorAll<HTMLElement>('button, textarea')).filter(
      (element) => !element.hasAttribute('disabled'),
    )
    const first = focusable[0]
    const last = focusable[focusable.length - 1]

    first.focus()
    await userEvent.tab({ shift: true })

    expect(document.activeElement).toBe(last)
  })

  it('locks background scrolling while open and restores it after', async () => {
    expect(document.body.style.overflow).toBe('')

    await openTutor()
    expect(document.body.style.overflow).toBe('hidden')

    await userEvent.keyboard('{Escape}')
    expect(document.body.style.overflow).toBe('')
  })
})

describe('TutorOverlay authentication boundary', () => {
  it('keeps the tutor discoverable but requires sign-in before interaction', async () => {
    authState.isAuthenticated = false
    renderHarness()
    await openTutor()

    expect(screen.getByRole('heading', { name: 'Sign in to ask the tutor' })).toBeDefined()
    expect(screen.getByRole('link', { name: 'Sign in and return here' })).toBeDefined()
    expect(screen.queryByPlaceholderText('Ask anything about this screen…')).toBeNull()
    expect(authState.accessOptions).toHaveBeenLastCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ enabled: false }) }),
    )
  })

  it.each([
    ['not_entitled', 'Tutor is unavailable for this account'],
    ['limit_exhausted', 'Tutor usage limit reached'],
  ])('renders the %s access state without an interaction form', async (status, heading) => {
    authState.isAuthenticated = true
    authState.accessStatus = status
    renderHarness()
    await openTutor()

    expect(screen.getByRole('heading', { name: heading })).toBeDefined()
    expect(screen.queryByPlaceholderText('Ask anything about this screen…')).toBeNull()
  })
})
