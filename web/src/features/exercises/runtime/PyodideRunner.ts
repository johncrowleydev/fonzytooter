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
}

export class PyodideRunner implements PythonRunner {
  private worker: WorkerLike | undefined
  private nextRequest = 0
  private readonly pending = new Map<string, PendingRequest>()

  constructor(
    private readonly workerFactory: WorkerFactory = () =>
      new Worker(new URL('./python.worker.ts', import.meta.url), { type: 'module' }),
    private readonly timeoutMs = 10_000,
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
      const timer = setTimeout(() => {
        this.pending.delete(id)
        reject(new Error(`Python execution exceeded ${this.timeoutMs}ms and was stopped`))
        this.failPending(new Error('Python execution was interrupted when the worker restarted'))
        this.restartWorker()
      }, this.timeoutMs)
      this.pending.set(id, {
        resolve: resolve as (value: PythonRunResult | PythonCheckResult) => void,
        reject,
        timer,
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

  private restartWorker() {
    if (!this.worker) return
    this.worker.removeEventListener('message', this.handleMessage)
    this.worker.removeEventListener('error', this.handleWorkerFailure)
    this.worker.terminate()
    this.worker = undefined
  }
}
