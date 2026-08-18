import { z } from 'zod'
import { Event, PageContext, TurnRequest } from '../../api/generated/schemas'

export type TutorMode = NonNullable<z.input<typeof TurnRequest>['mode']>
export type TutorPageContext = z.input<typeof PageContext>
export type TutorEvent = z.output<typeof Event>

export type TutorMessage = {
  id: string
  role: 'user' | 'assistant'
  text: string
}
