export type TutorMode = 'explain' | 'socratic' | 'exercise' | 'quiz' | 'explore'

export type TutorPageContext = {
  type: 'dashboard' | 'lesson' | 'exercise' | 'review' | 'project'
  moduleId?: string
  lessonId?: string
  objectiveIds?: string[]
  sectionId?: string
  exerciseId?: string
  selectedText?: string
  code?: string
  lastExecution?: {
    passed: number
    failed: number
    summary?: string
  }
}

export type TutorEvent = {
  type:
    | 'text_delta'
    | 'tool_started'
    | 'tool_completed'
    | 'citation'
    | 'usage'
    | 'completed'
    | 'error'
  text?: string
  tool?: string
  sourceId?: string
  error?: string
}

export type TutorMessage = {
  id: string
  role: 'user' | 'assistant'
  text: string
}
