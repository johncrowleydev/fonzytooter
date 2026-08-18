export type ModuleStatus = 'locked' | 'available' | 'in-progress' | 'completed'
export type LessonKind = 'lesson' | 'exercise' | 'video' | 'lab'
export type MasteryLevel = 'not-assessed' | 'developing' | 'strong'

export type LessonSummary = {
  id: string
  title: string
  kind: LessonKind
  completed: boolean
  objectiveIds: string[]
}

export type CurriculumModule = {
  id: string
  eyebrow: string
  title: string
  description: string
  status: ModuleStatus
  objectiveIds: string[]
  lessons: LessonSummary[]
  prerequisites?: string[]
  accent: string
}

export type Objective = {
  id: string
  title: string
  description: string
  moduleId: string
  prerequisiteIds: string[]
  introduced: boolean
  recall: MasteryLevel
  conceptual: MasteryLevel
  application: MasteryLevel
  transfer: MasteryLevel
}

export type ReviewCard = {
  id: string
  prompt: string
  answer: string
  objectiveId: string
  objectiveLabel: string
  lastReviewed: string
  hint: string
}

export type ActivityItem = {
  id: string
  label: string
  detail: string
  time: string
  kind: 'review' | 'lesson' | 'exercise' | 'tutor' | 'project'
}

export type Project = {
  id: string
  title: string
  description: string
  status: 'in-progress' | 'not-started' | 'complete'
  repository: string
  objectives: Array<{ label: string; state: 'done' | 'working' | 'todo' }>
  deliverables: string[]
  boundaryNote: string
}

export type MockCheckResult = {
  passed: number
  failed: number
  summary: string
}

