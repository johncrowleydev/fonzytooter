import type { PropsWithChildren } from 'react'
import { TutorButton } from '../features/tutor/TutorButton'
import { TutorOverlay } from '../features/tutor/TutorOverlay'

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-950/95">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-5 py-4">
          <div>
            <div className="text-lg font-semibold tracking-tight">Fonzytooter</div>
            <div className="text-xs text-slate-400">Personal AI/ML learning system</div>
          </div>
          <div className="text-xs text-slate-500">self-paced · one learner</div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-5 py-8">{children}</main>

      <TutorButton />
      <TutorOverlay />
    </div>
  )
}
