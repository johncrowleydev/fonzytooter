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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <button
        className="justify-self-start border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--muted)] hover:text-[var(--ink)]"
        onClick={() => navigate('/curriculum')}
        type="button"
      >
        ← Curriculum
      </button>
      <PageIntro
        compact
        eyebrow={`${module.eyebrow} / Module`}
        title={module.title}
        detail={module.description}
      >
        <div className="mt-[17px] flex items-center gap-[13px] max-[640px]:items-start max-[640px]:flex-col max-[640px]:gap-2">
          <Badge tone={module.accent as 'teal' | 'gold' | 'coral' | 'violet' | 'blue'}>
            {module.status === 'in-progress' ? 'In progress' : module.status}
          </Badge>
          <span className="text-[10px] text-[var(--faint)]">
            {module.objectiveIds.length} objectives · {module.lessons.length} learning items
          </span>
        </div>
      </PageIntro>
      <section className="grid grid-cols-[1.35fr_0.65fr] gap-3.5 max-[860px]:grid-cols-1">
        <Card>
          <SectionHeading
            eyebrow="Progress"
            title={`${completed} of ${module.lessons.length} complete`}
          />
          <div className="grid grid-cols-[130px_1fr] items-center gap-[17px]">
            <div className="relative h-[90px] w-28">
              <span className="absolute top-[26px] left-[38px] z-[2] grid size-[38px] place-items-center rounded-full bg-[var(--coral)] font-serif text-xl font-semibold text-[#0c1721]">
                ∂
              </span>
              <span className="absolute inset-[25px_5px] rotate-[25deg] rounded-[50%] border border-[rgba(239,145,110,0.28)]" />
              <span className="absolute inset-[8px_25px] rotate-[65deg] rounded-[50%] border border-[rgba(118,208,192,0.28)]" />
              <span className="absolute inset-[33px_22px_21px] rounded-[50%] border border-[rgba(225,184,106,0.3)]" />
            </div>
            <div>
              <ProgressBar
                value={module.lessons.length ? (completed / module.lessons.length) * 100 : 0}
                tone="coral"
              />
              <div className="mt-[9px] mb-[7px] flex justify-between gap-2">
                <span className="text-[11px] font-bold text-[var(--ink)]">
                  {completed} complete
                </span>
                <span className="text-[9px] text-[var(--faint)]">
                  {module.lessons.length - completed} remaining
                </span>
              </div>
            </div>
          </div>
        </Card>
        <Card muted>
          <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
            Prerequisite
          </p>
          <h3 className="my-[9px] text-[18px] leading-tight tracking-[-0.035em]">
            {module.prerequisites?.[0] ?? 'None'}
          </h3>
          <button
            className="border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--teal)] hover:text-[var(--ink)]"
            onClick={() => module.prerequisites && navigate('/curriculum/mathematical-foundations')}
          >
            {module.prerequisites ? 'Review foundation →' : 'Start here →'}
          </button>
        </Card>
      </section>
      <section className="grid grid-cols-2 gap-12 max-[1120px]:grid-cols-1">
        <div>
          <SectionHeading title="Objectives" />
          <div className="grid gap-[7px]">
            {moduleObjectives.length ? (
              moduleObjectives.map((objective) => (
                <ObjectiveRow key={objective.id} objective={objective} />
              ))
            ) : (
              <Card muted>
                <p className="text-xs leading-[1.65] text-[var(--muted)]">No objectives yet.</p>
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
          <div className="mt-8 border-t border-[var(--line)] pt-5">
            <h3 className="mb-3 text-xs">Curated resources</h3>
            <div className="grid grid-cols-[68px_1fr_18px] items-center gap-2.5 rounded-lg border border-[var(--line)] p-2">
              <div className="grid h-[45px] place-items-center rounded-[5px] bg-[linear-gradient(135deg,rgba(169,155,231,0.28),rgba(118,208,192,0.15))] text-[var(--ink)]">
                <span>▶</span>
              </div>
              <div>
                <Badge tone="coral">Video</Badge>
                <strong className="mt-[5px] mb-[3px] block text-[10px]">
                  Visual intuition for backpropagation
                </strong>
                <span className="block text-[10px] text-[var(--faint)]">
                  3Blue1Brown · supports chain rule
                </span>
              </div>
              <button className="border-0 bg-transparent p-0 text-xs text-[var(--faint)]">↗</button>
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
    <div className="grid grid-cols-[11px_1fr_auto] items-center gap-[11px] border-t border-[var(--line)] py-[13px]">
      <StatusDot state={strength === 'strong' ? 'completed' : 'in-progress'} />
      <div>
        <strong className="block text-[11px]">{objective.title}</strong>
        <span className="mt-[3px] block text-[10px] leading-[1.35] text-[var(--muted)]">
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
    <button
      type="button"
      className="grid w-full grid-cols-[24px_62px_minmax(0,1fr)_17px] items-center gap-[9px] border-0 border-t border-[var(--line)] bg-transparent py-[13px] text-left text-[var(--ink)] max-[640px]:grid-cols-[24px_62px_minmax(0,1fr)_17px]"
      onClick={() => navigate(path)}
    >
      <span className="text-[10px] text-[var(--faint)]">{String(index + 1).padStart(2, '0')}</span>
      <span>
        <Badge tone={tone}>{lesson.kind}</Badge>
      </span>
      <span>
        <strong className="block text-[11px]">{lesson.title}</strong>
        <small className="mt-[3px] block text-[9px] text-[var(--faint)]">
          {lesson.completed
            ? 'Completed'
            : lesson.kind === 'video'
              ? 'Curated resource'
              : lesson.kind === 'lab'
                ? 'Repository-based lab'
                : 'Lesson'}
        </small>
      </span>
      <span
        className={`text-right text-[var(--faint)] ${lesson.completed ? 'text-[var(--teal)]' : ''}`}
      >
        {lesson.completed ? '✓' : '→'}
      </span>
    </button>
  )
}
