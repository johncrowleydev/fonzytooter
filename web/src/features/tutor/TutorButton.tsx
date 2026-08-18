import { useTutor } from './TutorContext'

export function TutorButton() {
  const { isOpen, openTutor } = useTutor()

  if (isOpen) return null

  return (
    <button
      type="button"
      onClick={openTutor}
      className="tutor-fab"
    >
      <span className="tutor-spark">✦</span> Tutor
    </button>
  )
}
