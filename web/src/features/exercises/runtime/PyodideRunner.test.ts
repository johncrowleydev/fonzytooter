import { describe, expect, it, vi } from 'vitest'
import type { PythonWorkerRequest, PythonWorkerResponse } from './protocol'
import { PyodideRunner } from './PyodideRunner'

class FakeWorker {
  requests: PythonWorkerRequest[] = []
  terminated = false
  private messageListeners = new Set<(event: MessageEvent<PythonWorkerResponse>) => void>()
  private errorListeners = new Set<(event: ErrorEvent) => void>()

  postMessage(request: PythonWorkerRequest) {
    this.requests.push(request)
  }

  terminate() {
    this.terminated = true
  }

  addEventListener(type: string, listener: EventListenerOrEventListenerObject) {
    if (type === 'message') this.messageListeners.add(listener as never)
    if (type === 'error') this.errorListeners.add(listener as never)
  }

  removeEventListener(type: string, listener: EventListenerOrEventListenerObject) {
    if (type === 'message') this.messageListeners.delete(listener as never)
    if (type === 'error') this.errorListeners.delete(listener as never)
  }

  respond(response: PythonWorkerResponse) {
    for (const listener of this.messageListeners)
      listener({ data: response } as MessageEvent<PythonWorkerResponse>)
  }

  fail() {
    for (const listener of this.errorListeners) listener(new ErrorEvent('error'))
  }
}

describe('PyodideRunner', () => {
  it('correlates out-of-order worker responses by request ID', async () => {
    const worker = new FakeWorker()
    const runner = new PyodideRunner(() => worker as unknown as Worker)
    const first = runner.run({ code: 'print(1)' })
    const second = runner.run({ code: 'print(2)' })

    worker.respond({
      id: worker.requests[1].id,
      type: 'run-result',
      payload: { stdout: '2\n', stderr: '', durationMs: 2 },
    })
    worker.respond({
      id: worker.requests[0].id,
      type: 'run-result',
      payload: { stdout: '1\n', stderr: '', durationMs: 1 },
    })

    await expect(first).resolves.toMatchObject({ stdout: '1\n' })
    await expect(second).resolves.toMatchObject({ stdout: '2\n' })
    runner.dispose()
  })

  it('reports initialization failure and recreates the worker', async () => {
    const workers: FakeWorker[] = []
    const runner = new PyodideRunner(() => {
      const worker = new FakeWorker()
      workers.push(worker)
      return worker as unknown as Worker
    })
    const run = runner.run({ code: 'pass' })
    workers[0].fail()
    await expect(run).rejects.toThrow('failed to initialize')
    expect(workers[0].terminated).toBe(true)

    const recovered = runner.run({ code: 'print("ready")' })
    workers[1].respond({
      id: workers[1].requests[0].id,
      type: 'run-result',
      payload: { stdout: 'ready\n', stderr: '', durationMs: 1 },
    })
    await expect(recovered).resolves.toMatchObject({ stdout: 'ready\n' })
    runner.dispose()
  })

  it('recreates the worker after a structured initialization error', async () => {
    const workers: FakeWorker[] = []
    const runner = new PyodideRunner(() => {
      const worker = new FakeWorker()
      workers.push(worker)
      return worker as unknown as Worker
    })
    const failed = runner.run({ code: 'pass' })
    workers[0].respond({
      id: workers[0].requests[0].id,
      type: 'worker-error',
      error: { name: 'InitializationError', message: 'runtime unavailable' },
    })
    await expect(failed).rejects.toThrow('runtime unavailable')
    expect(workers[0].terminated).toBe(true)

    const recovered = runner.run({ code: 'pass' })
    workers[1].respond({
      id: workers[1].requests[0].id,
      type: 'run-result',
      payload: { stdout: '', stderr: '', durationMs: 1 },
    })
    await expect(recovered).resolves.toMatchObject({ durationMs: 1 })
    runner.dispose()
  })

  it('terminates a timed-out worker and recovers on the next request', async () => {
    vi.useFakeTimers()
    const workers: FakeWorker[] = []
    const runner = new PyodideRunner(
      () => {
        const worker = new FakeWorker()
        workers.push(worker)
        return worker as unknown as Worker
      },
      50,
      200,
    )
    const timedOut = runner.run({ code: 'while True: pass' })
    await vi.advanceTimersByTimeAsync(100)
    expect(workers[0].terminated).toBe(false)
    workers[0].respond({
      id: workers[0].requests[0].id,
      type: 'execution-started',
    })
    const timeoutExpectation = expect(timedOut).rejects.toThrow('was stopped')
    await vi.advanceTimersByTimeAsync(51)
    await timeoutExpectation
    expect(workers[0].terminated).toBe(true)

    const recovered = runner.run({ code: 'print(3)' })
    workers[1].respond({
      id: workers[1].requests[0].id,
      type: 'run-result',
      payload: { stdout: '3\n', stderr: '', durationMs: 1 },
    })
    await expect(recovered).resolves.toMatchObject({ stdout: '3\n' })
    runner.dispose()
    vi.useRealTimers()
  })
})
