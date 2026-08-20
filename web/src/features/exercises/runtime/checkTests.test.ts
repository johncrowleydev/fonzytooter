import { describe, expect, it, vi } from 'vitest'
import { runCheckTests } from './checkTests'

describe('runCheckTests', () => {
  it('normalizes visible, hidden, assertion, and runtime results in authored order', async () => {
    const execute = vi
      .fn()
      .mockResolvedValueOnce({ stdout: 'visible\n', stderr: '', durationMs: 2 })
      .mockResolvedValueOnce({
        stdout: '',
        stderr: '',
        durationMs: 3,
        error: { name: 'PythonError', message: 'AssertionError' },
      })
      .mockResolvedValueOnce({
        stdout: '',
        stderr: 'diagnostic\n',
        durationMs: 4,
        error: { name: 'PythonError', message: 'ValueError' },
      })

    const result = await runCheckTests(
      {
        code: 'value = 2',
        tests: [
          { id: 'visible', title: 'Visible', visibility: 'visible', code: 'assert value == 2' },
          { id: 'hidden', title: 'Hidden', visibility: 'hidden', code: 'assert value == 4' },
          { id: 'runtime', title: 'Runtime', visibility: 'hidden', code: 'raise ValueError()' },
        ],
      },
      execute,
    )

    expect(execute).toHaveBeenNthCalledWith(1, 'value = 2\n\nassert value == 2')
    expect(execute).toHaveBeenNthCalledWith(2, 'value = 2\n\nassert value == 4')
    expect(execute).toHaveBeenNthCalledWith(3, 'value = 2\n\nraise ValueError()')
    expect(result.tests.map((test) => [test.testId, test.visibility, test.status])).toEqual([
      ['visible', 'visible', 'passed'],
      ['hidden', 'hidden', 'failed'],
      ['runtime', 'hidden', 'error'],
    ])
    expect(result.stdout).toBe('visible\n')
    expect(result.stderr).toBe('diagnostic\n')
    expect(result.tests.some((test) => 'code' in test)).toBe(false)
  })
})
