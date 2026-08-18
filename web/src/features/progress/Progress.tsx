import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Card, PageIntro, SectionHeading, StatusDot } from '../../components/ui'
import { objectives } from '../../prototype/mockData'
import type { MasteryLevel, Objective } from '../../prototype/types'
import { useTutor } from '../tutor/TutorContext'

type Filter = 'all' | 'attention' | 'apply' | 'strong'

export function Progress() {
  const navigate = useNavigate()
  const { setPageContext } = useTutor()
  const [filter, setFilter] = useState<Filter>('all')
  const [selectedId, setSelectedId] = useState(objectives[4].id)
  const selected = objectives.find((item) => item.id === selectedId) ?? objectives[0]

  useEffect(() => {
    setPageContext({
      type: 'progress',
      title: 'Progress',
      objectiveId: selectedId,
      objectiveTitle: selected.title,
    })
  }, [selected.title, selectedId, setPageContext])

  const visible = useMemo(
    () =>
      objectives.filter((objective) => {
        if (filter === 'attention')
          return objective.application !== 'strong' || objective.recall !== 'strong'
        if (filter === 'apply') return objective.application !== 'strong'
        if (filter === 'strong')
          return objective.application === 'strong' && objective.recall === 'strong'
        return true
      }),
    [filter],
  )

  return (
    <div className="page-stack progress-page">
      <PageIntro compact title="Progress" />
      <section className="progress-summary-grid">
        <SummaryStat value="42" label="Introduced" note="concepts encountered" tone="teal" />
        <SummaryStat
          value="31"
          label="Recall strong"
          note="retrieval feels available"
          tone="gold"
        />
        <SummaryStat value="18" label="Applied" note="used in an exercise" tone="coral" />
        <SummaryStat
          value="7"
          label="Transfer tested"
          note="shown in a new context"
          tone="violet"
        />
      </section>
      <div className="progress-workspace">
        <Card className="objectives-card">
          <SectionHeading title="Objectives" />
          <div className="filter-row">
            {(
              [
                ['all', 'All'],
                ['attention', 'Needs attention'],
                ['apply', 'Ready to apply'],
                ['strong', 'Strong'],
              ] as const
            ).map(([value, label]) => (
              <button
                key={value}
                type="button"
                className={filter === value ? 'filter-chip active' : 'filter-chip'}
                onClick={() => setFilter(value)}
              >
                {label}
              </button>
            ))}
          </div>
          <div className="objective-browser">
            {visible.map((objective) => (
              <ObjectiveBrowserRow
                key={objective.id}
                objective={objective}
                selected={selected.id === objective.id}
                onClick={() => setSelectedId(objective.id)}
              />
            ))}
            {!visible.length ? (
              <p className="small-muted">No objectives match this filter yet.</p>
            ) : null}
          </div>
        </Card>
        <Card className="objective-detail-card">
          <div className="detail-topline">
            <Badge tone="coral">Objective</Badge>
            <button className="icon-button" aria-label="More objective actions">
              ···
            </button>
          </div>
          <h2>{selected.title}</h2>
          <p className="body-muted">{selected.description}</p>
          <div className="dimension-list">
            <Dimension
              label="Introduced"
              value={selected.introduced ? 'Yes' : 'Not yet'}
              state={selected.introduced ? 'strong' : 'not-assessed'}
            />
            <Dimension label="Recall" value={labelFor(selected.recall)} state={selected.recall} />
            <Dimension
              label="Conceptual"
              value={labelFor(selected.conceptual)}
              state={selected.conceptual}
            />
            <Dimension
              label="Application"
              value={labelFor(selected.application)}
              state={selected.application}
            />
            <Dimension
              label="Transfer"
              value={labelFor(selected.transfer)}
              state={selected.transfer}
            />
          </div>
          <div className="detail-next">
            <p className="eyebrow">Next</p>
            <h3>
              {selected.application === 'strong'
                ? 'Try a new context'
                : 'Do one small implementation'}
            </h3>
            <button
              className="text-link"
              onClick={() =>
                navigate(
                  selected.id.includes('backprop')
                    ? '/exercise/gradient-descent-exercise'
                    : '/lesson/backpropagation',
                )
              }
            >
              Choose practice →
            </button>
          </div>
        </Card>
      </div>
    </div>
  )
}

function SummaryStat({
  value,
  label,
  note,
  tone,
}: {
  value: string
  label: string
  note: string
  tone: 'teal' | 'gold' | 'coral' | 'violet'
}) {
  return (
    <Card className={`summary-stat summary-${tone}`}>
      <span className="summary-number">{value}</span>
      <strong>{label}</strong>
      <span>{note}</span>
    </Card>
  )
}

function ObjectiveBrowserRow({
  objective,
  selected,
  onClick,
}: {
  objective: Objective
  selected: boolean
  onClick: () => void
}) {
  const level =
    objective.application === 'strong' && objective.recall === 'strong'
      ? 'strong'
      : objective.recall === 'developing'
        ? 'attention'
        : 'developing'
  return (
    <button
      type="button"
      onClick={onClick}
      className={`objective-browser-row ${selected ? 'selected' : ''}`}
    >
      <StatusDot
        state={
          level === 'strong' ? 'completed' : level === 'attention' ? 'available' : 'in-progress'
        }
      />
      <span>
        <strong>{objective.title}</strong>
        <small>{objective.moduleId.replaceAll('-', ' ')}</small>
      </span>
      <Badge tone={level === 'strong' ? 'teal' : level === 'attention' ? 'coral' : 'gold'}>
        {level === 'strong' ? 'Strong' : level === 'attention' ? 'Needs attention' : 'Developing'}
      </Badge>
      <span className="row-arrow">→</span>
    </button>
  )
}

function Dimension({ label, value, state }: { label: string; value: string; state: MasteryLevel }) {
  return (
    <div className="dimension-row">
      <span>{label}</span>
      <span className={`dimension-line ${state}`} />
      <strong>{value}</strong>
    </div>
  )
}

function labelFor(level: MasteryLevel) {
  return level === 'strong' ? 'Strong' : level === 'developing' ? 'Developing' : 'Not assessed'
}
