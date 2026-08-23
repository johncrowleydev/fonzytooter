import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { BroadcastDebugChallenge } from './BroadcastDebugChallenge'
import { broadcastShapes, BroadcastShapeLab } from './BroadcastShapeLab'

afterEach(cleanup)

describe('broadcasting interactives', () => {
  it('implements right-aligned equal-or-one compatibility', () => {
    expect(broadcastShapes([5, 4], [4])).toMatchObject({ compatible: true, shape: [5, 4] })
    expect(broadcastShapes([2, 5, 4], [1, 4])).toMatchObject({ compatible: true, shape: [2, 5, 4] })
    expect(broadcastShapes([8, 3], [8])).toMatchObject({ compatible: false, shape: null })
    expect(broadcastShapes([4, 1, 6], [3, 6])).toMatchObject({ compatible: true, shape: [4, 3, 6] })
  })

  it('requires result-shape prediction before revealing a compatible result', () => {
    render(<BroadcastShapeLab />)
    fireEvent.click(screen.getByRole('button', { name: 'Case 6' }))
    fireEvent.click(screen.getByRole('button', { name: 'Compatible' }))

    expect(screen.getByText('Predict the result shape')).not.toBeNull()
    expect(screen.queryByText('result: (4, 3, 6)')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: '(3, 6)' }))
    expect(screen.getByRole('alert').textContent).toContain('larger compatible dimension')
    expect(screen.queryByText('result: (4, 3, 6)')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: '(4, 3, 6)' }))
    expect(screen.getByRole('status').textContent).toContain('result: (4, 3, 6)')
    expect(screen.getByRole('status').textContent).toContain('1 vs 3')
  })

  it('treats an incompatibility prediction as terminal', () => {
    render(<BroadcastShapeLab />)
    fireEvent.click(screen.getByRole('button', { name: 'Case 5' }))
    fireEvent.click(screen.getByRole('button', { name: 'Incompatible' }))

    expect(screen.queryByText('Predict the result shape')).toBeNull()
    expect(screen.getByRole('status').textContent).toContain('broadcasting fails')
    expect(screen.getByRole('status').textContent).toContain('3 vs 8')
  })

  it('uses keepdims to repair row-wise centering', () => {
    render(<BroadcastDebugChallenge />)
    fireEvent.click(screen.getByRole('button', { name: 'Use mean(axis=1, keepdims=True)' }))
    expect(screen.getByRole('status').textContent).toContain('(4, 1)')
  })

  it('recognizes a legal pairwise broadcast as the wrong grid', () => {
    render(<BroadcastDebugChallenge />)
    fireEvent.click(screen.getByRole('button', { name: 'Legal but wrong' }))
    fireEvent.click(screen.getByRole('button', { name: 'Flatten left to (3,) before adding' }))
    expect(screen.getByRole('status').textContent).toContain('Legal broadcasting is not proof')
  })
})
