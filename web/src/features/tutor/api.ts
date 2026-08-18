import type { TutorEvent, TutorMode, TutorPageContext } from './types'

type StreamTutorTurnRequest = {
  message: string
  mode: TutorMode
  pageContext: TutorPageContext
  onEvent: (event: TutorEvent) => void
  signal?: AbortSignal
}

export async function getMockTutorResponse({ message, mode, pageContext }: Omit<StreamTutorTurnRequest, 'onEvent' | 'signal'>) {
  await new Promise((resolve) => window.setTimeout(resolve, 480))

  const contextName = pageContext.lessonTitle ?? pageContext.exerciseTitle ?? pageContext.projectTitle ?? pageContext.objectiveTitle ?? pageContext.title ?? pageContext.type
  const selected = pageContext.selectedText ? ` You also highlighted “${pageContext.selectedText.slice(0, 88)}${pageContext.selectedText.length > 88 ? '…' : ''}”.` : ''

  if (pageContext.type === 'exercise') {
    if (mode === 'socratic') return `Let’s isolate the behavior first. In ${contextName}, what does one update do to the loss? Look at the direction of the gradient and the learning rate before changing the loop.${selected}`
    if (pageContext.lastExecution?.failed) return `Your mock check shows ${pageContext.lastExecution.failed} failing test. The likely question is whether the update is moving with the gradient or against it. Try tracing one step on f(x) = x², then compare the loss before and after.${selected}`
    return `This exercise is about making the optimization loop legible: compute a gradient, scale it by the learning rate, and update without mutating the input. I can help you reason through each part without handing over a finished solution.${selected}`
  }

  if (pageContext.type === 'progress') return `Your mock progress is stronger in recall than in application for ${contextName}. A useful next move would be a small implementation or worked example, then a transfer question that changes the surface details.`
  if (pageContext.type === 'review') return `For this review, try retrieving the idea before looking for perfect wording. A compact explanation plus one concrete example is stronger evidence than memorizing a sentence.`
  if (pageContext.type === 'project') return `For ${contextName}, use the repository, tests, and notes to track the work.${selected}`
  if (mode === 'socratic') return `You’re looking at ${contextName}. Start by naming the quantity that changes, then ask what a small change in each input would do to the output. What local relationship do you think the diagram is encoding?`
  if (mode === 'quiz') return `Quick check on ${contextName}: if you changed the learning rate while keeping the gradient fixed, what would change about the next step? Explain the tradeoff in your own words.`
  if (mode === 'explore') return `A useful connection from ${contextName} is that optimization is a repeated local approximation. The same idea appears in numerical methods, control, and even how we choose experiments: measure a direction, take a bounded step, and inspect what happened.`
  return `You’re currently working in ${contextName}. The key idea to keep in view is the relationship between a local signal and the next action. If you tell me which part feels slippery, I’ll anchor the explanation to this screen.${selected}`
}
