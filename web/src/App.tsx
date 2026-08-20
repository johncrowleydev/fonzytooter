import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from './app/AppShell'
import { coursePath, DEFAULT_COURSE_ID } from './app/routes'
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
import { Worksheet } from './features/worksheets/Worksheet'

export function App() {
  return (
    <TutorProvider>
      <AppShell>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route
            path="/curriculum"
            element={<Navigate to={coursePath(DEFAULT_COURSE_ID)} replace />}
          />
          <Route path="/courses/:courseId" element={<Curriculum />} />
          <Route path="/courses/:courseId/modules/:moduleId" element={<ModuleDetail />} />
          <Route
            path="/courses/:courseId/modules/:moduleId/lessons/:lessonId"
            element={<Lesson />}
          />
          <Route
            path="/courses/:courseId/modules/:moduleId/worksheets/:worksheetId"
            element={<Worksheet />}
          />
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
