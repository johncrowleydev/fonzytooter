import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { executeNotebookCell, KernelStateExplorer } from './KernelStateExplorer'
import { seededSequence, SeedReproducibilityExplorer } from './SeedReproducibilityExplorer'

afterEach(cleanup)

describe('Jupyter experiment interactives', () => {
  it('keeps visible source separate from executed kernel memory', () => {
    let memory = executeNotebookCell({}, 0, 2)
    memory = executeNotebookCell(memory, 1, 2)
    memory = executeNotebookCell(memory, 2, 5)
    expect(memory).toEqual({ rate: 2, result: 20 })
  })

  it('restart-and-run-all recovers a clean top-to-bottom result', () => {
    render(<KernelStateExplorer />)
    fireEvent.change(screen.getByLabelText('Visible rate source'), {
      target: { value: '5' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Restart and run all top to bottom' }))
    expect(screen.getByLabelText('Cell 3 output').textContent).toBe('50')
    expect(screen.getByText('rate = 5')).not.toBeNull()
  })

  it('keeps Cell 3 output when Cells 1 and 2 run afterward', () => {
    render(<KernelStateExplorer />)
    fireEvent.click(screen.getByRole('button', { name: 'Restart and run all top to bottom' }))
    expect(screen.getByLabelText('Cell 3 output').textContent).toBe('20')

    fireEvent.change(screen.getByLabelText('Visible rate source'), {
      target: { value: '5' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Run cell 1' }))
    fireEvent.click(screen.getByRole('button', { name: 'Run cell 2' }))

    expect(screen.getByLabelText('Cell 3 output').textContent).toBe('20')
    expect(screen.getByText('result = 50')).not.toBeNull()
  })

  it('repeats a sequence only from the same starting state', () => {
    expect(seededSequence(7, 5)).toEqual(seededSequence(7, 5))
    expect(seededSequence(7, 5)).not.toEqual(seededSequence(99, 5))
    expect(seededSequence(7, 5, 1)).not.toEqual(seededSequence(7, 5))
  })

  it('shows that advancing one same-seed generator changes later draws', () => {
    render(<SeedReproducibilityExplorer />)
    expect(screen.getByRole('status').textContent).toContain('Sequences match')
    fireEvent.click(screen.getByRole('button', { name: 'Advance generator A once' }))
    expect(screen.getByRole('status').textContent).toContain('later state')
  })
})
