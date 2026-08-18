import { useEffect } from 'react'
import { navigateTo } from '../../app/navigation'
import { Badge, Card, PageIntro, ProgressBar, SectionHeading } from '../../components/ui'
import { projects } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

export function Projects() {
  const { setPageContext } = useTutor()
  useEffect(() => { setPageContext({ type: 'project', title: 'Projects' }) }, [setPageContext])

  return <div className="page-stack projects-page"><PageIntro compact title="Projects" /><div className="project-page-grid"><div><SectionHeading title="Labs" /><div className="project-list">{projects.map((project) => <ProjectRow key={project.id} project={project} />)}</div></div></div></div>
}

function ProjectRow({ project }: { project: (typeof projects)[number] }) {
  const done = project.objectives.filter((item) => item.state === 'done').length
  const percent = Math.round((done / project.objectives.length) * 100)
  return <button className="project-row" type="button" onClick={() => navigateTo(`/projects/${project.id}`)}><div className="project-row-top"><Badge tone={project.status === 'in-progress' ? 'coral' : 'neutral'}>{project.status === 'in-progress' ? 'In progress' : 'Not started'}</Badge><span className="row-arrow">→</span></div><h3>{project.title}</h3><p>{project.description}</p><div className="project-row-bottom"><div><ProgressBar value={percent} tone="coral" /><span>{done} of {project.objectives.length} objectives</span></div><span className="repo-short">{project.repository.replace('github.com/', '')}</span></div></button>
}
