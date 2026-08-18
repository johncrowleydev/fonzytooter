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
    <div className="page-stack">
      <PageIntro compact title="Curriculum" />
      <div className="toolbar">
        <div className="search-field">
          <span>⌕</span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search modules or concepts"
          />
        </div>
        <div className="view-toggle">
          <button className="active">List</button>
          <button>Compact</button>
        </div>
      </div>
      <div className="curriculum-layout">
        <div className="curriculum-list">
          {groups.map((group) => {
            const groupModules = filteredModules.filter((module) => module.eyebrow === group)
            if (!groupModules.length) return null
            return (
              <section key={group} className="module-group">
                <SectionHeading title={group} detail={`${groupModules.length} modules`} />
                {groupModules.map((module) => (
                  <ModuleRow key={module.id} module={module} />
                ))}
              </section>
            )
          })}
        </div>
        <Card className="curriculum-aside">
          <p className="eyebrow">Status</p>
          <div className="legend">
            <div>
              <StatusDot state="completed" />
              <span>Complete</span>
            </div>
            <div>
              <StatusDot state="in-progress" />
              <span>In progress</span>
            </div>
            <div>
              <StatusDot state="available" />
              <span>Available</span>
            </div>
            <div>
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
      className={`module-row module-${module.status}`}
      onClick={() => navigate(`/curriculum/${module.id}`)}
    >
      <div className="module-status">
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
      <div className="module-main">
        <div className="module-title-line">
          <h3>{module.title}</h3>
          {module.prerequisites?.length ? (
            <Badge tone="neutral">Needs {module.prerequisites[0]}</Badge>
          ) : null}
        </div>
        <p>{module.description}</p>
        <div className="module-meta">
          <span>{module.objectiveIds.length} objectives</span>
          <span>·</span>
          <span>
            {total || 'Coming soon'} {total === 1 ? 'item' : 'items'}
          </span>
        </div>
      </div>
      <div className="module-progress">
        <div className="progress-copy">
          <span>{percent ? `${percent}%` : module.status === 'locked' ? '—' : 'Ready'}</span>
          <span>{module.status === 'locked' ? 'Unlocks later' : 'module progress'}</span>
        </div>
        <ProgressBar value={percent} tone={tone} />
      </div>
      <span className="row-arrow">→</span>
    </button>
  )
}
