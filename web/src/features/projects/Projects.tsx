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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
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
      className="w-full rounded-[10px] border border-[var(--line)] bg-[var(--panel)] p-[18px] text-left text-[var(--ink)] hover:border-[var(--line-strong)] hover:bg-[var(--panel-soft)]"
      type="button"
      onClick={() => navigate(`/projects/${project.id}`)}
    >
      <div className="flex items-center justify-between gap-3">
        <Badge tone={project.status === 'in-progress' ? 'coral' : 'neutral'}>
          {project.status === 'in-progress' ? 'In progress' : 'Not started'}
        </Badge>
        <span className="text-right text-base text-[var(--faint)]">→</span>
      </div>
      <h3 className="my-[16px_6px] text-[17px] tracking-[-0.03em]">{project.title}</h3>
      <p className="m-0 max-w-[620px] text-[11px] leading-[1.55] text-[var(--muted)]">
        {project.description}
      </p>
      <div className="mt-[22px] flex items-end justify-between gap-5 max-[640px]:block">
        <div className="w-[180px]">
          <ProgressBar value={percent} tone="coral" />
          <span className="text-[9px] text-[var(--faint)]">
            {done} of {project.objectives.length} objectives
          </span>
        </div>
        <span className="font-mono text-[9px] text-[var(--faint)] max-[640px]:mt-[11px] max-[640px]:block">
          {project.repository.replace('github.com/', '')}
        </span>
      </div>
    </button>
  )
}
