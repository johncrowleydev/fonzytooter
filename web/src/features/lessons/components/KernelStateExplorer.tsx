import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader } from './ArrayVisual'

export type KernelMemory = { rate?: number; result?: number }

export function executeNotebookCell(
  memory: KernelMemory,
  cell: number,
  visibleRate: number,
): KernelMemory {
  if (cell === 0) return { ...memory, rate: visibleRate }
  if (cell === 1) {
    return memory.rate === undefined ? memory : { ...memory, result: memory.rate * 10 }
  }
  return memory
}

function cellThreeOutput(memory: KernelMemory) {
  return memory.result === undefined
    ? "NameError: name 'result' is not defined"
    : String(memory.result)
}

const emptyKernel: KernelMemory = {}

export function KernelStateExplorer() {
  const [rateSource, setRateSource] = useState('2')
  const [kernel, setKernel] = useState<KernelMemory>(emptyKernel)
  const [cell3Output, setCell3Output] = useState('')
  const [counts, setCounts] = useState<(number | null)[]>([null, null, null])
  const [nextCount, setNextCount] = useState(1)
  const [history, setHistory] = useState<string[]>([])
  const visibleRate = Number(rateSource) || 0

  const sources = [`rate = ${rateSource}`, 'result = rate * 10', 'print(result)']

  function runCell(cell: number) {
    const nextKernel = executeNotebookCell(kernel, cell, visibleRate)
    setKernel(nextKernel)
    if (cell === 2) setCell3Output(cellThreeOutput(nextKernel))
    setCounts((current) => current.map((count, index) => (index === cell ? nextCount : count)))
    setHistory((current) => [...current, `[${nextCount}] Cell ${cell + 1}: ${sources[cell]}`])
    setNextCount((current) => current + 1)
  }

  function restartKernel() {
    setKernel(emptyKernel)
    setCounts([null, null, null])
    setNextCount(1)
    setHistory([])
  }

  function runAll() {
    let memory = emptyKernel
    memory = executeNotebookCell(memory, 0, visibleRate)
    memory = executeNotebookCell(memory, 1, visibleRate)
    memory = executeNotebookCell(memory, 2, visibleRate)
    setKernel(memory)
    setCell3Output(cellThreeOutput(memory))
    setCounts([1, 2, 3])
    setNextCount(4)
    setHistory(sources.map((source, index) => `[${index + 1}] Cell ${index + 1}: ${source}`))
  }

  function resetSimulation() {
    setRateSource('2')
    restartKernel()
    setCell3Output('')
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="The page has an order; the kernel has a history"
        description="Run cells out of order, edit visible source without executing it, and inspect what the simulated kernel actually remembers."
        badge={<Badge tone="gold">deterministic simulation</Badge>}
      />
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(15rem,0.45fr)]">
        <div className="grid gap-3">
          <NotebookCell count={counts[0]} number={1} onRun={() => runCell(0)}>
            <label className="grid gap-2 text-sm text-muted" htmlFor="kernel-rate-source">
              <span>Visible source (editing does not execute)</span>
              <span className="flex items-center gap-2 font-mono text-sm text-ink">
                rate ={' '}
                <input
                  id="kernel-rate-source"
                  aria-label="Visible rate source"
                  className="w-20 rounded-md border border-line-strong bg-panel px-2 py-1 text-ink"
                  type="number"
                  value={rateSource}
                  onChange={(event) => setRateSource(event.target.value)}
                />
              </span>
            </label>
          </NotebookCell>
          <NotebookCell count={counts[1]} number={2} onRun={() => runCell(1)}>
            <code>result = rate * 10</code>
          </NotebookCell>
          <NotebookCell count={counts[2]} number={3} onRun={() => runCell(2)}>
            <code>print(result)</code>
            {cell3Output ? (
              <pre
                className="mt-3 rounded-md border border-line bg-panel-soft p-3 text-sm text-ink"
                aria-label="Cell 3 output"
              >
                {cell3Output}
              </pre>
            ) : null}
          </NotebookCell>
        </div>
        <aside className="grid content-start gap-3 rounded-lg border border-accent-violet/30 bg-accent-violet/10 p-4">
          <h3 className="text-sm font-semibold text-ink">Remembered kernel state</h3>
          <p className="font-mono text-sm text-muted">rate = {kernel.rate ?? 'undefined'}</p>
          <p className="font-mono text-sm text-muted">result = {kernel.result ?? 'undefined'}</p>
          <h3 className="mt-2 text-sm font-semibold text-ink">Execution history</h3>
          {history.length ? (
            <ol className="grid gap-1 text-sm text-muted">
              {history.map((entry, index) => (
                <li key={`${entry}-${index}`}>{entry}</li>
              ))}
            </ol>
          ) : (
            <p className="text-sm text-muted">No cells executed.</p>
          )}
        </aside>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button onClick={runAll}>Restart and run all top to bottom</Button>
        <Button variant="secondary" onClick={restartKernel}>
          Restart kernel
        </Button>
        <Button variant="outline" onClick={resetSimulation}>
          Reset simulation
        </Button>
      </div>
      <p className="rounded-lg border border-accent-teal/30 bg-accent-teal/10 p-4 text-sm leading-relaxed text-ink">
        Try this: run Cells 1 and 2, edit Cell 1 without running it, then run Cell 3. The visible
        page suggests one result while the kernel remembers another. Restart-and-run-all makes
        document order and execution history agree.
      </p>
    </Card>
  )
}

function NotebookCell({
  count,
  number,
  onRun,
  children,
}: {
  count: number | null
  number: number
  onRun: () => void
  children: React.ReactNode
}) {
  return (
    <section className="grid gap-3 rounded-lg border border-line bg-raised p-4">
      <div className="flex items-center justify-between gap-3">
        <span className="font-mono text-sm text-faint">
          In [{count ?? ' '}]: Cell {number}
        </span>
        <Button variant="outline" onClick={onRun}>
          Run cell {number}
        </Button>
      </div>
      <div className="font-mono text-sm text-ink">{children}</div>
    </section>
  )
}
