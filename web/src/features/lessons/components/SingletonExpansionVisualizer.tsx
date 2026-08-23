import { Badge, Card } from '../../../components/ui'
import { InteractiveHeader, ValueGrid } from './ArrayVisual'

const columnExpanded = [10, 10, 10, 10, 20, 20, 20, 20, 30, 30, 30, 30]
const rowExpanded = [1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4]
const result = columnExpanded.map((value, index) => value + rowExpanded[index])

export function SingletonExpansionVisualizer() {
  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Singleton axes expand in different directions"
        description="Track how (3, 1) and (1, 4) can both behave like (3, 4) for one element-wise addition."
        badge={<Badge tone="violet">reasoning model</Badge>}
      />
      <div className="grid gap-4 lg:grid-cols-3">
        <ExpansionPanel
          title="column (3, 1) → (3, 4)"
          direction="conceptually repeats horizontally"
          values={columnExpanded}
        />
        <ExpansionPanel
          title="row (1, 4) → (3, 4)"
          direction="conceptually repeats vertically"
          values={rowExpanded}
        />
        <ExpansionPanel
          title="column + row → (3, 4)"
          direction="element-wise result"
          values={result}
        />
      </div>
      <p className="rounded-lg border border-accent-gold/30 bg-accent-gold/10 p-4 text-sm leading-relaxed text-ink">
        <strong>Conceptual, not physical:</strong> repeated values make the result easy to reason
        about. NumPy normally does not need to materialize these expanded operand arrays in memory.
      </p>
    </Card>
  )
}

function ExpansionPanel({
  title,
  direction,
  values,
}: {
  title: string
  direction: string
  values: readonly number[]
}) {
  return (
    <section className="grid gap-3 rounded-lg border border-line bg-raised p-4">
      <h3 className="font-mono text-sm text-ink">{title}</h3>
      <ValueGrid values={values} columns={4} label={`${title}; ${values.join(', ')}`} />
      <p className="text-xs font-bold uppercase tracking-widest text-faint">{direction}</p>
    </section>
  )
}
