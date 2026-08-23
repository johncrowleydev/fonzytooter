import { useCallback, useEffect, useRef, useState, type FormEvent, type KeyboardEvent } from 'react'
import { getMockTutorResponse } from './api'
import { useTutor } from './TutorContext'
import type { TutorMessage, TutorMode } from './types'
import { Button } from '../../components/ui'

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'

const modeButtonStyles = {
  active: 'border-accent-teal/40 bg-accent-teal/10 text-accent-teal',
  inactive:
    'border-line bg-transparent text-faint hover:border-accent-teal/40 hover:bg-accent-teal/10 hover:text-accent-teal',
} as const

const messageStyles = {
  user: 'ml-11 bg-accent-slate/15 text-ink',
  assistant: 'mr-6 border border-line bg-raised text-muted',
} as const

const messageLabelStyles = {
  user: 'text-accent-teal',
  assistant: 'text-faint',
} as const

export function TutorOverlay() {
  const { isOpen, closeTutor, pageContext } = useTutor()
  const [input, setInput] = useState('')
  const [mode, setMode] = useState<TutorMode>('explain')
  const [messages, setMessages] = useState<TutorMessage[]>([])
  const [isThinking, setIsThinking] = useState(false)
  const panelRef = useRef<HTMLElement>(null)
  const promptRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (!isOpen) return
    setInput('')
  }, [isOpen, pageContext.type])

  // Opening a dialog should land the caret where the user is going to type, and closing it should
  // return them to whatever they were on rather than dropping focus to the top of the document.
  useEffect(() => {
    if (!isOpen) return

    const previouslyFocused = document.activeElement as HTMLElement | null
    promptRef.current?.focus()

    return () => previouslyFocused?.focus()
  }, [isOpen])

  // The page behind a full-height panel should not scroll under it.
  useEffect(() => {
    if (!isOpen) return

    const { overflow } = document.body.style
    document.body.style.overflow = 'hidden'

    return () => {
      document.body.style.overflow = overflow
    }
  }, [isOpen])

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.stopPropagation()
        closeTutor()
        return
      }

      if (event.key !== 'Tab' || !panelRef.current) return

      // Without a trap, Tab walks straight out of the dialog and into the page behind it, which
      // is still covered by the scrim.
      //
      // Filtering on `hidden`/`aria-hidden` rather than `offsetParent`: the latter needs layout,
      // which means it reads as null under jsdom and would silently disable the trap in tests.
      const focusable = Array.from(
        panelRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
      ).filter((element) => !element.hidden && element.getAttribute('aria-hidden') !== 'true')

      if (focusable.length === 0) return

      const first = focusable[0]
      const last = focusable[focusable.length - 1]

      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault()
        last.focus()
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault()
        first.focus()
      }
    },
    [closeTutor],
  )

  if (!isOpen) return null

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()

    const message = input.trim()
    if (!message || isThinking) return

    const userMessage: TutorMessage = { id: crypto.randomUUID(), role: 'user', text: message }
    const assistantID = crypto.randomUUID()

    setMessages((current) => [
      ...current,
      userMessage,
      { id: assistantID, role: 'assistant', text: '' },
    ])
    setInput('')
    setIsThinking(true)

    try {
      const response = await getMockTutorResponse({
        message,
        mode,
        pageContext,
      })
      setMessages((current) =>
        current.map((item) => (item.id === assistantID ? { ...item, text: response } : item)),
      )
    } catch (error) {
      const text = error instanceof Error ? error.message : 'Tutor request failed.'
      setMessages((current) =>
        current.map((item) => (item.id === assistantID ? { ...item, text } : item)),
      )
    } finally {
      setIsThinking(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex justify-end bg-black/65 backdrop-blur-sm"
      onKeyDown={handleKeyDown}
      onClick={(event) => {
        if (event.target === event.currentTarget) closeTutor()
      }}
    >
      {/* role and aria-modal belong on the panel, not the scrim, so the dialog is what gets named. */}
      <section
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="tutor-dialog-title"
        className="flex h-full w-full max-w-md flex-col border-l border-line-strong bg-panel shadow-2xl max-sm:max-w-none"
      >
        <header className="flex items-center justify-between border-b border-line px-6 pb-5 pt-6 max-sm:px-4 max-sm:pb-4 max-sm:pt-5">
          <div>
            <p className="mb-1.5 text-xs font-bold uppercase tracking-widest text-faint">
              Always available
            </p>
            <h2 className="m-0 text-2xl tracking-tight" id="tutor-dialog-title">
              Tutor{' '}
              <span className="align-middle text-sm text-accent-teal" aria-hidden="true">
                ●
              </span>
            </h2>
          </div>
          <button
            type="button"
            onClick={closeTutor}
            className="grid size-9 shrink-0 place-items-center rounded-lg border-0 bg-transparent text-xl text-faint hover:text-ink pointer-coarse:size-11"
            aria-label="Close tutor"
          >
            ×
          </button>
        </header>

        <div className="flex flex-wrap gap-1.5 px-5 pb-3.5 pt-4 max-sm:px-4">
          {(
            [
              ['explain', 'Explain'],
              ['socratic', 'Socratic'],
              ['exercise', 'Exercise help'],
              ['quiz', 'Quiz'],
              ['explore', 'Explore'],
            ] as const
          ).map(([value, label]) => (
            <button
              key={value}
              type="button"
              onClick={() => setMode(value)}
              className={`rounded-full border px-3 py-2 text-sm pointer-coarse:min-h-11 pointer-coarse:px-4 ${mode === value ? modeButtonStyles.active : modeButtonStyles.inactive}`}
            >
              {label}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto px-5 py-2 max-sm:px-4">
          {messages.length === 0 ? (
            <div className="p-10 text-center">
              <span className="block text-2xl text-accent-gold">✦</span>
              <p className="text-sm leading-relaxed text-muted">
                Ask about what you are currently studying. I’ll keep the answer anchored to this
                screen.
              </p>
              <div className="mt-5 flex flex-wrap justify-center gap-1.5">
                <button
                  className="rounded-full border border-line bg-transparent px-3 py-2 text-sm text-faint hover:text-ink pointer-coarse:min-h-11 pointer-coarse:px-4"
                  type="button"
                  onClick={() => setInput('Can you explain the key idea here?')}
                >
                  Explain the key idea
                </button>
                <button
                  className="rounded-full border border-line bg-transparent px-3 py-2 text-sm text-faint hover:text-ink pointer-coarse:min-h-11 pointer-coarse:px-4"
                  type="button"
                  onClick={() => setInput('Give me a hint')}
                >
                  Give me a hint
                </button>
              </div>
            </div>
          ) : null}

          {messages.map((message) => (
            <div
              key={message.id}
              className={`my-3 rounded-lg px-3 py-3 text-sm leading-relaxed ${messageStyles[message.role]}`}
            >
              <span
                className={`mb-1.5 block text-xs uppercase tracking-wide ${messageLabelStyles[message.role]}`}
              >
                {message.role === 'user' ? 'You' : 'Tutor'}
              </span>
              <div>
                {message.text || (isThinking && message.role === 'assistant' ? 'Thinking…' : '')}
              </div>
            </div>
          ))}
        </div>

        <form onSubmit={handleSubmit} className="border-t border-line px-5 pb-5 pt-4 max-sm:px-4">
          <textarea
            ref={promptRef}
            value={input}
            onChange={(event) => setInput(event.target.value)}
            placeholder="Ask anything about this screen…"
            rows={3}
            className="block min-h-20 w-full resize-none rounded-lg border border-line-strong bg-panel-soft p-3 text-sm leading-normal text-ink placeholder:text-faint focus-visible:border-accent-teal"
          />
          <div className="mt-2 flex items-center justify-between gap-2.5">
            <span className="text-sm text-faint">
              {mode === 'exercise' ? 'Hints, not solutions' : 'Ask a question'}
            </span>
            <Button type="submit" disabled={isThinking || !input.trim()}>
              {isThinking ? 'Thinking…' : 'Send'} <span>↗</span>
            </Button>
          </div>
        </form>
      </section>
    </div>
  )
}
