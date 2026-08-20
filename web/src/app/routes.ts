export const DEFAULT_COURSE_ID = 'ai-ml'

export function coursePath(courseId: string) {
  return `/courses/${courseId}`
}

export function modulePath(courseId: string, moduleId: string) {
  return `${coursePath(courseId)}/modules/${moduleId}`
}

export function lessonPath(courseId: string, moduleId: string, lessonId: string) {
  return `${modulePath(courseId, moduleId)}/lessons/${lessonId}`
}

export function worksheetPath(courseId: string, moduleId: string, worksheetId: string) {
  return `${modulePath(courseId, moduleId)}/worksheets/${worksheetId}`
}

export function exercisePath(courseId: string, moduleId: string, exerciseId: string) {
  return `${modulePath(courseId, moduleId)}/exercises/${exerciseId}`
}
