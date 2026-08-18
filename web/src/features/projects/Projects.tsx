import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Card, PageIntro, ProgressBar, SectionHeading } from '../../components/ui'
import { projects } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

export function Projects() {
  const { setPageContext } = useTutor()
  useEffect(() => {
    setPageContext({ type: 'project', title: 'Projects' })
  }, [setPageContext])

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Projects" />
      <div className="grid grid-cols-1 gap-10">
        <div>
          <SectionHeading title="Labs" />
          <div className="grid gap-2.5">
            {projects.map((project) => (
              <ProjectRow key={project.id} project={project} />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function ProjectRow({ project }: { project: (typeof projects)[number] }) {
  const navigate = useNavigate()
  const done = project.objectives.filter((item) => item.state === 'done').length
  const percent = Math.round((done / project.objectives.length) * 100)
  return (
    <button
      className="w-full rounded-lg border border-line bg-panel p-5 text-left text-ink hover:border-line-strong hover:bg-panel-soft"
      type="button"
      onClick={() => navigate(`/projects/${project.id}`)}
    >
      <div className="flex items-center justify-between gap-3">
        <Badge tone={project.status === 'in-progress' ? 'coral' : 'neutral'}>
          {project.status === 'in-progress' ? 'In progress' : 'Not started'}
        </Badge>
        <span className="text-right text-base text-faint">→</span>
      </div>
      <h3 className="my-4 text-lg tracking-tight">{project.title}</h3>
      <p className="m-0 max-w-2xl text-xs leading-relaxed text-muted">{project.description}</p>
      <div className="mt-6 flex items-end justify-between gap-5 max-sm:block">
        <div className="w-44">
          <ProgressBar value={percent} tone="coral" />
          <span className="text-2xs text-faint">
            {done} of {project.objectives.length} objectives
          </span>
        </div>
        <span className="font-mono text-2xs text-faint max-sm:mt-3 max-sm:block">
          {project.repository.replace('github.com/', '')}
        </span>
      </div>
    </button>
  )
}
