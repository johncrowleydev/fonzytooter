import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Card, PageIntro, SectionHeading, StatusDot } from '../../components/ui'
import { objectives } from '../../prototype/mockData'
import type { MasteryLevel, Objective } from '../../prototype/types'
import { useTutor } from '../tutor/TutorContext'

type Filter = 'all' | 'attention' | 'apply' | 'strong'

const filterButtonStyles = {
  active: 'border-brand-teal/40 bg-brand-teal/10 text-ink',
  inactive:
    'border-line bg-transparent text-faint hover:border-brand-teal/40 hover:bg-brand-teal/10 hover:text-ink',
} as const

const summaryToneStyles = {
  teal: 'border-t-brand-teal',
  gold: 'border-t-brand-gold',
  coral: 'border-t-brand-coral',
  violet: 'border-t-brand-violet',
} as const

const dimensionToneStyles = {
  strong: 'bg-brand-teal',
  developing: 'bg-brand-gold',
  'not-assessed': 'bg-brand-slate/20',
} as const

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
    <div className="grid max-w-6xl gap-7 max-sm:gap-5">
      <PageIntro compact title="Progress" />
      <section className="grid grid-cols-4 gap-2.5 max-lg:grid-cols-2 max-sm:grid-cols-2">
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
      <div className="grid grid-cols-[1.15fr_0.85fr] gap-3.5 max-lg:grid-cols-1 max-sm:gap-2.5">
        <Card className="min-h-128">
          <SectionHeading title="Objectives" />
          <div className="flex flex-wrap gap-1.5 border-b border-line pb-3.5">
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
                className={`rounded-full border px-2.5 py-2 text-2xs ${filter === value ? filterButtonStyles.active : filterButtonStyles.inactive}`}
                onClick={() => setFilter(value)}
              >
                {label}
              </button>
            ))}
          </div>
          <div className="grid">
            {visible.map((objective) => (
              <ObjectiveBrowserRow
                key={objective.id}
                objective={objective}
                selected={selected.id === objective.id}
                onClick={() => setSelectedId(objective.id)}
              />
            ))}
            {!visible.length ? (
              <p className="text-2xs text-faint">No objectives match this filter yet.</p>
            ) : null}
          </div>
        </Card>
        <Card className="min-h-128 p-6 max-lg:min-h-0 max-sm:p-5">
          <div className="flex items-center justify-between gap-3">
            <Badge tone="coral">Objective</Badge>
            <button
              className="border-0 bg-transparent p-1 text-xl text-faint hover:text-ink"
              aria-label="More objective actions"
            >
              ···
            </button>
          </div>
          <h2 className="my-5 max-w-sm text-3xl leading-tight tracking-tight">{selected.title}</h2>
          <p className="text-xs leading-relaxed text-muted">{selected.description}</p>
          <div className="mt-8 grid gap-0">
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
          <div className="mt-9 rounded-lg border border-brand-coral/30 bg-brand-coral/10 p-4">
            <p className="text-2xs font-bold uppercase tracking-widest text-faint">Next</p>
            <h3 className="my-2 text-sm">
              {selected.application === 'strong'
                ? 'Try a new context'
                : 'Do one small implementation'}
            </h3>
            <button
              className="border-0 bg-transparent p-0 text-xs font-bold text-brand-teal hover:text-ink"
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
    <Card className={`grid gap-1 border-t-2 border-t-transparent p-5 ${summaryToneStyles[tone]}`}>
      <span className="text-4xl tracking-tight">{value}</span>
      <strong className="text-xs">{label}</strong>
      <span className="text-2xs text-faint">{note}</span>
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
      className={`grid w-full grid-cols-[11px_minmax(0,1fr)_auto_15px] items-center gap-2.5 border-0 border-b border-line bg-transparent py-3 text-left text-ink max-sm:flex max-sm:min-w-0 ${selected ? 'border-l-2 border-brand-teal bg-brand-teal/5 pl-2' : 'hover:bg-brand-teal/5'}`}
    >
      <StatusDot
        state={
          level === 'strong' ? 'completed' : level === 'attention' ? 'available' : 'in-progress'
        }
      />
      <span className="min-w-0 max-sm:min-w-0 max-sm:flex-1">
        <strong className="block overflow-wrap-anywhere text-xs font-semibold">
          {objective.title}
        </strong>
        <small className="mt-1 block text-2xs capitalize text-faint">
          {objective.moduleId.replaceAll('-', ' ')}
        </small>
      </span>
      <Badge tone={level === 'strong' ? 'teal' : level === 'attention' ? 'coral' : 'gold'}>
        {level === 'strong' ? 'Strong' : level === 'attention' ? 'Needs attention' : 'Developing'}
      </Badge>
      <span className="text-right text-base text-faint max-sm:ml-auto max-sm:w-4">→</span>
    </button>
  )
}

function Dimension({ label, value, state }: { label: string; value: string; state: MasteryLevel }) {
  return (
    <div className="grid grid-cols-[85px_1fr_100px] items-center gap-3 border-b border-line py-3 max-sm:grid-cols-[72px_1fr_84px]">
      <span className="text-2xs text-muted">{label}</span>
      <span className={`h-1 rounded-full ${dimensionToneStyles[state]}`} />
      <strong className="text-right text-2xs font-semibold">{value}</strong>
    </div>
  )
}

function labelFor(level: MasteryLevel) {
  return level === 'strong' ? 'Strong' : level === 'developing' ? 'Developing' : 'Not assessed'
}
