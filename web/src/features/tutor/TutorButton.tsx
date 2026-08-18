import { useTutor } from './TutorContext'

export function TutorButton() {
  const { isOpen, openTutor } = useTutor()

  if (isOpen) return null

  return (
    <button
      type="button"
      onClick={openTutor}
      className="fixed right-7 bottom-[26px] z-30 flex items-center gap-2 rounded-full border border-[rgba(118,208,192,0.34)] bg-[var(--teal)] px-[15px] py-[11px] text-[11px] font-extrabold text-[#0a171e] shadow-[0_12px_30px_rgba(0,0,0,0.3)] transition hover:-translate-y-px hover:bg-[#9ce1d5] max-[640px]:right-[15px] max-[640px]:bottom-[75px] max-[640px]:px-3 max-[640px]:py-2.5 max-[640px]:text-[10px]"
    >
      <span className="text-[#0a171e] max-[640px]:hidden">✦</span> Tutor
    </button>
  )
}
