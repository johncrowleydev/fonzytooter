import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ArrayStructureExplorer } from './ArrayStructureExplorer'
import { formatShape } from './ArrayVisual'
import { IndexingShapeVisualizer } from './IndexingShapeVisualizer'
import { ViewCopyExplorer } from './ViewCopyExplorer'

afterEach(cleanup)

describe('array interactives', () => {
  it('formats scalar and one-dimensional shapes without ambiguity', () => {
    expect(formatShape([])).toBe('scalar')
    expect(formatShape([4])).toBe('(4,)')
    expect(formatShape([2, 3, 4])).toBe('(2, 3, 4)')
  })

  it('connects a selected axis to its length', () => {
    render(<ArrayStructureExplorer />)

    fireEvent.click(screen.getByRole('button', { name: 'batch (2, 3, 4)' }))
    fireEvent.click(screen.getByRole('button', { name: 'axis 1' }))
    fireEvent.click(screen.getByRole('button', { name: '3' }))

    expect(screen.getByRole('status').textContent).toContain('axis 1 has 3 positions')
  })

  it('distinguishes integer indexing from a one-position slice', () => {
    render(<IndexingShapeVisualizer />)

    fireEvent.click(screen.getByRole('button', { name: 'x[0:1]' }))

    expect(screen.getByRole('status').textContent).toContain('result shape: (1, 4)')
    expect(screen.getByRole('status').textContent).toContain('preserves axis 0')
  })

  it('propagates view mutation but isolates copy mutation', () => {
    render(<ViewCopyExplorer />)

    fireEvent.click(screen.getByRole('button', { name: 'Set middle[0] = 999' }))
    expect(screen.getByRole('img', { name: /original contains 10, 999, 30, 40/ })).not.toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Set independent[0] = 555' }))
    expect(screen.getByRole('img', { name: /independent.*contains 555, 30/ })).not.toBeNull()
    expect(screen.getByRole('img', { name: /original contains 10, 999, 30, 40/ })).not.toBeNull()
  })
})
