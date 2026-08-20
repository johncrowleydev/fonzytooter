/// <reference lib="webworker" />

import { loadPyodide, type PyodideInterface } from 'pyodide'
import type { PythonExecutionError, PythonRunResult } from '../types'
import { runCheckTests } from './checkTests'
import type { PythonWorkerRequest, PythonWorkerResponse } from './protocol'

export const PYODIDE_VERSION = '314.0.5'
export const PYODIDE_INDEX_URL = `https://cdn.jsdelivr.net/pyodide/v${PYODIDE_VERSION}/full/`

let pyodidePromise: Promise<PyodideInterface> | undefined

function getPyodide() {
  pyodidePromise ??= loadPyodide({ indexURL: PYODIDE_INDEX_URL })
  return pyodidePromise
}

function errorDetails(error: unknown): PythonExecutionError {
  if (error instanceof Error) {
    const lines = error.message
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
    return {
      name: error.name,
      message: lines.at(-1) ?? error.message,
      traceback: error.stack,
    }
  }
  return { name: 'Error', message: String(error) }
}

async function execute(code: string): Promise<PythonRunResult> {
  const startedAt = performance.now()
  const pyodide = await getPyodide()
  let stdout = ''
  let stderr = ''
  pyodide.setStdout({ batched: (text) => (stdout += `${text}\n`) })
  pyodide.setStderr({ batched: (text) => (stderr += `${text}\n`) })
  const globals = pyodide.toPy({})
  try {
    await pyodide.loadPackagesFromImports(code)
    const result = await pyodide.runPythonAsync(code, { globals })
    if (result && typeof result === 'object' && 'destroy' in result) result.destroy()
    return { stdout, stderr, durationMs: Math.round(performance.now() - startedAt) }
  } catch (error) {
    return {
      stdout,
      stderr,
      durationMs: Math.round(performance.now() - startedAt),
      error: errorDetails(error),
    }
  } finally {
    globals.destroy()
  }
}

self.addEventListener('message', async (event: MessageEvent<PythonWorkerRequest>) => {
  const request = event.data
  try {
    const response: PythonWorkerResponse =
      request.type === 'run'
        ? { id: request.id, type: 'run-result', payload: await execute(request.payload.code) }
        : {
            id: request.id,
            type: 'check-result',
            payload: await runCheckTests(request.payload, execute),
          }
    self.postMessage(response)
  } catch (error) {
    self.postMessage({
      id: request.id,
      type: 'worker-error',
      error: errorDetails(error),
    } satisfies PythonWorkerResponse)
  }
})

export {}
