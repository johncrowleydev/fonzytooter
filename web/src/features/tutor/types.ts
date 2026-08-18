export type TutorMode = 'explain' | 'socratic' | 'exercise' | 'quiz' | 'explore'

import type { MockCheckResult } from '../../prototype/types'

export type TutorPageContext = {
  type: 'dashboard' | 'curriculum' | 'lesson' | 'exercise' | 'review' | 'progress' | 'project'
  title?: string
  moduleId?: string
  moduleTitle?: string
  lessonId?: string
  lessonTitle?: string
  objectiveIds?: string[]
  sectionId?: string
  exerciseId?: string
  exerciseTitle?: string
  selectedText?: string
  code?: string
  lastExecution?: MockCheckResult
  objectiveId?: string
  objectiveTitle?: string
  projectId?: string
  projectTitle?: string
}

export type TutorEvent = {
  type:
    'text_delta' | 'tool_started' | 'tool_completed' | 'citation' | 'usage' | 'completed' | 'error'
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
