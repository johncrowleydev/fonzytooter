import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { LoopVectorizationExplorer } from './LoopVectorizationExplorer'
import { OperationKindCheck } from './OperationKindCheck'
import { reducedShape, reductionGroups, ReductionAxisExplorer } from './ReductionAxisExplorer'

afterEach(cleanup)

describe('vectorization interactives', () => {
  it('removes exactly the named reduction axis', () => {
    expect(reducedShape([2, 3, 4], 0)).toEqual([3, 4])
    expect(reducedShape([2, 3, 4], 1)).toEqual([2, 4])
    expect(reducedShape([2, 3, 4], 2)).toEqual([2, 3])
  })

  it('groups row-major values along the collapsed axis', () => {
    const values = [80, 90, 100, 70, 85, 95]
    expect(reductionGroups([2, 3], 0, values)).toEqual([
      [80, 70],
      [90, 85],
      [100, 95],
    ])
    expect(reductionGroups([2, 3], 1, values)).toEqual([
      [80, 90, 100],
      [70, 85, 95],
    ])
  })

  it('reveals the groups after a correct 3D shape prediction', () => {
    render(<ReductionAxisExplorer />)
    fireEvent.click(screen.getByRole('button', { name: 'batch (2, 3, 4)' }))
    fireEvent.click(screen.getByRole('button', { name: 'axis 1' }))
    fireEvent.click(screen.getByRole('button', { name: '(2, 4)' }))
    expect(screen.getByRole('status').textContent).toContain('(2, 3, 4) → (2, 4)')
    expect(screen.getByRole('status').textContent).toContain('mean(')
  })

  it('makes compiled delegation explicit', () => {
    render(<LoopVectorizationExplorer />)
    fireEvent.click(screen.getByRole('button', { name: 'Delegate whole array to NumPy' }))
    expect(screen.getByRole('status').textContent).toContain('compiled machinery')
  })

  it('classifies comparison followed by reduction as composition', () => {
    render(<OperationKindCheck />)
    fireEvent.click(screen.getByRole('button', { name: '5' }))
    fireEvent.click(screen.getByRole('button', { name: 'composition' }))
    expect(screen.getByRole('status').textContent).toContain('boolean array')
    expect(screen.getByRole('status').textContent).toContain('scalar')
  })
})
