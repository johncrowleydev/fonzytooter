import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  Badge,
  Button,
  Card,
  PageIntro,
  ProgressBar,
  SectionHeading,
  StatusDot,
} from '../../components/ui'
import { projects } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

const objectiveStateStyles = {
  done: 'text-brand-teal',
  working: 'text-brand-gold',
  todo: 'text-faint',
} as const

export function ProjectDetail() {
  const navigate = useNavigate()
  const { projectId = 'nn-scratch' } = useParams()
  const { setPageContext } = useTutor()
  const project = projects.find((item) => item.id === projectId) ?? projects[0]
  const done = project.objectives.filter((item) => item.state === 'done').length
  useEffect(() => {
    setPageContext({
      type: 'project',
      title: project.title,
      projectId: project.id,
      projectTitle: project.title,
      objectiveIds: ['nn.neuron', 'nn.backpropagation'],
    })
  }, [project.id, project.title, projectId, setPageContext])

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <button
        className="justify-self-start border-0 bg-transparent p-0 text-xs font-bold text-muted hover:text-ink"
        onClick={() => navigate('/projects')}
        type="button"
      >
        ← Projects
      </button>
      <PageIntro compact eyebrow="Project" title={project.title}>
        <div className="mt-4 flex items-center gap-3 max-sm:items-start max-sm:flex-col max-sm:gap-2">
          <Badge tone="coral">
            {project.status === 'in-progress' ? 'In progress' : 'Not started'}
          </Badge>
          <span className="text-2xs text-faint">
            {done} of {project.objectives.length} objectives
          </span>
        </div>
      </PageIntro>
      <div className="grid grid-cols-[minmax(0,1fr)_270px] gap-5 max-lg:grid-cols-1">
        <main className="grid content-start gap-3.5">
          <Card className="flex items-center gap-3 max-sm:items-start max-sm:flex-wrap">
            <div className="grid size-10 place-items-center rounded-lg bg-brand-teal/10 text-xl text-brand-teal">
              ⌘
            </div>
            <div>
              <p className="text-2xs font-bold uppercase tracking-widest text-faint">Repository</p>
              <h3 className="my-2 font-mono text-xs">{project.repository}</h3>
            </div>
            <Button variant="secondary">Open repo ↗</Button>
          </Card>
          <Card className="p-6">
            <SectionHeading title="Objectives" />
            <div className="grid border-t border-line">
              {project.objectives.map((objective) => (
                <div
                  key={objective.label}
                  className="grid grid-cols-[11px_1fr_auto] items-center gap-2.5 border-b border-line py-3.5 text-xs"
                >
                  <StatusDot state={objective.state} />
                  <span>{objective.label}</span>
                  <span className={`text-2xs ${objectiveStateStyles[objective.state]}`}>
                    {objective.state === 'done'
                      ? 'Complete'
                      : objective.state === 'working'
                        ? 'Working on it'
                        : 'Not started'}
                  </span>
                </div>
              ))}
            </div>
            <div className="mt-6">
              <p className="text-2xs font-bold uppercase tracking-widest text-faint">
                Deliverables
              </p>
              <div className="mt-2.5 flex flex-wrap gap-1.5">
                {project.deliverables.map((deliverable) => (
                  <Badge key={deliverable} tone="neutral">
                    {deliverable}
                  </Badge>
                ))}
              </div>
            </div>
          </Card>
        </main>
        <aside className="grid content-start gap-3.5">
          <Card className="text-left">
            <p className="text-2xs font-bold uppercase tracking-widest text-faint">Progress</p>
            <div className="my-3 text-5xl tracking-tight text-brand-coral">
              {Math.round((done / project.objectives.length) * 100)}
              <span className="ml-1 text-xl text-faint">%</span>
            </div>
            <ProgressBar value={(done / project.objectives.length) * 100} tone="coral" />
            <p className="mt-3 block text-2xs leading-normal text-faint">
              {done} of {project.objectives.length} objectives demonstrated
            </p>
          </Card>
        </aside>
      </div>
    </div>
  )
}
