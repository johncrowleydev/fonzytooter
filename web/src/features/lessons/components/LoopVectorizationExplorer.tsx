import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader, ValueGrid } from './ArrayVisual'
import { HighlightedCode } from '../../../components/HighlightedCode'

const inputs = [0, 10, 20, 30]

export function LoopVectorizationExplorer() {
  const [visited, setVisited] = useState(0)
  const outputs = inputs.map((value) => value * 1.8 + 32)
  const selected = Array.from({ length: visited }, (_, index) => index)

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Loop mechanics and array-level intent"
        description="Both forms compute the same transformation. Step through the Python loop, then compare the level of description."
        badge={<Badge tone="blue">same values</Badge>}
      />
      <div className="grid gap-4 lg:grid-cols-2">
        <CodePanel
          title="Explicit Python loop"
          code={'result = []\nfor c in celsius:\n    result.append(c * 9 / 5 + 32)'}
          detail="Your code manages iteration and result construction."
        />
        <CodePanel
          title="NumPy array expression"
          code={'result = celsius * 9 / 5 + 32'}
          detail="Your code states one transformation at the array's structural level."
        />
      </div>
      <ValueGrid
        values={outputs}
        columns={4}
        selected={selected}
        label={`${visited} of 4 Fahrenheit outputs visited by the illustrated Python loop`}
      />
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() => setVisited((current) => Math.min(inputs.length, current + 1))}
          disabled={visited === inputs.length}
        >
          Run one Python iteration
        </Button>
        <Button variant="secondary" onClick={() => setVisited(inputs.length)}>
          Delegate whole array to NumPy
        </Button>
        <Button variant="outline" onClick={() => setVisited(0)}>
          Reset
        </Button>
      </div>
      <p
        className="rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-4 text-sm leading-relaxed text-ink"
        role="status"
        aria-live="polite"
      >
        {visited === inputs.length
          ? 'All outputs are computed. Vectorization does not make repetition disappear at the machine level; it delegates repeated numerical work to NumPy’s compiled machinery.'
          : `The explicit loop has visited ${visited} of ${inputs.length} values.`}
      </p>
    </Card>
  )
}

function CodePanel({ title, code, detail }: { title: string; code: string; detail: string }) {
  return (
    <section className="rounded-lg border border-line bg-raised p-4">
      <h3 className="text-sm font-semibold text-ink">{title}</h3>
      <pre className="mt-3 overflow-x-auto overscroll-x-contain rounded-md bg-code-surface p-3 font-mono text-sm leading-relaxed text-code-ink">
        <code>
          <HighlightedCode code={code} language="python" />
        </code>
      </pre>
      <p className="mt-3 text-sm leading-relaxed text-muted">{detail}</p>
    </section>
  )
}
