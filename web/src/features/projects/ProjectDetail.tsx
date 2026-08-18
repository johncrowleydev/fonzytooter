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
    <div className="page-stack project-detail-page">
      <button className="back-link" onClick={() => navigate('/projects')} type="button">
        ← Projects
      </button>
      <PageIntro compact eyebrow="Project" title={project.title}>
        <div className="intro-actions">
          <Badge tone="coral">
            {project.status === 'in-progress' ? 'In progress' : 'Not started'}
          </Badge>
          <span className="small-muted">
            {done} of {project.objectives.length} objectives
          </span>
        </div>
      </PageIntro>
      <div className="project-detail-layout">
        <main>
          <Card className="repository-card">
            <div className="repo-icon">⌘</div>
            <div>
              <p className="eyebrow">Repository</p>
              <h3>{project.repository}</h3>
            </div>
            <Button variant="secondary">Open repo ↗</Button>
          </Card>
          <Card className="deliverables-card">
            <SectionHeading title="Objectives" />
            <div className="project-objectives">
              {project.objectives.map((objective) => (
                <div key={objective.label} className="project-objective">
                  <StatusDot state={objective.state} />
                  <span>{objective.label}</span>
                  <span className={`objective-state ${objective.state}`}>
                    {objective.state === 'done'
                      ? 'Complete'
                      : objective.state === 'working'
                        ? 'Working on it'
                        : 'Not started'}
                  </span>
                </div>
              ))}
            </div>
            <div className="deliverable-block">
              <p className="eyebrow">Deliverables</p>
              <div className="deliverable-tags">
                {project.deliverables.map((deliverable) => (
                  <Badge key={deliverable} tone="neutral">
                    {deliverable}
                  </Badge>
                ))}
              </div>
            </div>
          </Card>
        </main>
        <aside>
          <Card className="project-progress-card">
            <p className="eyebrow">Progress</p>
            <div className="project-big-number">
              {Math.round((done / project.objectives.length) * 100)}
              <span>%</span>
            </div>
            <ProgressBar value={(done / project.objectives.length) * 100} tone="coral" />
            <p className="small-muted">
              {done} of {project.objectives.length} objectives demonstrated
            </p>
          </Card>
        </aside>
      </div>
    </div>
  )
}
