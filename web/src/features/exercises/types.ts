export type PythonExerciseDefinition = {
  id: string
  title: string
  objectiveIds: string[]
  starterCode: string
  visibleTests?: string
  hiddenTests?: string
}

export type PythonRunResult = {
  stdout: string
  stderr: string
  durationMs?: number
}

export type PythonTestResult = {
  name: string
  passed: boolean
  message?: string
}

export type PythonCheckResult = PythonRunResult & {
  tests: PythonTestResult[]
}

export interface PythonRunner {
  run(code: string): Promise<PythonRunResult>
  check(code: string, exercise: PythonExerciseDefinition): Promise<PythonCheckResult>
}

// The only in-app implementation of this interface should be the Pyodide Web Worker runner.
// Do not add a backend Python execution implementation; see docs/coding-exercises.md.
