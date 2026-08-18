import { useEffect, useState, type FormEvent } from 'react'
import { getMockTutorResponse } from './api'
import { useTutor } from './TutorContext'
import type { TutorMessage, TutorMode } from './types'
import { Button } from '../../components/ui'

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
      className="fixed inset-0 z-50 flex justify-end bg-[rgba(3,8,14,0.65)] backdrop-blur-[4px]"
      role="dialog"
      aria-modal="true"
    >
      <section className="flex h-full w-[min(445px,100%)] flex-col border-l border-[var(--line-strong)] bg-[var(--panel)] shadow-[-25px_0_80px_rgba(0,0,0,0.28)] max-[640px]:w-full">
        <header className="flex items-center justify-between border-b border-[var(--line)] px-[25px] pb-[18px] pt-[23px] max-[640px]:px-[18px] max-[640px]:pb-3.5 max-[640px]:pt-[18px]">
          <div>
            <p className="mb-1.5 text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
              Always available
            </p>
            <h2 className="m-0 text-[21px] tracking-[-0.04em]">
              Tutor <span className="align-middle text-[10px] text-[var(--teal)]">●</span>
            </h2>
          </div>
          <button
            type="button"
            onClick={closeTutor}
            className="border-0 bg-transparent p-[3px] text-[19px] text-[var(--faint)] hover:text-[var(--ink)]"
            aria-label="Close tutor"
          >
            ×
          </button>
        </header>

        <div className="flex flex-wrap gap-[5px] px-[22px] pb-3.5 pt-[17px] max-[640px]:px-4">
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
              className={`rounded-full border px-[9px] py-[7px] text-[9px] ${mode === value ? 'border-[rgba(118,208,192,0.35)] bg-[rgba(118,208,192,0.08)] text-[var(--teal)]' : 'border-[var(--line)] bg-transparent text-[var(--faint)] hover:border-[rgba(118,208,192,0.35)] hover:bg-[rgba(118,208,192,0.08)] hover:text-[var(--teal)]'}`}
            >
              {label}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto px-[22px] py-[7px] max-[640px]:px-4">
          {messages.length === 0 ? (
            <div className="p-[38px_15px] text-center">
              <span className="block text-[21px] text-[var(--gold)]">✦</span>
              <p className="text-xs leading-[1.6] text-[var(--muted)]">
                Ask about what you are currently studying. I’ll keep the answer anchored to this
                screen.
              </p>
              <div className="mt-5 flex flex-wrap justify-center gap-[5px]">
                <button
                  className="rounded-full border border-[var(--line)] bg-transparent px-[9px] py-[7px] text-[9px] text-[var(--faint)] hover:text-[var(--ink)]"
                  type="button"
                  onClick={() => setInput('Can you explain the key idea here?')}
                >
                  Explain the key idea
                </button>
                <button
                  className="rounded-full border border-[var(--line)] bg-transparent px-[9px] py-[7px] text-[9px] text-[var(--faint)] hover:text-[var(--ink)]"
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
              className={`my-3 rounded-[10px] px-[13px] py-3 text-xs leading-[1.65] ${message.role === 'user' ? 'ml-11 bg-[rgba(157,185,194,0.12)] text-[var(--ink)]' : 'mr-[25px] border border-[var(--line)] bg-white/[0.025] text-[var(--muted)]'}`}
            >
              <span
                className={`mb-1.5 block text-[9px] uppercase tracking-[0.1em] ${message.role === 'user' ? 'text-[var(--teal)]' : 'text-[var(--faint)]'}`}
              >
                {message.role === 'user' ? 'You' : 'Tutor'}
              </span>
              <div>
                {message.text || (isThinking && message.role === 'assistant' ? 'Thinking…' : '')}
              </div>
            </div>
          ))}
        </div>

        <form
          onSubmit={handleSubmit}
          className="border-t border-[var(--line)] px-[22px] pb-5 pt-[15px] max-[640px]:px-4"
        >
          <textarea
            value={input}
            onChange={(event) => setInput(event.target.value)}
            placeholder="Ask anything about this screen…"
            rows={3}
            className="block min-h-[78px] w-full resize-none rounded-[9px] border border-[var(--line-strong)] bg-[var(--panel-soft)] p-[11px] text-xs leading-[1.5] text-[var(--ink)] outline-0 placeholder:text-[var(--faint)] focus:border-[rgba(118,208,192,0.5)]"
          />
          <div className="mt-[9px] flex items-center justify-between gap-2.5">
            <span className="text-[9px] text-[var(--faint)]">
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
