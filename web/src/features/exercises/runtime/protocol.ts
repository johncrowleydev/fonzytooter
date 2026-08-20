import type {
  PythonCheckRequest,
  PythonCheckResult,
  PythonExecutionError,
  PythonRunRequest,
  PythonRunResult,
} from '../types'

export type PythonWorkerRequest =
  | { id: string; type: 'run'; payload: PythonRunRequest }
  | { id: string; type: 'check'; payload: PythonCheckRequest }

export type PythonWorkerResponse =
  | { id: string; type: 'run-result'; payload: PythonRunResult }
  | { id: string; type: 'check-result'; payload: PythonCheckResult }
  | { id: string; type: 'worker-error'; error: PythonExecutionError }
