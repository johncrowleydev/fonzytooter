import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { compile } from '@mdx-js/mdx'

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url))
const curriculumModulesDirectory = path.resolve(scriptDirectory, '../../curriculum/modules')

async function collectMdxFiles(directory) {
  const entries = await fs.readdir(directory, { withFileTypes: true })
  const files = []

  for (const entry of entries) {
    const entryPath = path.join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...(await collectMdxFiles(entryPath)))
    } else if (entry.isFile() && entry.name.endsWith('.mdx')) {
      files.push(entryPath)
    }
  }

  return files.sort()
}

function stripFrontmatter(source) {
  const lines = source.split(/\r?\n/)
  if (lines[0]?.replace(/^\uFEFF/, '') !== '---') return source

  const closingDelimiter = lines.findIndex((line, index) => index > 0 && line === '---')
  return closingDelimiter === -1 ? source : lines.slice(closingDelimiter + 1).join('\n')
}

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error)
}

const files = await collectMdxFiles(curriculumModulesDirectory)
const failures = []

for (const filePath of files) {
  try {
    const source = await fs.readFile(filePath, 'utf8')
    await compile(stripFrontmatter(source), { jsxRuntime: 'automatic' })
  } catch (error) {
    failures.push(
      `${path.relative(path.resolve(scriptDirectory, '..'), filePath)}: ${errorMessage(error)}`,
    )
  }
}

if (failures.length > 0) {
  console.error('Curriculum MDX validation failed:')
  for (const failure of failures) console.error(`- ${failure}`)
  process.exitCode = 1
} else {
  console.log(`Curriculum MDX validation passed: ${files.length} files`)
}
