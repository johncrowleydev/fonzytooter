export type PythonExecutionError = {
  name: string
  message: string
  traceback?: string
}

export type PythonRunRequest = { code: string }

export type PythonRunResult = {
  stdout: string
  stderr: string
  durationMs: number
  error?: PythonExecutionError
}

export type PythonCheckTest = {
  id: string
  title: string
  visibility: 'visible' | 'hidden'
  code: string
}

export type PythonCheckRequest = { code: string; tests: PythonCheckTest[] }

export type PythonTestResult = {
  testId: string
  title: string
  visibility: 'visible' | 'hidden'
  status: 'passed' | 'failed' | 'error'
  message: string
  durationMs: number
}

export type PythonCheckResult = {
  stdout: string
  stderr: string
  durationMs: number
  tests: PythonTestResult[]
  error?: PythonExecutionError
}

export interface PythonRunner {
  run(request: PythonRunRequest): Promise<PythonRunResult>
  check(request: PythonCheckRequest): Promise<PythonCheckResult>
  dispose(): void
}
