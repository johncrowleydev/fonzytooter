import { useEffect } from 'react'
import { AppShell } from './app/AppShell'
import { useCurrentPath, navigateTo } from './app/navigation'
import { Dashboard } from './features/dashboard/Dashboard'
import { Curriculum } from './features/curriculum/Curriculum'
import { ModuleDetail } from './features/curriculum/ModuleDetail'
import { Lesson } from './features/lessons/Lesson'
import { Review } from './features/reviews/Review'
import { Exercise } from './features/exercises/Exercise'
import { Progress } from './features/progress/Progress'
import { Projects } from './features/projects/Projects'
import { ProjectDetail } from './features/projects/ProjectDetail'
import { TutorProvider } from './features/tutor/TutorContext'
import type { TutorPageContext } from './features/tutor/types'

export function App() {
  const path = useCurrentPath()

  useEffect(() => {
    if (path !== '/' && !path.startsWith('/curriculum') && !path.startsWith('/lesson') && !path.startsWith('/review') && !path.startsWith('/exercise') && !path.startsWith('/progress') && !path.startsWith('/projects')) navigateTo('/')
  }, [path])

  return (
    <TutorProvider initialPageContext={getInitialContext(path)}>
      <AppShell currentPath={path}><RouterView path={path} /></AppShell>
    </TutorProvider>
  )
}

function RouterView({ path }: { path: string }) {
  if (path === '/') return <Dashboard />
  if (path === '/curriculum') return <Curriculum />
  if (path.startsWith('/curriculum/')) return <ModuleDetail moduleId={path.split('/')[2] ?? 'neural-networks'} />
  if (path.startsWith('/lesson/')) return <Lesson lessonId={path.split('/')[2] ?? 'backpropagation'} />
  if (path === '/review') return <Review />
  if (path.startsWith('/exercise/')) return <Exercise exerciseId={path.split('/')[2] ?? 'gradient-descent-exercise'} />
  if (path === '/progress') return <Progress />
  if (path === '/projects') return <Projects />
  if (path.startsWith('/projects/')) return <ProjectDetail projectId={path.split('/')[2] ?? 'nn-scratch'} />
  return <Dashboard />
}

function getInitialContext(path: string): TutorPageContext {
  if (path.startsWith('/lesson/')) return { type: 'lesson', title: 'Backpropagation', lessonId: path.split('/')[2], lessonTitle: 'Backpropagation', moduleId: 'neural-networks', moduleTitle: 'Neural Networks From Scratch', objectiveIds: ['nn.chain-rule', 'nn.backpropagation'] }
  if (path.startsWith('/exercise/')) return { type: 'exercise', title: 'Implement gradient descent', exerciseId: path.split('/')[2], exerciseTitle: 'Implement gradient descent', objectiveIds: ['nn.backpropagation'] }
  if (path.startsWith('/curriculum/')) return { type: 'curriculum', title: 'Curriculum', moduleId: path.split('/')[2] }
  if (path === '/review') return { type: 'review', title: 'Review' }
  if (path === '/progress') return { type: 'progress', title: 'Objective progress' }
  if (path.startsWith('/projects/')) return { type: 'project', title: 'Projects', projectId: path.split('/')[2], projectTitle: 'Neural Network From Scratch' }
  if (path === '/projects') return { type: 'project', title: 'Projects' }
  if (path === '/curriculum') return { type: 'curriculum', title: 'Curriculum' }
  return { type: 'dashboard', title: 'Home' }
}

