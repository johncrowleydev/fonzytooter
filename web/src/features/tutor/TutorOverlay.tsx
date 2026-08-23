import { useEffect, useState, type FormEvent } from 'react'
import { getMockTutorResponse } from './api'
import { useTutor } from './TutorContext'
import type { TutorMessage, TutorMode } from './types'
import { Button } from '../../components/ui'

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

  useEffect(() => {
    if (!isOpen) return
    setInput('')
  }, [isOpen, pageContext.type])

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
      role="dialog"
      aria-modal="true"
      onClick={(event) => {
        if (event.target === event.currentTarget) closeTutor()
      }}
    >
      <section className="flex h-full w-full max-w-md flex-col border-l border-line-strong bg-panel shadow-2xl max-sm:max-w-none">
        <header className="flex items-center justify-between border-b border-line px-6 pb-5 pt-6 max-sm:px-4 max-sm:pb-4 max-sm:pt-5">
          <div>
            <p className="mb-1.5 text-xs font-bold uppercase tracking-widest text-faint">
              Always available
            </p>
            <h2 className="m-0 text-2xl tracking-tight">
              Tutor <span className="align-middle text-sm text-accent-teal">●</span>
            </h2>
          </div>
          <button
            type="button"
            onClick={closeTutor}
            className="border-0 bg-transparent p-1 text-xl text-faint hover:text-ink"
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
              className={`rounded-full border px-2 py-2 text-sm ${mode === value ? modeButtonStyles.active : modeButtonStyles.inactive}`}
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
                  className="rounded-full border border-line bg-transparent px-2 py-2 text-sm text-faint hover:text-ink"
                  type="button"
                  onClick={() => setInput('Can you explain the key idea here?')}
                >
                  Explain the key idea
                </button>
                <button
                  className="rounded-full border border-line bg-transparent px-2 py-2 text-sm text-faint hover:text-ink"
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
            value={input}
            onChange={(event) => setInput(event.target.value)}
            placeholder="Ask anything about this screen…"
            rows={3}
            className="block min-h-20 w-full resize-none rounded-lg border border-line-strong bg-panel-soft p-3 text-sm leading-normal text-ink outline-0 placeholder:text-faint focus:border-accent-teal/50"
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
