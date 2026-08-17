import type { TutorEvent, TutorMode, TutorPageContext } from './types'

type StreamTutorTurnRequest = {
  message: string
  mode: TutorMode
  pageContext: TutorPageContext
  onEvent: (event: TutorEvent) => void
  signal?: AbortSignal
}

export async function streamTutorTurn({
  message,
  mode,
  pageContext,
  onEvent,
  signal,
}: StreamTutorTurnRequest) {
  const response = await fetch('/api/tutor/turn', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message, mode, pageContext }),
    signal,
  })

  if (!response.ok) {
    const detail = await response.text()
    throw new Error(detail || `Tutor request failed with ${response.status}`)
  }

  if (!response.body) {
    throw new Error('Tutor response did not include a stream')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const frames = buffer.split('\n\n')
    buffer = frames.pop() ?? ''

    for (const frame of frames) {
      const data = frame
        .split('\n')
        .find((line) => line.startsWith('data: '))
        ?.slice(6)

      if (!data) continue
      onEvent(JSON.parse(data) as TutorEvent)
    }
  }
}
