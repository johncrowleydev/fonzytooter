import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Card, PageIntro, ProgressBar, SectionHeading, StatusDot } from '../../components/ui'
import { modules } from '../../prototype/mockData'
import { useTutor } from '../tutor/TutorContext'

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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <PageIntro compact title="Curriculum" />
      <div className="flex items-center justify-between gap-[15px] max-[640px]:items-stretch max-[640px]:flex-col">
        <div className="flex w-[min(350px,100%)] items-center gap-[9px] rounded-lg border border-[var(--line)] bg-[var(--panel-soft)] px-[11px] py-[9px] text-[var(--faint)] max-[640px]:w-full">
          <span>⌕</span>
          <input
            className="w-full border-0 bg-transparent text-[11px] text-[var(--ink)] outline-0 placeholder:text-[var(--faint)]"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search modules or concepts"
          />
        </div>
        <div className="flex gap-[3px] rounded-[7px] border border-[var(--line)] p-[3px] max-[640px]:self-start">
          <button className="rounded-[5px] bg-[rgba(157,185,194,0.14)] px-2.5 py-1.5 text-[10px] text-[var(--ink)]">
            List
          </button>
          <button className="rounded-[5px] border-0 bg-transparent px-2.5 py-1.5 text-[10px] text-[var(--faint)]">
            Compact
          </button>
        </div>
      </div>
      <div className="grid grid-cols-[minmax(0,1fr)_270px] gap-10 max-[1120px]:grid-cols-1">
        <div className="grid gap-8 max-[640px]:gap-[21px]">
          {groups.map((group) => {
            const groupModules = filteredModules.filter((module) => module.eyebrow === group)
            if (!groupModules.length) return null
            return (
              <section key={group} className="grid">
                <SectionHeading title={group} detail={`${groupModules.length} modules`} />
                {groupModules.map((module) => (
                  <ModuleRow key={module.id} module={module} />
                ))}
              </section>
            )
          })}
        </div>
        <Card className="sticky top-[30px] mt-[35px] self-start max-[1120px]:hidden">
          <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
            Status
          </p>
          <div className="mt-[25px] grid gap-3">
            <div className="flex items-center gap-2.5 text-[10px] text-[var(--muted)]">
              <StatusDot state="completed" />
              <span>Complete</span>
            </div>
            <div className="flex items-center gap-2.5 text-[10px] text-[var(--muted)]">
              <StatusDot state="in-progress" />
              <span>In progress</span>
            </div>
            <div className="flex items-center gap-2.5 text-[10px] text-[var(--muted)]">
              <StatusDot state="available" />
              <span>Available</span>
            </div>
            <div className="flex items-center gap-2.5 text-[10px] text-[var(--muted)]">
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
  const navigate = useNavigate()
  const complete = module.lessons.filter((lesson) => lesson.completed).length
  const total = module.lessons.length
  const percent = total
    ? Math.round((complete / total) * 100)
    : module.status === 'completed'
      ? 100
      : 0
  const tone = module.accent as 'teal' | 'gold' | 'coral' | 'violet' | 'blue'
  return (
    <button
      type="button"
      className={`grid w-full grid-cols-[90px_minmax(0,1fr)_145px_18px] items-center gap-4 rounded-[10px] border border-[var(--line)] bg-[var(--panel)] px-4 py-[17px] text-left text-[var(--ink)] transition hover:translate-x-0.5 hover:border-[var(--line-strong)] hover:bg-[var(--panel-soft)] max-[860px]:grid-cols-[80px_minmax(0,1fr)_115px_15px] max-[860px]:gap-[11px] max-[640px]:grid-cols-[1fr_15px] max-[640px]:gap-3 max-[640px]:px-3.5 max-[640px]:py-4`}
      onClick={() => navigate(`/curriculum/${module.id}`)}
    >
      <div
        className={`flex items-center gap-2 text-[9px] uppercase tracking-[0.06em] max-[640px]:col-start-1 max-[640px]:order-0 ${module.status === 'in-progress' ? 'text-[var(--gold)]' : module.status === 'completed' ? 'text-[var(--teal)]' : 'text-[var(--faint)]'}`}
      >
        <StatusDot state={module.status} />
        <span>
          {module.status === 'in-progress'
            ? 'Current'
            : module.status === 'completed'
              ? 'Complete'
              : module.status === 'available'
                ? 'Available'
                : 'Locked'}
        </span>
      </div>
      <div className="min-w-0 max-[640px]:col-start-1 max-[640px]:order-1">
        <div className="flex items-center gap-2 max-[640px]:items-start max-[640px]:flex-col max-[640px]:gap-[5px]">
          <h3 className="m-0 text-[15px] font-semibold tracking-[-0.025em]">{module.title}</h3>
          {module.prerequisites?.length ? (
            <Badge tone="neutral">Needs {module.prerequisites[0]}</Badge>
          ) : null}
        </div>
        <p className="my-[6px_8px] text-[11px] leading-[1.45] text-[var(--muted)] max-[860px]:max-w-[400px]">
          {module.description}
        </p>
        <div className="flex gap-[7px] text-[9px] text-[var(--faint)]">
          <span>{module.objectiveIds.length} objectives</span>
          <span>·</span>
          <span>
            {total || 'Coming soon'} {total === 1 ? 'item' : 'items'}
          </span>
        </div>
      </div>
      <div className="min-w-0 max-[640px]:col-start-1 max-[640px]:order-2 max-[640px]:max-w-[260px]">
        <div className="mb-[7px] flex justify-between gap-2">
          <span className="text-[11px] font-bold text-[var(--ink)]">
            {percent ? `${percent}%` : module.status === 'locked' ? '—' : 'Ready'}
          </span>
          <span className="text-[9px] text-[var(--faint)]">
            {module.status === 'locked' ? 'Unlocks later' : 'module progress'}
          </span>
        </div>
        <ProgressBar value={percent} tone={tone} />
      </div>
      <span className="text-right text-base text-[var(--faint)] max-[640px]:col-start-2 max-[640px]:row-span-3 max-[640px]:row-start-1">
        →
      </span>
    </button>
  )
}
