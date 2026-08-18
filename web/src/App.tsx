import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from './app/AppShell'
import { Curriculum } from './features/curriculum/Curriculum'
import { ModuleDetail } from './features/curriculum/ModuleDetail'
import { Dashboard } from './features/dashboard/Dashboard'
import { Exercise } from './features/exercises/Exercise'
import { Lesson } from './features/lessons/Lesson'
import { Progress } from './features/progress/Progress'
import { ProjectDetail } from './features/projects/ProjectDetail'
import { Projects } from './features/projects/Projects'
import { Review } from './features/reviews/Review'
import { TutorProvider } from './features/tutor/TutorContext'

export function App() {
  return (
    <TutorProvider>
      <AppShell>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/curriculum" element={<Curriculum />} />
          <Route path="/curriculum/:moduleId" element={<ModuleDetail />} />
          <Route path="/lesson/:lessonId" element={<Lesson />} />
          <Route path="/review" element={<Review />} />
          <Route path="/exercise/:exerciseId" element={<Exercise />} />
          <Route path="/progress" element={<Progress />} />
          <Route path="/projects" element={<Projects />} />
          <Route path="/projects/:projectId" element={<ProjectDetail />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AppShell>
    </TutorProvider>
  )
}
