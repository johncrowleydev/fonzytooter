export class LatestTaskQueue {
  private pending: (() => Promise<void>) | undefined
  private running = false

  enqueue(task: () => Promise<void>) {
    this.pending = task
    if (!this.running) void this.drain()
  }

  private async drain() {
    this.running = true
    while (this.pending) {
      const task = this.pending
      this.pending = undefined
      await task()
    }
    this.running = false
  }
}
