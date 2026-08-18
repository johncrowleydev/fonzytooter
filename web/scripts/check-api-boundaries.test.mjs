import assert from 'node:assert/strict'
import path from 'node:path'
import { describe, it } from 'node:test'
import { fileURLToPath } from 'node:url'
import { findApiBoundaryViolations } from './check-api-boundaries.mjs'

const fixturesRoot = path.join(path.dirname(fileURLToPath(import.meta.url)), 'fixtures', 'api-boundaries')

function checkFixture(name) {
  return findApiBoundaryViolations(path.join(fixturesRoot, name, 'src'))
}

describe('API boundary checker', () => {
  it('allows generated clients, Zod inference aliases, and useQueryClient', () => {
    assert.deepEqual(checkFixture('pass'), [])
  })

  it('rejects feature-level fetch and hard-coded application URLs', () => {
    const messages = checkFixture('fetch')
    assert.ok(messages.some((violation) => violation.message.includes('raw global fetch')))
    assert.ok(messages.some((violation) => violation.message.includes('hard-coded /api URL')))
  })

  it('rejects axios imports', () => {
    assert.ok(checkFixture('axios').some((violation) => violation.message.includes('axios')))
  })

  it('rejects XMLHttpRequest construction', () => {
    assert.ok(checkFixture('xhr').some((violation) => violation.message.includes('XMLHttpRequest')))
  })

  it('rejects endpoint-level React Query hooks', () => {
    assert.ok(checkFixture('react-query').some((violation) => violation.message.includes('endpoint-level React Query')))
  })

  it('rejects handwritten object-shaped API DTOs', () => {
    assert.ok(checkFixture('dto').some((violation) => violation.message.includes('object-shaped API')))
  })
})
