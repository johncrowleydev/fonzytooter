import { AppShell } from './app/AppShell'
import { Dashboard } from './features/dashboard/Dashboard'
import { TutorProvider } from './features/tutor/TutorContext'

export function App() {
  return (
    <TutorProvider initialPageContext={{ type: 'dashboard' }}>
      <AppShell>
        <Dashboard />
      </AppShell>
    </TutorProvider>
  )
}
