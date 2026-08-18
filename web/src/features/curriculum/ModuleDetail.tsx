import { useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge, Card, PageIntro, ProgressBar, SectionHeading, StatusDot } from '../../components/ui'
import { modules, objectives } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

const lessonToneByKind = {
  exercise: 'coral',
  video: 'violet',
  lab: 'gold',
  lesson: 'teal',
} as const

export function ModuleDetail() {
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
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <Link
        className="justify-self-start text-xs font-bold text-muted no-underline hover:text-ink"
        to="/curriculum"
      >
        ← Curriculum
      </Link>
      <PageIntro
        compact
        eyebrow={`${module.eyebrow} / Module`}
        title={module.title}
        detail={module.description}
      >
        <div className="mt-4 flex items-center gap-3 max-sm:items-start max-sm:flex-col max-sm:gap-2">
          <Badge tone={module.accent as 'teal' | 'gold' | 'coral' | 'violet' | 'blue'}>
            {module.status === 'in-progress' ? 'In progress' : module.status}
          </Badge>
          <span className="text-2xs text-faint">
            {module.objectiveIds.length} objectives · {module.lessons.length} learning items
          </span>
        </div>
      </PageIntro>
      <section className="grid grid-cols-[1.35fr_0.65fr] gap-3.5 max-lg:grid-cols-1">
        <Card>
          <SectionHeading
            eyebrow="Progress"
            title={`${completed} of ${module.lessons.length} complete`}
          />
          <div className="grid grid-cols-[130px_1fr] items-center gap-4">
            <div className="relative h-24 w-28">
              <span className="absolute top-[26px] left-[38px] z-2 grid size-10 place-items-center rounded-full bg-brand-coral font-serif text-xl font-semibold text-brand-ink">
                ∂
              </span>
              <span className="absolute inset-[25px_5px] rotate-[25deg] rounded-full border border-brand-coral/30" />
              <span className="absolute inset-[8px_25px] rotate-[65deg] rounded-full border border-brand-teal/30" />
              <span className="absolute inset-[33px_22px_21px] rounded-full border border-brand-gold/30" />
            </div>
            <div>
              <ProgressBar
                value={module.lessons.length ? (completed / module.lessons.length) * 100 : 0}
                tone="coral"
              />
              <div className="mt-2 mb-2 flex justify-between gap-2">
                <span className="text-xs font-bold text-ink">{completed} complete</span>
                <span className="text-2xs text-faint">
                  {module.lessons.length - completed} remaining
                </span>
              </div>
            </div>
          </div>
        </Card>
        <Card muted>
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">Prerequisite</p>
          <h3 className="my-2 text-lg leading-tight tracking-tight">
            {module.prerequisites?.[0] ?? 'None'}
          </h3>
          {module.prerequisites ? (
            <Link
              className="text-xs font-bold text-brand-teal no-underline hover:text-ink"
              to="/curriculum/mathematical-foundations"
            >
              Review foundation →
            </Link>
          ) : (
            <span className="text-xs font-bold text-brand-teal">Start here →</span>
          )}
        </Card>
      </section>
      <section className="grid grid-cols-2 gap-12 max-xl:grid-cols-1">
        <div>
          <SectionHeading title="Objectives" />
          <div className="grid gap-2">
            {moduleObjectives.length ? (
              moduleObjectives.map((objective) => (
                <ObjectiveRow key={objective.id} objective={objective} />
              ))
            ) : (
              <Card muted>
                <p className="text-xs leading-relaxed text-muted">No objectives yet.</p>
              </Card>
            )}
          </div>
        </div>
        <div>
          <SectionHeading title="Lessons & practice" />
          <div className="grid">
            {module.lessons.map((lesson, index) => (
              <LessonRow key={lesson.id} lesson={lesson} index={index} />
            ))}
          </div>
          <div className="mt-8 border-t border-line pt-5">
            <h3 className="mb-3 text-xs">Curated resources</h3>
            <div className="grid grid-cols-[68px_1fr_18px] items-center gap-2.5 rounded-lg border border-line p-2">
              <div className="grid h-12 place-items-center rounded bg-gradient-to-br from-brand-violet/30 to-brand-teal/15 text-ink">
                <span>▶</span>
              </div>
              <div>
                <Badge tone="coral">Video</Badge>
                <strong className="mt-1 mb-1 block text-2xs">
                  Visual intuition for backpropagation
                </strong>
                <span className="block text-2xs text-faint">3Blue1Brown · supports chain rule</span>
              </div>
              <button className="border-0 bg-transparent p-0 text-xs text-faint">↗</button>
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
    <div className="grid grid-cols-[11px_1fr_auto] items-center gap-3 border-t border-line py-3">
      <StatusDot state={strength === 'strong' ? 'completed' : 'in-progress'} />
      <div>
        <strong className="block text-xs">{objective.title}</strong>
        <span className="mt-1 block text-2xs leading-normal text-muted">
          {objective.description}
        </span>
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
  const tone = lessonToneByKind[lesson.kind]
  const path =
    lesson.kind === 'exercise'
      ? `/exercise/${lesson.id}`
      : lesson.kind === 'lab'
        ? '/projects/nn-scratch'
        : `/lesson/${lesson.id}`
  return (
    <Link
      className="grid w-full grid-cols-[24px_62px_minmax(0,1fr)_17px] items-center gap-2 border-0 border-t border-line bg-transparent py-3 text-left text-ink no-underline max-sm:grid-cols-[24px_62px_minmax(0,1fr)_17px]"
      to={path}
    >
      <span className="text-2xs text-faint">{String(index + 1).padStart(2, '0')}</span>
      <span>
        <Badge tone={tone}>{lesson.kind}</Badge>
      </span>
      <span>
        <strong className="block text-xs">{lesson.title}</strong>
        <small className="mt-1 block text-2xs text-faint">
          {lesson.completed
            ? 'Completed'
            : lesson.kind === 'video'
              ? 'Curated resource'
              : lesson.kind === 'lab'
                ? 'Repository-based lab'
                : 'Lesson'}
        </small>
      </span>
      <span className={`text-right text-faint ${lesson.completed ? 'text-brand-teal' : ''}`}>
        {lesson.completed ? '✓' : '→'}
      </span>
    </Link>
  )
}
