import type {
  PythonCheckRequest,
  PythonCheckResult,
  PythonRunner,
  PythonRunRequest,
  PythonRunResult,
} from '../types'
import type { PythonWorkerRequest, PythonWorkerResponse } from './protocol'

type WorkerLike = Pick<
  Worker,
  'postMessage' | 'terminate' | 'addEventListener' | 'removeEventListener'
>
type WorkerFactory = () => WorkerLike
type PendingRequest = {
  resolve: (value: PythonRunResult | PythonCheckResult) => void
  reject: (reason: Error) => void
  timer: ReturnType<typeof setTimeout>
  executionStarted: boolean
}

export class PyodideRunner implements PythonRunner {
  private worker: WorkerLike | undefined
  private nextRequest = 0
  private readonly pending = new Map<string, PendingRequest>()

  constructor(
    private readonly workerFactory: WorkerFactory = () =>
      new Worker(new URL('./python.worker.ts', import.meta.url), { type: 'module' }),
    private readonly timeoutMs = 10_000,
    private readonly initializationTimeoutMs = 120_000,
  ) {}

  run(request: PythonRunRequest): Promise<PythonRunResult> {
    return this.request<PythonRunResult>({ type: 'run', payload: request })
  }

  check(request: PythonCheckRequest): Promise<PythonCheckResult> {
    return this.request<PythonCheckResult>({ type: 'check', payload: request })
  }

  dispose() {
    this.worker?.terminate()
    this.worker = undefined
    for (const pending of this.pending.values()) {
      clearTimeout(pending.timer)
      pending.reject(new Error('Python runner was disposed'))
    }
    this.pending.clear()
  }

  private request<Result>(request: Omit<PythonWorkerRequest, 'id'>): Promise<Result> {
    const id = `python-${++this.nextRequest}`
    const worker = this.getWorker()
    return new Promise<Result>((resolve, reject) => {
      const timer = this.timeout(id, reject, 'initialization', this.initializationTimeoutMs)
      this.pending.set(id, {
        resolve: resolve as (value: PythonRunResult | PythonCheckResult) => void,
        reject,
        timer,
        executionStarted: false,
      })
      worker.postMessage({ ...request, id } as PythonWorkerRequest)
    })
  }

  private getWorker() {
    if (!this.worker) {
      this.worker = this.workerFactory()
      this.worker.addEventListener('message', this.handleMessage)
      this.worker.addEventListener('error', this.handleWorkerFailure)
    }
    return this.worker
  }

  private readonly handleMessage = (event: MessageEvent<PythonWorkerResponse>) => {
    const response = event.data
    const pending = this.pending.get(response.id)
    if (!pending) return
    if (response.type === 'execution-started') {
      if (pending.executionStarted) return
      clearTimeout(pending.timer)
      pending.executionStarted = true
      pending.timer = this.timeout(response.id, pending.reject, 'execution', this.timeoutMs)
      return
    }
    clearTimeout(pending.timer)
    this.pending.delete(response.id)
    if (response.type === 'worker-error') {
      pending.reject(new Error(`${response.error.name}: ${response.error.message}`))
      this.failPending(new Error('Python runtime was restarted after a worker failure'))
      this.restartWorker()
      return
    }
    pending.resolve(response.payload)
  }

  private readonly handleWorkerFailure = () => {
    this.failPending(new Error('Python runtime failed to initialize'))
    this.restartWorker()
  }

  private failPending(failure: Error) {
    for (const pending of this.pending.values()) {
      clearTimeout(pending.timer)
      pending.reject(failure)
    }
    this.pending.clear()
  }

  private timeout(
    id: string,
    reject: (reason: Error) => void,
    phase: 'initialization' | 'execution',
    timeoutMs: number,
  ) {
    return setTimeout(() => {
      this.pending.delete(id)
      const message =
        phase === 'initialization'
          ? `Python runtime initialization exceeded ${timeoutMs}ms and was stopped`
          : `Python execution exceeded ${timeoutMs}ms and was stopped`
      reject(new Error(message))
      this.failPending(new Error(`Python ${phase} was interrupted when the worker restarted`))
      this.restartWorker()
    }, timeoutMs)
  }

  private restartWorker() {
    if (!this.worker) return
    this.worker.removeEventListener('message', this.handleMessage)
    this.worker.removeEventListener('error', this.handleWorkerFailure)
    this.worker.terminate()
    this.worker = undefined
  }
}
