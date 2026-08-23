import { useId } from 'react'

export type MappingEdge = {
  from: string
  to: string
}

export type MappingAnalysis = {
  outgoingCounts: Record<string, number>
  incomingCounts: Record<string, number>
  image: string[]
  isFunction: boolean
  isInjective: boolean
  isSurjective: boolean
  isBijective: boolean
}

export function analyzeMapping(
  domain: readonly string[],
  codomain: readonly string[],
  edges: readonly MappingEdge[],
): MappingAnalysis {
  const outgoingCounts = Object.fromEntries(domain.map((value) => [value, 0]))
  const incomingCounts = Object.fromEntries(codomain.map((value) => [value, 0]))

  for (const edge of edges) {
    if (edge.from in outgoingCounts) outgoingCounts[edge.from] += 1
    if (edge.to in incomingCounts) incomingCounts[edge.to] += 1
  }

  const image = codomain.filter((value) => incomingCounts[value] > 0)
  const isFunction = domain.every((value) => outgoingCounts[value] === 1)
  const isInjective = isFunction && codomain.every((value) => incomingCounts[value] <= 1)
  const isSurjective = isFunction && codomain.every((value) => incomingCounts[value] >= 1)

  return {
    outgoingCounts,
    incomingCounts,
    image,
    isFunction,
    isInjective,
    isSurjective,
    isBijective: isInjective && isSurjective,
  }
}

export function setFunctionOutput(edges: readonly MappingEdge[], input: string, output: string) {
  return [...edges.filter((edge) => edge.from !== input), { from: input, to: output }]
}

type MappingDiagramProps = {
  domain: readonly string[]
  codomain: readonly string[]
  edges: readonly MappingEdge[]
  selectedInput?: string
  onSelectInput?: (input: string) => void
  onToggleEdge?: (input: string, output: string) => void
  compact?: boolean
  singleOutput?: boolean
}

const nodeButtonClass = 'rounded-lg border px-3 py-2 text-sm font-semibold transition'

export function MappingDiagram({
  domain,
  codomain,
  edges,
  selectedInput,
  onSelectInput,
  onToggleEdge,
  compact = false,
  singleOutput = false,
}: MappingDiagramProps) {
  const markerId = useId().replace(/:/g, '')
  const rowCount = Math.max(domain.length, codomain.length)
  const height = 52 + rowCount * 58
  const yFor = (index: number, count: number) => 52 + index * 58 + ((rowCount - count) * 58) / 2
  const editable = Boolean(onSelectInput && onToggleEdge)
  const selectedTargets = new Set(
    edges.filter((edge) => edge.from === selectedInput).map((edge) => edge.to),
  )

  return (
    <div className="grid min-w-0 gap-4">
      <div className="rounded-lg border border-line bg-panel-soft p-3">
        <div className="flex justify-between text-xs font-bold uppercase tracking-widest text-faint">
          <span>Domain</span>
          <span>Codomain</span>
        </div>
        <svg
          className={`mt-2 block w-full min-w-0 max-w-full ${compact ? 'max-h-52' : 'max-h-72'}`}
          viewBox={`0 0 600 ${height}`}
          role="img"
          aria-label={formatMappingSummary(domain, edges)}
        >
          <defs>
            <marker id={markerId} markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">
              <path className="fill-accent-teal" d="M 0 0 L 8 4 L 0 8 z" />
            </marker>
          </defs>

          {edges.map((edge) => {
            const fromIndex = domain.indexOf(edge.from)
            const toIndex = codomain.indexOf(edge.to)
            if (fromIndex < 0 || toIndex < 0) return null

            return (
              <line
                key={`${edge.from}-${edge.to}`}
                className="stroke-accent-teal"
                x1="105"
                y1={yFor(fromIndex, domain.length)}
                x2="495"
                y2={yFor(toIndex, codomain.length)}
                strokeWidth="3"
                markerEnd={`url(#${markerId})`}
              />
            )
          })}

          {domain.map((value, index) => {
            const selected = selectedInput === value
            return (
              <g key={value}>
                <circle
                  className={
                    selected
                      ? 'fill-accent-gold/20 stroke-accent-gold'
                      : 'fill-panel stroke-line-strong'
                  }
                  cx="75"
                  cy={yFor(index, domain.length)}
                  r="24"
                  strokeWidth={selected ? 4 : 2}
                />
                <text
                  className="fill-ink text-sm font-semibold"
                  x="75"
                  y={yFor(index, domain.length) + 5}
                  textAnchor="middle"
                >
                  {value}
                </text>
              </g>
            )
          })}

          {codomain.map((value, index) => (
            <g key={value}>
              <circle
                className="fill-panel stroke-line-strong"
                cx="525"
                cy={yFor(index, codomain.length)}
                r="24"
                strokeWidth="2"
              />
              <text
                className="fill-ink text-sm font-semibold"
                x="525"
                y={yFor(index, codomain.length) + 5}
                textAnchor="middle"
              >
                {value}
              </text>
            </g>
          ))}
        </svg>
        <p className="sr-only">{formatMappingSummary(domain, edges)}</p>
      </div>

      {editable ? (
        <div className="grid gap-3 rounded-lg border border-line bg-raised p-3">
          <fieldset>
            <legend className="text-xs font-bold uppercase tracking-widest text-faint">
              1. Select an input
            </legend>
            <div className="mt-2 flex flex-wrap gap-2">
              {domain.map((input) => (
                <button
                  key={input}
                  className={`${nodeButtonClass} ${
                    selectedInput === input
                      ? 'border-accent-gold/60 bg-accent-gold/10 text-ink'
                      : 'border-line bg-panel text-muted hover:border-accent-gold/50 hover:text-ink'
                  }`}
                  type="button"
                  onClick={() => onSelectInput?.(input)}
                  aria-pressed={selectedInput === input}
                >
                  Input {input}
                </button>
              ))}
            </div>
          </fieldset>

          <fieldset disabled={!selectedInput}>
            <legend className="text-xs font-bold uppercase tracking-widest text-faint">
              2. {singleOutput ? 'Choose an output' : 'Toggle outputs'} for{' '}
              {selectedInput ? `input ${selectedInput}` : 'the selected input'}
            </legend>
            <div className="mt-2 flex flex-wrap gap-2">
              {codomain.map((output) => {
                const connected = selectedTargets.has(output)
                return (
                  <button
                    key={output}
                    className={`${nodeButtonClass} ${
                      connected
                        ? 'border-accent-teal/60 bg-accent-teal/10 text-ink'
                        : 'border-line bg-panel text-muted hover:border-accent-teal/50 hover:text-ink'
                    } disabled:cursor-not-allowed disabled:opacity-50`}
                    type="button"
                    onClick={() => selectedInput && onToggleEdge?.(selectedInput, output)}
                    aria-pressed={connected}
                  >
                    {connected ? 'Connected to' : singleOutput ? 'Choose' : 'Connect to'} {output}
                  </button>
                )
              })}
            </div>
          </fieldset>
        </div>
      ) : (
        <p className="text-sm leading-relaxed text-muted">{formatMappingSummary(domain, edges)}</p>
      )}
    </div>
  )
}

function formatMappingSummary(domain: readonly string[], edges: readonly MappingEdge[]) {
  return domain
    .map((input) => {
      const outputs = edges.filter((edge) => edge.from === input).map((edge) => edge.to)
      return `${input} maps to ${outputs.length > 0 ? outputs.join(' and ') : 'no output'}`
    })
    .join('; ')
}
