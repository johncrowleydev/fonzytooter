import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, Card, PageIntro, ProgressBar, SectionHeading, StatusDot } from '../../components/ui'
import { modules } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

const moduleStatusStyles = {
  'in-progress': 'text-brand-gold',
  completed: 'text-brand-teal',
  available: 'text-faint',
  locked: 'text-faint',
} as const

const moduleStatusLabels = {
  'in-progress': 'Current',
  completed: 'Complete',
  available: 'Available',
  locked: 'Locked',
} as const

export function Curriculum() {
  const { setPageContext } = useTutor()
  const [query, setQuery] = useState('')
  useEffect(() => {
    setPageContext({ type: 'curriculum', title: 'Curriculum' })
  }, [setPageContext])
  const filteredModules = useMemo(
    () =>
      modules.filter((module) =>
        `${module.title} ${module.description}`.toLowerCase().includes(query.toLowerCase()),
      ),
    [query],
  )
  const groups = ['Foundations', 'Neural networks', 'Modern AI']

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Curriculum" />
      <div className="flex items-center justify-between gap-4 max-sm:items-stretch max-sm:flex-col">
        <div className="flex w-full max-w-sm items-center gap-2 rounded-lg border border-line bg-panel-soft px-3 py-2 text-faint max-sm:max-w-none">
          <span>⌕</span>
          <input
            className="w-full border-0 bg-transparent text-xs text-ink outline-0 placeholder:text-faint"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search modules or concepts"
          />
        </div>
        <div className="flex gap-1 rounded-md border border-line p-1 max-sm:self-start">
          <button className="rounded px-2.5 py-1.5 text-2xs text-ink bg-brand-slate/15">
            List
          </button>
          <button className="rounded border-0 bg-transparent px-2.5 py-1.5 text-2xs text-faint">
            Compact
          </button>
        </div>
      </div>
      <div className="grid grid-cols-[minmax(0,1fr)_270px] gap-10 max-xl:grid-cols-1">
        <div className="grid gap-8 max-sm:gap-5">
          {groups.map((group) => {
            const groupModules = filteredModules.filter((module) => module.eyebrow === group)
            if (!groupModules.length) return null
            return (
              <section key={group} className="grid">
                <SectionHeading title={group} detail={`${groupModules.length} modules`} />
                <div className="grid gap-2.5">
                  {groupModules.map((module) => (
                    <ModuleRow key={module.id} module={module} />
                  ))}
                </div>
              </section>
            )
          })}
        </div>
        <Card className="sticky top-8 mt-9 self-start max-xl:hidden">
          <p className="text-2xs font-bold uppercase tracking-widest text-faint">Status</p>
          <div className="mt-6 grid gap-3">
            <div className="flex items-center gap-2.5 text-xs text-muted">
              <StatusDot state="completed" />
              <span>Complete</span>
            </div>
            <div className="flex items-center gap-2.5 text-xs text-muted">
              <StatusDot state="in-progress" />
              <span>In progress</span>
            </div>
            <div className="flex items-center gap-2.5 text-xs text-muted">
              <StatusDot state="available" />
              <span>Available</span>
            </div>
            <div className="flex items-center gap-2.5 text-xs text-muted">
              <StatusDot state="locked" />
              <span>Locked</span>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}

function ModuleRow({ module }: { module: (typeof modules)[number] }) {
  const complete = module.lessons.filter((lesson) => lesson.completed).length
  const total = module.lessons.length
  const percent = total
    ? Math.round((complete / total) * 100)
    : module.status === 'completed'
      ? 100
      : 0
  const tone = module.accent as 'teal' | 'gold' | 'coral' | 'violet' | 'blue'
  return (
    <Link
      className="grid w-full grid-cols-[90px_minmax(0,1fr)_145px_18px] items-center gap-4 rounded-lg border border-line bg-panel px-4 py-5 text-left text-ink no-underline transition hover:translate-x-0.5 hover:border-line-strong hover:bg-panel-soft max-lg:grid-cols-[80px_minmax(0,1fr)_115px_15px] max-lg:gap-3 max-sm:grid-cols-[1fr_15px] max-sm:gap-3 max-sm:px-3.5 max-sm:py-4"
      to={`/curriculum/${module.id}`}
    >
      <div
        className={`flex items-center gap-2 text-2xs uppercase tracking-wide max-sm:col-start-1 max-sm:order-0 ${moduleStatusStyles[module.status]}`}
      >
        <StatusDot state={module.status} />
        <span>{moduleStatusLabels[module.status]}</span>
      </div>
      <div className="min-w-0 max-sm:col-start-1 max-sm:order-1">
        <div className="flex items-center gap-2 max-sm:items-start max-sm:flex-col max-sm:gap-1">
          <h3 className="m-0 text-base font-semibold tracking-tight">{module.title}</h3>
          {module.prerequisites?.length ? (
            <Badge tone="neutral">Needs {module.prerequisites[0]}</Badge>
          ) : null}
        </div>
        <p className="my-2 text-xs leading-relaxed text-muted max-lg:max-w-sm">
          {module.description}
        </p>
        <div className="flex gap-2 text-2xs text-faint">
          <span>{module.objectiveIds.length} objectives</span>
          <span>·</span>
          <span>
            {total || 'Coming soon'} {total === 1 ? 'item' : 'items'}
          </span>
        </div>
      </div>
      <div className="min-w-0 max-sm:col-start-1 max-sm:order-2 max-sm:max-w-64">
        <div className="mb-2 flex justify-between gap-2">
          <span className="text-xs font-bold text-ink">
            {percent ? `${percent}%` : module.status === 'locked' ? '—' : 'Ready'}
          </span>
          <span className="text-2xs text-faint">
            {module.status === 'locked' ? 'Unlocks later' : 'module progress'}
          </span>
        </div>
        <ProgressBar value={percent} tone={tone} />
      </div>
      <span className="text-right text-base text-faint max-sm:col-start-2 max-sm:row-span-3 max-sm:row-start-1">
        →
      </span>
    </Link>
  )
}
