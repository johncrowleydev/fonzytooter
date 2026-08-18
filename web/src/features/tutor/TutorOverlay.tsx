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
    <div className="tutor-backdrop" role="dialog" aria-modal="true">
      <section className="tutor-panel">
        <header className="tutor-header">
          <div>
            <p className="eyebrow">Always available</p>
            <h2>
              Tutor <span className="tutor-online">●</span>
            </h2>
          </div>
          <button
            type="button"
            onClick={closeTutor}
            className="icon-button"
            aria-label="Close tutor"
          >
            ×
          </button>
        </header>

        <div className="tutor-modes">
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
              className={mode === value ? 'mode-chip active' : 'mode-chip'}
            >
              {label}
            </button>
          ))}
        </div>

        <div className="tutor-messages">
          {messages.length === 0 ? (
            <div className="tutor-empty">
              <span className="empty-mark">✦</span>
              <p>
                Ask about what you are currently studying. I’ll keep the answer anchored to this
                screen.
              </p>
              <div className="suggestion-row">
                <button
                  type="button"
                  onClick={() => setInput('Can you explain the key idea here?')}
                >
                  Explain the key idea
                </button>
                <button type="button" onClick={() => setInput('Give me a hint')}>
                  Give me a hint
                </button>
              </div>
            </div>
          ) : null}

          {messages.map((message) => (
            <div
              key={message.id}
              className={message.role === 'user' ? 'tutor-message user' : 'tutor-message assistant'}
            >
              <span className="message-role">{message.role === 'user' ? 'You' : 'Tutor'}</span>
              <div>
                {message.text || (isThinking && message.role === 'assistant' ? 'Thinking…' : '')}
              </div>
            </div>
          ))}
        </div>

        <form onSubmit={handleSubmit} className="tutor-compose">
          <textarea
            value={input}
            onChange={(event) => setInput(event.target.value)}
            placeholder="Ask anything about this screen…"
            rows={3}
            className="tutor-input"
          />
          <div className="compose-footer">
            <span className="compose-hint">
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
