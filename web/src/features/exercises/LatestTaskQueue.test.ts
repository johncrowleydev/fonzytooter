import { describe, expect, it, vi } from 'vitest'
import { LatestTaskQueue } from './LatestTaskQueue'

describe('LatestTaskQueue', () => {
  it('serializes work and coalesces queued tasks to the latest one', async () => {
    let finishFirst: (() => void) | undefined
    const firstFinished = new Promise<void>((resolve) => {
      finishFirst = resolve
    })
    const completed: string[] = []
    const queue = new LatestTaskQueue()
    const first = vi.fn(async () => {
      await firstFinished
      completed.push('first')
    })
    const superseded = vi.fn(async () => {
      completed.push('superseded')
    })
    const latest = vi.fn(async () => {
      completed.push('latest')
    })

    queue.enqueue(first)
    queue.enqueue(superseded)
    queue.enqueue(latest)

    expect(first).toHaveBeenCalledOnce()
    expect(superseded).not.toHaveBeenCalled()
    expect(latest).not.toHaveBeenCalled()

    finishFirst?.()
    await vi.waitFor(() => expect(latest).toHaveBeenCalledOnce())

    expect(superseded).not.toHaveBeenCalled()
    expect(completed).toEqual(['first', 'latest'])
  })
})
