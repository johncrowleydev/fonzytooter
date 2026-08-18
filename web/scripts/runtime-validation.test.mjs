import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { test } from 'node:test'
import { fileURLToPath, pathToFileURL } from 'node:url'
import * as ts from 'typescript-compiler-api'

const webRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const generatedRoot = path.join(webRoot, 'src', 'api', 'generated')

function rewriteRelativeImports(source) {
  return source.replace(/(from\s+['"])(\.[^'"]+)(['"])/g, (_match, prefix, importPath, suffix) => {
    if (importPath === './schemas') return `${prefix}./schemas/index.mjs${suffix}`
    if (importPath.endsWith('.js') || importPath.endsWith('.mjs')) return `${prefix}${importPath}${suffix}`
    return `${prefix}${importPath}.mjs${suffix}`
  })
}

function transpile(source, fileName) {
  const result = ts.transpileModule(source, {
    compilerOptions: {
      esModuleInterop: true,
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
    fileName,
  })
  return rewriteRelativeImports(result.outputText)
}

function copyGeneratedModule(sourcePath, targetPath) {
  fs.mkdirSync(path.dirname(targetPath), { recursive: true })
  fs.writeFileSync(targetPath, transpile(fs.readFileSync(sourcePath, 'utf8'), sourcePath))
}

async function importGeneratedClient() {
  const tempRoot = fs.mkdtempSync(path.join(webRoot, '.api-runtime-test-'))
  copyGeneratedModule(path.join(generatedRoot, 'endpoints.ts'), path.join(tempRoot, 'endpoints.mjs'))
  copyGeneratedModule(path.join(generatedRoot, 'schemas', 'index.ts'), path.join(tempRoot, 'schemas', 'index.mjs'))

  for (const fileName of fs.readdirSync(path.join(generatedRoot, 'schemas'))) {
    if (!fileName.endsWith('.ts') || fileName === 'index.ts') continue
    copyGeneratedModule(
      path.join(generatedRoot, 'schemas', fileName),
      path.join(tempRoot, 'schemas', fileName.replace(/\.ts$/, '.mjs')),
    )
  }

  try {
    return await import(`${pathToFileURL(path.join(tempRoot, 'endpoints.mjs')).href}?test=${Date.now()}`)
  } finally {
    fs.rmSync(tempRoot, { force: true, recursive: true })
  }
}

test('generated health fetch rejects an invalid successful JSON response with ZodError', async () => {
  const client = await importGeneratedClient()
  const originalFetch = globalThis.fetch

  try {
    globalThis.fetch = async () =>
      new Response(JSON.stringify({ status: 123 }), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      })

    await assert.rejects(() => client.getHealth(), (error) => error?.name === 'ZodError')
  } finally {
    globalThis.fetch = originalFetch
  }
})

test('generated fetch responses preserve serialized response headers', async () => {
  const client = await importGeneratedClient()
  const originalFetch = globalThis.fetch

  try {
    globalThis.fetch = async () =>
      new Response(JSON.stringify({ status: 'ok' }), {
        status: 200,
        headers: {
          'content-type': 'application/json',
          etag: '"health-v1"',
          link: '</api/health>; rel="self"',
          'pagination-total-count': '1',
        },
      })

    const response = await client.getHealth()
    assert.equal(response.data.status, 'ok')
    assert.equal(response.headers.etag, '"health-v1"')
    assert.equal(response.headers.link, '</api/health>; rel="self"')
    assert.equal(response.headers['pagination-total-count'], '1')
  } finally {
    globalThis.fetch = originalFetch
  }
})
