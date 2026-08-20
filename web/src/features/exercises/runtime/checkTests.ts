import type { PythonCheckRequest, PythonCheckResult, PythonRunResult } from '../types'

export async function runCheckTests(
  request: PythonCheckRequest,
  execute: (code: string) => Promise<PythonRunResult>,
): Promise<PythonCheckResult> {
  const startedAt = performance.now()
  const tests = []
  let stdout = ''
  let stderr = ''
  for (const test of request.tests) {
    const result = await execute(`${request.code}\n\n${test.code}`)
    stdout += result.stdout
    stderr += result.stderr
    const assertion =
      result.error?.name === 'AssertionError' ||
      result.error?.message.includes('AssertionError') ||
      result.error?.traceback?.includes('AssertionError')
    tests.push({
      testId: test.id,
      title: test.title,
      visibility: test.visibility,
      status: result.error
        ? assertion
          ? ('failed' as const)
          : ('error' as const)
        : ('passed' as const),
      message: result.error?.message ?? '',
      durationMs: result.durationMs,
    })
  }
  return { stdout, stderr, durationMs: Math.round(performance.now() - startedAt), tests }
}
