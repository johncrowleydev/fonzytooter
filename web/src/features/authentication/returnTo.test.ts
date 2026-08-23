import { describe, expect, it } from 'vitest'
import { safeReturnTo } from './returnTo'

describe('safeReturnTo', () => {
  it('preserves application-local paths, queries, and fragments', () => {
    expect(safeReturnTo('/courses/ai-ml?view=outline#lesson')).toBe(
      '/courses/ai-ml?view=outline#lesson',
    )
  })

  it.each([
    'https://attacker.example/path',
    '//attacker.example/path',
    '/\\attacker.example/path',
    'javascript:alert(1)',
  ])('rejects an unsafe return destination: %s', (destination) => {
    expect(safeReturnTo(destination, '/curriculum')).toBe('/curriculum')
  })
})
