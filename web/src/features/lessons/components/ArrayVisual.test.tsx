import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ValueGrid } from './ArrayVisual'

afterEach(cleanup)

describe('ValueGrid accessibility', () => {
  /**
   * `role="img"` prunes descendants from the accessibility tree. The cells each used to carry an
   * aria-label announcing ", selected" and none of it was reachable, which hid exactly the thing
   * the indexing lesson teaches.
   */
  it('names the selected values in the image label', () => {
    render(
      <ValueGrid
        values={[0, 1, 2, 3, 4, 5]}
        columns={3}
        selected={[1, 4]}
        label="Source x; values selected by x[:, 1] are marked"
      />,
    )

    const image = screen.getByRole('img')
    expect(image.getAttribute('aria-label')).toBe(
      'Source x; values selected by x[:, 1] are marked. Selected values: 1, 4.',
    )
  })

  it('leaves the label alone when nothing is selected', () => {
    render(<ValueGrid values={[7, 8]} columns={2} label="original contains 7, 8" />)

    expect(screen.getByRole('img').getAttribute('aria-label')).toBe('original contains 7, 8')
  })

  it('hides the cells explicitly rather than relying on role="img" pruning them', () => {
    const { container } = render(
      <ValueGrid values={[0, 1, 2]} columns={3} selected={[2]} label="grid" />,
    )
    const cells = container.querySelectorAll('span')

    expect(cells.length).toBe(3)
    for (const cell of cells) {
      expect(cell.getAttribute('aria-hidden')).toBe('true')
    }
  })

  it('still renders every value visually', () => {
    const { container } = render(
      <ValueGrid values={[10, 20, 30, 40]} columns={2} selected={[0]} label="grid" />,
    )

    expect(container.textContent).toBe('10203040')
  })
})
