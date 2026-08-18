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
    <div className="grid max-w-[1140px] gap-[29px] max-[640px]:gap-[21px]">
      <PageIntro compact title="Progress" />
      <section className="grid grid-cols-4 gap-2.5 max-[860px]:grid-cols-2 max-[640px]:grid-cols-2">
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
      <div className="grid grid-cols-[1.15fr_0.85fr] gap-3.5 max-[860px]:grid-cols-1 max-[640px]:gap-2.5">
        <Card className="min-h-[550px]">
          <SectionHeading title="Objectives" />
          <div className="flex flex-wrap gap-1.5 border-b border-[var(--line)] pb-3.5">
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
                className={`rounded-full border px-2.5 py-[7px] text-[10px] ${filter === value ? 'border-[rgba(118,208,192,0.35)] bg-[rgba(118,208,192,0.08)] text-[var(--ink)]' : 'border-[var(--line)] bg-transparent text-[var(--faint)] hover:border-[rgba(118,208,192,0.35)] hover:bg-[rgba(118,208,192,0.08)] hover:text-[var(--ink)]'}`}
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
              <p className="text-[10px] text-[var(--faint)]">
                No objectives match this filter yet.
              </p>
            ) : null}
          </div>
        </Card>
        <Card className="min-h-[550px] p-[25px] max-[860px]:min-h-0 max-[640px]:p-[18px]">
          <div className="flex items-center justify-between gap-3">
            <Badge tone="coral">Objective</Badge>
            <button
              className="border-0 bg-transparent p-[3px] text-[19px] text-[var(--faint)] hover:text-[var(--ink)]"
              aria-label="More objective actions"
            >
              ···
            </button>
          </div>
          <h2 className="my-[21px_10px] max-w-[390px] text-[26px] leading-[1.15] tracking-[-0.045em]">
            {selected.title}
          </h2>
          <p className="text-xs leading-[1.65] text-[var(--muted)]">{selected.description}</p>
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
          <div className="mt-[35px] rounded-[9px] border border-[rgba(239,145,110,0.24)] bg-[rgba(239,145,110,0.055)] p-4">
            <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--faint)]">
              Next
            </p>
            <h3 className="my-[7px] text-sm">
              {selected.application === 'strong'
                ? 'Try a new context'
                : 'Do one small implementation'}
            </h3>
            <button
              className="border-0 bg-transparent p-0 text-[11px] font-bold text-[var(--teal)] hover:text-[var(--ink)]"
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
    <Card
      className={`grid gap-[5px] border-t-2 border-t-transparent p-[19px] ${tone === 'teal' ? 'border-t-[var(--teal)]' : tone === 'gold' ? 'border-t-[var(--gold)]' : tone === 'coral' ? 'border-t-[var(--coral)]' : 'border-t-[var(--violet)]'}`}
    >
      <span className="text-[30px] tracking-[-0.055em]">{value}</span>
      <strong className="text-[11px]">{label}</strong>
      <span className="text-[9px] text-[var(--faint)]">{note}</span>
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
      className={`grid w-full grid-cols-[11px_minmax(0,1fr)_auto_15px] items-center gap-2.5 border-0 border-b border-[var(--line)] bg-transparent py-[13px] text-left text-[var(--ink)] max-[640px]:flex max-[640px]:min-w-0 ${selected ? 'bg-[rgba(118,208,192,0.04)] pl-[9px] shadow-[inset_2px_0_0_var(--teal)]' : 'hover:bg-[rgba(118,208,192,0.04)]'}`}
    >
      <StatusDot
        state={
          level === 'strong' ? 'completed' : level === 'attention' ? 'available' : 'in-progress'
        }
      />
      <span className="min-w-0 max-[640px]:min-w-0 max-[640px]:flex-1">
        <strong className="block overflow-wrap-anywhere text-[11px] font-semibold">
          {objective.title}
        </strong>
        <small className="mt-1 block text-[9px] capitalize text-[var(--faint)]">
          {objective.moduleId.replaceAll('-', ' ')}
        </small>
      </span>
      <Badge tone={level === 'strong' ? 'teal' : level === 'attention' ? 'coral' : 'gold'}>
        {level === 'strong' ? 'Strong' : level === 'attention' ? 'Needs attention' : 'Developing'}
      </Badge>
      <span className="text-right text-base text-[var(--faint)] max-[640px]:ml-auto max-[640px]:w-[15px]">
        →
      </span>
    </button>
  )
}

function Dimension({ label, value, state }: { label: string; value: string; state: MasteryLevel }) {
  return (
    <div className="grid grid-cols-[85px_1fr_100px] items-center gap-3 border-b border-[var(--line)] py-3 max-[640px]:grid-cols-[72px_1fr_84px]">
      <span className="text-[10px] text-[var(--muted)]">{label}</span>
      <span
        className={`h-[3px] rounded-full ${state === 'strong' ? 'bg-[var(--teal)]' : state === 'developing' ? 'bg-[var(--gold)]' : 'bg-[rgba(157,185,194,0.16)]'}`}
      />
      <strong className="text-right text-[10px] font-semibold">{value}</strong>
    </div>
  )
}

function labelFor(level: MasteryLevel) {
  return level === 'strong' ? 'Strong' : level === 'developing' ? 'Developing' : 'Not assessed'
}
