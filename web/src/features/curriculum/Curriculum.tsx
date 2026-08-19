import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useListModules } from '../../api/generated/endpoints'
import type { ModuleSummary } from '../../api/generated/schemas/moduleSummary.zod'
import { Card, PageIntro, SectionHeading } from '../../components/ui'
import { useTutor } from '../tutor/TutorContext'

export function Curriculum() {
  const { setPageContext } = useTutor()
  const [query, setQuery] = useState('')
  const modulesQuery = useListModules()

  useEffect(() => {
    setPageContext({ type: 'curriculum', title: 'Curriculum' })
  }, [setPageContext])

  if (modulesQuery.isPending) {
    return <CurriculumState title="Loading curriculum" detail="Fetching the available modules…" />
  }

  if (modulesQuery.isError) {
    return (
      <CurriculumState
        title="Curriculum unavailable"
        detail={getErrorMessage(modulesQuery.error)}
        action={<RetryButton onClick={() => void modulesQuery.refetch()} />}
      />
    )
  }

  const modules = modulesQuery.data.data

  if (modules.length === 0) {
    return (
      <CurriculumState
        title="No modules yet"
        detail="Curriculum modules will appear here when they are published."
      />
    )
  }

  return <CurriculumContent modules={modules} query={query} onQueryChange={setQuery} />
}

function CurriculumContent({
  modules,
  query,
  onQueryChange,
}: {
  modules: ModuleSummary[]
  query: string
  onQueryChange: (value: string) => void
}) {
  const filteredModules = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase()
    if (!normalizedQuery) return modules

    return modules.filter((module) =>
      `${module.title} ${module.id}`.toLowerCase().includes(normalizedQuery),
    )
  }, [modules, query])

  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Curriculum" />
      <div className="flex items-center justify-between gap-4 max-sm:items-stretch max-sm:flex-col">
        <label className="flex w-full max-w-sm items-center gap-2 rounded-lg border border-line bg-panel-soft px-3 py-2 text-faint max-sm:max-w-none">
          <span aria-hidden="true">⌕</span>
          <span className="sr-only">Search curriculum</span>
          <input
            className="w-full border-0 bg-transparent text-xs text-ink outline-0 placeholder:text-faint"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search modules"
          />
        </label>
        <span className="text-xs text-muted">
          {modules.length} {modules.length === 1 ? 'module' : 'modules'}
        </span>
      </div>

      <section>
        <SectionHeading title="Modules" />
        {filteredModules.length > 0 ? (
          <div className="grid gap-2.5">
            {filteredModules.map((module) => (
              <ModuleRow key={module.id} module={module} />
            ))}
          </div>
        ) : (
          <Card muted>
            <p className="text-xs leading-relaxed text-muted">No modules match “{query}”.</p>
          </Card>
        )}
      </section>
    </div>
  )
}

function ModuleRow({ module }: { module: ModuleSummary }) {
  return (
    <Link
      className="flex items-center gap-4 rounded-lg border border-line bg-panel px-4 py-5 text-ink no-underline transition hover:translate-x-0.5 hover:border-line-strong hover:bg-panel-soft max-sm:gap-3 max-sm:px-3.5 max-sm:py-4"
      to={`/curriculum/${module.id}`}
    >
      <span className="w-16 shrink-0 text-2xs font-bold uppercase tracking-wide text-faint">
        Module {String(module.order + 1).padStart(2, '0')}
      </span>
      <span className="min-w-0 flex-1">
        <strong className="block text-base tracking-tight">{module.title}</strong>
        <span className="mt-1 block text-2xs text-faint">{module.id}</span>
      </span>
      <span className="text-base text-faint" aria-hidden="true">
        →
      </span>
    </Link>
  )
}

function CurriculumState({
  title,
  detail,
  action,
}: {
  title: string
  detail: string
  action?: ReactNode
}) {
  return (
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Curriculum" />
      <Card muted>
        <h2 className="text-base tracking-tight text-ink">{title}</h2>
        <p className="mt-2 text-xs leading-relaxed text-muted">{detail}</p>
        {action ? <div className="mt-4">{action}</div> : null}
      </Card>
    </div>
  )
}

function RetryButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      className="rounded-lg border border-line-strong bg-brand-slate/10 px-3 py-2 text-xs font-bold text-ink transition hover:bg-brand-slate/20"
      type="button"
      onClick={onClick}
    >
      Try again
    </button>
  )
}

function getErrorMessage(error: unknown) {
  return error instanceof Error && error.message
    ? error.message
    : 'The curriculum could not be loaded. Try again in a moment.'
}
