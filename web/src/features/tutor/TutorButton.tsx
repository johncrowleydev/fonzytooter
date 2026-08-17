import { useTutor } from './TutorContext'

export function TutorButton() {
  const { isOpen, openTutor } = useTutor()

  if (isOpen) return null

  return (
    <button
      type="button"
      onClick={openTutor}
      className="fixed bottom-5 right-5 rounded-full bg-slate-100 px-5 py-3 text-sm font-semibold text-slate-950 shadow-xl shadow-black/30 transition hover:bg-white"
    >
      Ask tutor
    </button>
  )
}
