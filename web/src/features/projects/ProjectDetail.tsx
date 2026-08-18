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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <button
        className="justify-self-start border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--muted)] hover:text-[var(--ink)]"
        onClick={() => navigate('/projects')}
        type="button"
      >
        ← Projects
      </button>
      <PageIntro compact eyebrow="Project" title={project.title}>
        <div className="mt-[17px] flex items-center gap-[13px] max-[640px]:items-start max-[640px]:flex-col max-[640px]:gap-2">
          <Badge tone="coral">
            {project.status === 'in-progress' ? 'In progress' : 'Not started'}
          </Badge>
          <span className="text-[10px] text-[var(--faint)]">
            {done} of {project.objectives.length} objectives
          </span>
        </div>
      </PageIntro>
      <div className="grid grid-cols-[minmax(0,1fr)_270px] gap-[18px] max-[860px]:grid-cols-1">
        <main className="grid content-start gap-3.5">
          <Card className="flex items-center gap-[13px] max-[640px]:items-start max-[640px]:flex-wrap">
            <div className="grid size-10 place-items-center rounded-[9px] bg-[rgba(118,208,192,0.1)] text-[19px] text-[var(--teal)]">
              ⌘
            </div>
            <div>
              <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
                Repository
              </p>
              <h3 className="my-[7px_4px] font-mono text-xs">{project.repository}</h3>
            </div>
            <Button variant="secondary">Open repo ↗</Button>
          </Card>
          <Card className="p-6">
            <SectionHeading title="Objectives" />
            <div className="grid border-t border-[var(--line)]">
              {project.objectives.map((objective) => (
                <div
                  key={objective.label}
                  className="grid grid-cols-[11px_1fr_auto] items-center gap-2.5 border-b border-[var(--line)] py-3.5 text-[11px]"
                >
                  <StatusDot state={objective.state} />
                  <span>{objective.label}</span>
                  <span
                    className={`text-[9px] ${objective.state === 'done' ? 'text-[var(--teal)]' : objective.state === 'working' ? 'text-[var(--gold)]' : 'text-[var(--faint)]'}`}
                  >
                    {objective.state === 'done'
                      ? 'Complete'
                      : objective.state === 'working'
                        ? 'Working on it'
                        : 'Not started'}
                  </span>
                </div>
              ))}
            </div>
            <div className="mt-[25px]">
              <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
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
            <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
              Progress
            </p>
            <div className="my-[12px_14px] text-[51px] tracking-[-0.08em] text-[var(--coral)]">
              {Math.round((done / project.objectives.length) * 100)}
              <span className="ml-[3px] text-xl text-[var(--faint)]">%</span>
            </div>
            <ProgressBar value={(done / project.objectives.length) * 100} tone="coral" />
            <p className="mt-[11px] block text-[10px] leading-[1.5] text-[var(--faint)]">
              {done} of {project.objectives.length} objectives demonstrated
            </p>
          </Card>
        </aside>
      </div>
    </div>
  )
}
