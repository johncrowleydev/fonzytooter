import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { test } from 'node:test'
import { pathToFileURL } from 'node:url'
import * as ts from 'typescript'

const webRoot = path.resolve(import.meta.dirname, '..')

async function importTypeScriptModule(relativePath) {
  const sourcePath = path.join(webRoot, relativePath)
  const source = fs.readFileSync(sourcePath, 'utf8')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
    fileName: sourcePath,
  }).outputText
  const tempDirectory = fs.mkdtempSync(path.join(os.tmpdir(), 'helix-academy-workbook-test-'))
  const outputPath = path.join(tempDirectory, 'module.mjs')
  fs.writeFileSync(outputPath, output)
  try {
    return await import(`${pathToFileURL(outputPath).href}?test=${Date.now()}`)
  } finally {
    fs.rmSync(tempDirectory, { force: true, recursive: true })
  }
}

test('module workbook availability requires at least one worksheet', async () => {
  const { hasWorkbook } = await importTypeScriptModule(
    'src/features/worksheets/workbookAvailability.ts',
  )
  assert.equal(hasWorkbook(0), false)
  assert.equal(hasWorkbook(1), true)
})

test('PDF downloads honor server filenames, fall back deterministically, and revoke object URLs', async () => {
  const { downloadPdf } = await importTypeScriptModule('src/features/worksheets/downloadPdf.ts')
  const originalDocument = globalThis.document
  const originalWindow = globalThis.window
  const originalCreateObjectURL = URL.createObjectURL
  const originalRevokeObjectURL = URL.revokeObjectURL
  const clicked = []
  const revoked = []

  try {
    globalThis.document = {
      createElement() {
        return {
          click() {
            clicked.push(this.download)
          },
          download: '',
          href: '',
        }
      },
    }
    globalThis.window = { setTimeout(callback) { callback(); return 1 } }
    URL.createObjectURL = () => 'blob:test'
    URL.revokeObjectURL = (url) => revoked.push(url)

    downloadPdf(new Blob(['%PDF-test']), 'attachment; filename="server-workbook.pdf"', 'fallback.pdf')
    downloadPdf(new Blob(['%PDF-test']), undefined, 'fallback.pdf')

    assert.deepEqual(clicked, ['server-workbook.pdf', 'fallback.pdf'])
    assert.deepEqual(revoked, ['blob:test', 'blob:test'])
  } finally {
    globalThis.document = originalDocument
    globalThis.window = originalWindow
    URL.createObjectURL = originalCreateObjectURL
    URL.revokeObjectURL = originalRevokeObjectURL
  }
})
