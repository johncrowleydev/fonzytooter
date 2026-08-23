import { useTutor } from './TutorContext'

export function TutorButton() {
  const { isOpen, openTutor } = useTutor()

  if (isOpen) return null

  return (
    <button
      type="button"
      onClick={openTutor}
      className="fixed right-7 bottom-6 z-30 flex items-center gap-2 rounded-full border border-accent-teal/40 bg-brand-teal px-4 py-3 text-sm font-extrabold text-brand-ink shadow-2xl transition hover:-translate-y-px hover:bg-brand-teal-light max-sm:right-4 max-sm:bottom-20 max-sm:px-3 max-sm:py-2.5 max-sm:text-sm"
    >
      <span className="text-brand-ink max-sm:hidden">✦</span> Tutor
    </button>
  )
}
