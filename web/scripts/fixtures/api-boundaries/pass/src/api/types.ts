import { z } from 'zod'
import { Event } from './generated/event.zod'

export type TutorEventWire = z.infer<typeof Event>
