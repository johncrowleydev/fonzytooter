import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Badge, Card, PageIntro, ProgressBar, SectionHeading, StatusDot } from '../../components/ui'
import { modules, objectives } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

export function ModuleDetail() {
  const navigate = useNavigate()
  const { moduleId = 'neural-networks' } = useParams()
  const { setPageContext } = useTutor()
  const module =
    modules.find((item) => item.id === moduleId) ??
    modules.find((item) => item.id === 'neural-networks')!
  const moduleObjectives = objectives.filter((objective) =>
    module.objectiveIds.includes(objective.id),
  )
  const completed = module.lessons.filter((lesson) => lesson.completed).length

  useEffect(() => {
    setPageContext({
      type: 'curriculum',
      title: module.title,
      moduleId: module.id,
      moduleTitle: module.title,
      objectiveIds: module.objectiveIds,
    })
  }, [module.id, module.title, module.objectiveIds, setPageContext])

  return (
    <div className="page-stack module-detail-page">
      <button className="back-link" onClick={() => navigate('/curriculum')} type="button">
        ← Curriculum
      </button>
      <PageIntro
        compact
        eyebrow={`${module.eyebrow} / Module`}
        title={module.title}
        detail={module.description}
      >
        <div className="intro-actions">
          <Badge tone={module.accent as 'teal' | 'gold' | 'coral' | 'violet' | 'blue'}>
            {module.status === 'in-progress' ? 'In progress' : module.status}
          </Badge>
          <span className="small-muted">
            {module.objectiveIds.length} objectives · {module.lessons.length} learning items
          </span>
        </div>
      </PageIntro>
      <section className="module-overview-grid">
        <Card>
          <SectionHeading
            eyebrow="Progress"
            title={`${completed} of ${module.lessons.length} complete`}
          />
          <div className="module-signal">
            <div className="signal-graphic">
              <span className="signal-core">∂</span>
              <span className="signal-orbit one" />
              <span className="signal-orbit two" />
              <span className="signal-orbit three" />
            </div>
            <div>
              <ProgressBar
                value={module.lessons.length ? (completed / module.lessons.length) * 100 : 0}
                tone="coral"
              />
              <div className="progress-copy wide">
                <span>{completed} complete</span>
                <span>{module.lessons.length - completed} remaining</span>
              </div>
            </div>
          </div>
        </Card>
        <Card muted>
          <p className="eyebrow">Prerequisite</p>
          <h3>{module.prerequisites?.[0] ?? 'None'}</h3>
          <button
            className="text-link"
            onClick={() => module.prerequisites && navigate('/curriculum/mathematical-foundations')}
          >
            {module.prerequisites ? 'Review foundation →' : 'Start here →'}
          </button>
        </Card>
      </section>
      <section className="content-two-col">
        <div>
          <SectionHeading title="Objectives" />
          <div className="objective-list">
            {moduleObjectives.length ? (
              moduleObjectives.map((objective) => (
                <ObjectiveRow key={objective.id} objective={objective} />
              ))
            ) : (
              <Card muted>
                <p className="body-muted">No objectives yet.</p>
              </Card>
            )}
          </div>
        </div>
        <div>
          <SectionHeading title="Lessons & practice" />
          <div className="lesson-list">
            {module.lessons.map((lesson, index) => (
              <LessonRow key={lesson.id} lesson={lesson} index={index} />
            ))}
          </div>
          <div className="module-resources">
            <h3>Curated resources</h3>
            <div className="resource-card">
              <div className="video-thumb">
                <span>▶</span>
              </div>
              <div>
                <Badge tone="coral">Video</Badge>
                <strong>Visual intuition for backpropagation</strong>
                <span>3Blue1Brown · supports chain rule</span>
              </div>
              <button className="icon-button">↗</button>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

function ObjectiveRow({ objective }: { objective: (typeof objectives)[number] }) {
  const strength =
    objective.application === 'strong'
      ? 'strong'
      : objective.conceptual === 'developing'
        ? 'developing'
        : 'strong'
  return (
    <div className="objective-row">
      <StatusDot state={strength === 'strong' ? 'completed' : 'in-progress'} />
      <div>
        <strong>{objective.title}</strong>
        <span>{objective.description}</span>
      </div>
      <Badge tone={strength === 'strong' ? 'teal' : 'gold'}>
        {strength === 'strong' ? 'Strong' : 'Developing'}
      </Badge>
    </div>
  )
}

function LessonRow({
  lesson,
  index,
}: {
  lesson: (typeof modules)[number]['lessons'][number]
  index: number
}) {
  const navigate = useNavigate()
  const tone =
    lesson.kind === 'exercise'
      ? 'coral'
      : lesson.kind === 'video'
        ? 'violet'
        : lesson.kind === 'lab'
          ? 'gold'
          : 'teal'
  const path =
    lesson.kind === 'exercise'
      ? `/exercise/${lesson.id}`
      : lesson.kind === 'lab'
        ? '/projects/nn-scratch'
        : `/lesson/${lesson.id}`
  return (
    <button type="button" className="lesson-row" onClick={() => navigate(path)}>
      <span className="lesson-number">{String(index + 1).padStart(2, '0')}</span>
      <span className="lesson-kind">
        <Badge tone={tone}>{lesson.kind}</Badge>
      </span>
      <span className="lesson-name">
        <strong>{lesson.title}</strong>
        <small>
          {lesson.completed
            ? 'Completed'
            : lesson.kind === 'video'
              ? 'Curated resource'
              : lesson.kind === 'lab'
                ? 'Repository-based lab'
                : 'Lesson'}
        </small>
      </span>
      <span className={`lesson-state ${lesson.completed ? 'complete' : ''}`}>
        {lesson.completed ? '✓' : '→'}
      </span>
    </button>
  )
}
