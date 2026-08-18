import { createContext, useContext, useMemo, useState, type PropsWithChildren } from 'react'
import type { TutorPageContext } from './types'

type TutorContextValue = {
  isOpen: boolean
  pageContext: TutorPageContext
  openTutor: () => void
  openTutorWithContext: (context: TutorPageContext) => void
  closeTutor: () => void
  setPageContext: (context: TutorPageContext) => void
}

const TutorContext = createContext<TutorContextValue | null>(null)

export function TutorProvider({ children }: PropsWithChildren) {
  const [isOpen, setIsOpen] = useState(false)
  const [pageContext, setPageContext] = useState<TutorPageContext>({
    type: 'dashboard',
    title: 'Home',
  })

  const value = useMemo<TutorContextValue>(
    () => ({
      isOpen,
      pageContext,
      openTutor: () => setIsOpen(true),
      openTutorWithContext: (context) => {
        setPageContext(context)
        setIsOpen(true)
      },
      closeTutor: () => setIsOpen(false),
      setPageContext,
    }),
    [isOpen, pageContext],
  )

  return <TutorContext.Provider value={value}>{children}</TutorContext.Provider>
}

export function useTutor() {
  const value = useContext(TutorContext)
  if (!value) {
    throw new Error('useTutor must be used inside TutorProvider')
  }

  return value
}
