import type { Objective, Project } from './types'

export type ObjectiveSummary = {
  introduced: number
  recallStrong: number
  applied: number
  transferTested: number
}

export type ProjectProgress = {
  done: number
  total: number
  percent: number
}

export function getObjectiveSummary(objectives: Objective[]): ObjectiveSummary {
  return objectives.reduce(
    (summary, objective) => ({
      introduced: summary.introduced + (objective.introduced ? 1 : 0),
      recallStrong: summary.recallStrong + (objective.recall === 'strong' ? 1 : 0),
      applied: summary.applied + (objective.application !== 'not-assessed' ? 1 : 0),
      transferTested: summary.transferTested + (objective.transfer !== 'not-assessed' ? 1 : 0),
    }),
    { introduced: 0, recallStrong: 0, applied: 0, transferTested: 0 },
  )
}

export function getProjectProgress(project: Project): ProjectProgress {
  const done = project.objectives.filter((objective) => objective.state === 'done').length
  const total = project.objectives.length

  return {
    done,
    total,
    percent: total ? Math.round((done / total) * 100) : 0,
  }
}

export function isReadyToApply(objective: Objective): boolean {
  return (
    objective.introduced &&
    objective.recall === 'strong' &&
    objective.conceptual === 'strong' &&
    objective.application !== 'strong'
  )
}
