import { useState, type Dispatch, type FormEvent, type SetStateAction } from 'react'
import { streamTutorTurn } from './api'
import { useTutor } from './TutorContext'
import type { TutorEvent, TutorMessage, TutorMode } from './types'

export function TutorOverlay() {
  const { isOpen, closeTutor, pageContext } = useTutor()
  const [input, setInput] = useState('')
  const [mode, setMode] = useState<TutorMode>('explain')
  const [messages, setMessages] = useState<TutorMessage[]>([])
  const [isStreaming, setIsStreaming] = useState(false)

  if (!isOpen) return null

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()

    const message = input.trim()
    if (!message || isStreaming) return

    const userMessage: TutorMessage = { id: crypto.randomUUID(), role: 'user', text: message }
    const assistantID = crypto.randomUUID()

    setMessages((current) => [
      ...current,
      userMessage,
      { id: assistantID, role: 'assistant', text: '' },
    ])
    setInput('')
    setIsStreaming(true)

    try {
      await streamTutorTurn({
        message,
        mode,
        pageContext,
        onEvent: (tutorEvent) => applyTutorEvent(assistantID, tutorEvent, setMessages),
      })
    } catch (error) {
      const text = error instanceof Error ? error.message : 'Tutor request failed.'
      setMessages((current) =>
        current.map((item) => (item.id === assistantID ? { ...item, text } : item)),
      )
    } finally {
      setIsStreaming(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/50" role="dialog" aria-modal="true">
      <section className="flex h-full w-full flex-col border-l border-slate-800 bg-slate-950 shadow-2xl sm:max-w-xl">
        <header className="flex items-center justify-between border-b border-slate-800 px-5 py-4">
          <div>
            <h2 className="font-semibold">Tutor</h2>
            <p className="text-xs text-slate-500">Context: {pageContext.type}</p>
          </div>
          <button
            type="button"
            onClick={closeTutor}
            className="rounded-lg px-3 py-2 text-sm text-slate-400 hover:bg-slate-900 hover:text-white"
          >
            Close
          </button>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto p-5">
          {messages.length === 0 ? (
            <p className="text-sm leading-6 text-slate-500">
              Ask about what you are currently studying. Screen context is sent with each turn.
            </p>
          ) : null}

          {messages.map((message) => (
            <div
              key={message.id}
              className={
                message.role === 'user'
                  ? 'ml-10 rounded-2xl bg-slate-800 px-4 py-3 text-sm'
                  : 'mr-10 whitespace-pre-wrap text-sm leading-6 text-slate-300'
              }
            >
              {message.text || (isStreaming && message.role === 'assistant' ? '…' : '')}
            </div>
          ))}
        </div>

        <form onSubmit={handleSubmit} className="border-t border-slate-800 p-4">
          <div className="mb-3 flex items-center gap-2">
            <label htmlFor="tutor-mode" className="text-xs text-slate-500">
              Mode
            </label>
            <select
              id="tutor-mode"
              value={mode}
              onChange={(event) => setMode(event.target.value as TutorMode)}
              className="rounded-lg border border-slate-800 bg-slate-900 px-2 py-1 text-xs"
            >
              <option value="explain">Explain</option>
              <option value="socratic">Socratic</option>
              <option value="exercise">Exercise help</option>
              <option value="quiz">Quiz</option>
              <option value="explore">Explore</option>
            </select>
          </div>

          <div className="flex gap-2">
            <textarea
              value={input}
              onChange={(event) => setInput(event.target.value)}
              placeholder="Ask the tutor…"
              rows={3}
              className="min-h-20 flex-1 resize-none rounded-xl border border-slate-800 bg-slate-900 px-3 py-2 text-sm outline-none placeholder:text-slate-600 focus:border-slate-600"
            />
            <button
              type="submit"
              disabled={isStreaming || !input.trim()}
              className="self-end rounded-xl bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-950 disabled:cursor-not-allowed disabled:opacity-40"
            >
              Send
            </button>
          </div>
        </form>
      </section>
    </div>
  )
}

function applyTutorEvent(
  assistantID: string,
  event: TutorEvent,
  setMessages: Dispatch<SetStateAction<TutorMessage[]>>,
) {
  if (event.type !== 'text_delta' && event.type !== 'error') return

  const delta = event.type === 'error' ? event.error ?? 'Tutor error.' : event.text ?? ''
  setMessages((current) =>
    current.map((message) =>
      message.id === assistantID ? { ...message, text: message.text + delta } : message,
    ),
  )
}
