import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import * as ts from 'typescript-compiler-api'

const HUMAN_SOURCE_EXTENSIONS = new Set(['.ts', '.tsx', '.mts', '.cts'])
const BANNED_QUERY_HOOKS = new Set([
  'useInfiniteQuery',
  'useQueries',
  'useQuery',
  'useSuspenseInfiniteQuery',
  'useSuspenseQueries',
  'useSuspenseQuery',
  'useMutation',
])

function isGeneratedPath(relativePath) {
  return relativePath.split(path.sep).join('/').startsWith('api/generated/')
}

function isRuntimePath(relativePath) {
  return relativePath.split(path.sep).join('/').startsWith('api/runtime/')
}

function collectSourceFiles(directory) {
  const files = []
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const absolutePath = path.join(directory, entry.name)
    if (entry.isDirectory()) {
      files.push(...collectSourceFiles(absolutePath))
      continue
    }
    if (HUMAN_SOURCE_EXTENSIONS.has(path.extname(entry.name))) files.push(absolutePath)
  }
  return files.sort()
}

function diagnostic(sourceFile, node, message) {
  const { line, character } = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile))
  return {
    file: sourceFile.fileName,
    line: line + 1,
    column: character + 1,
    message,
  }
}

function expressionName(expression) {
  if (ts.isIdentifier(expression)) return expression.text
  if (ts.isPropertyAccessExpression(expression)) {
    const left = expressionName(expression.expression)
    return left ? `${left}.${expression.name.text}` : expression.name.text
  }
  return undefined
}

function isRequestExpression(expression) {
  const name = expressionName(expression)
  return name === 'fetch' || name === 'window.fetch' || name === 'globalThis.fetch'
}

function isRequestConstructor(expression) {
  const name = expressionName(expression)
  return name === 'Request' || name === 'URL'
}

function stringStartsWithApi(value) {
  return value.startsWith('/api/') || value === '/api'
}

function nodeStartsWithApi(node) {
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return stringStartsWithApi(node.text)
  return ts.isTemplateExpression(node) && stringStartsWithApi(node.head.text)
}

function isRequestUrlArgument(node) {
  const parent = node.parent
  if (ts.isCallExpression(parent) && parent.arguments.some((argument) => argument === node)) {
    return isRequestExpression(parent.expression)
  }
  if (ts.isNewExpression(parent) && parent.arguments?.some((argument) => argument === node)) {
    return isRequestConstructor(parent.expression)
  }
  if (ts.isVariableDeclaration(parent) && parent.initializer === node) {
    const variableName = parent.name.getText().toLowerCase()
    return variableName.includes('url') || variableName.includes('path') || variableName.includes('endpoint')
  }
  return false
}

function typeContainsObjectShape(node) {
  let found = false
  function visit(current) {
    if (found) return
    if (ts.isTypeLiteralNode(current)) {
      found = true
      return
    }
    ts.forEachChild(current, visit)
  }
  visit(node)
  return found
}

function isZodInferenceAlias(node) {
  if (!ts.isTypeReferenceNode(node)) return false
  if (!ts.isQualifiedName(node.typeName)) return false
  return node.typeName.left.getText() === 'z' && ['infer', 'input', 'output'].includes(node.typeName.right.text)
}

function isAxiosRequire(node) {
  return (
    ts.isCallExpression(node) &&
    ts.isIdentifier(node.expression) &&
    node.expression.text === 'require' &&
    node.arguments.length === 1 &&
    ts.isStringLiteral(node.arguments[0]) &&
    node.arguments[0].text === 'axios'
  )
}

export function findApiBoundaryViolations(sourceRoot) {
  const root = path.resolve(sourceRoot)
  const violations = []

  for (const fileName of collectSourceFiles(root)) {
    const relativePath = path.relative(root, fileName)
    if (isGeneratedPath(relativePath)) continue
    const runtime = isRuntimePath(relativePath)
    const sourceText = fs.readFileSync(fileName, 'utf8')
    const sourceFile = ts.createSourceFile(
      fileName,
      sourceText,
      ts.ScriptTarget.Latest,
      true,
      path.extname(fileName) === '.tsx' ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
    )

    function visit(node) {
      if (!runtime && ts.isCallExpression(node) && isRequestExpression(node.expression)) {
        violations.push(diagnostic(sourceFile, node, 'raw global fetch is not allowed outside generated or runtime API infrastructure'))
      }

      if (!runtime && ts.isNewExpression(node) && expressionName(node.expression) === 'XMLHttpRequest') {
        violations.push(diagnostic(sourceFile, node, 'XMLHttpRequest is not allowed for application API access'))
      }

      if (
        !runtime &&
        ts.isPropertyAccessExpression(node) &&
        node.name.text === 'XMLHttpRequest'
      ) {
        violations.push(diagnostic(sourceFile, node, 'XMLHttpRequest is not allowed for application API access'))
      }

      if (!runtime && isAxiosRequire(node)) {
        violations.push(diagnostic(sourceFile, node, 'axios is not allowed in human-authored frontend code'))
      }

      if (!runtime && (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node) || ts.isTemplateExpression(node))) {
        if (nodeStartsWithApi(node) && isRequestUrlArgument(node)) {
          violations.push(diagnostic(sourceFile, node, 'hard-coded /api URL is not allowed in feature code'))
        }
      }

      if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
        if (!runtime && node.moduleSpecifier.text === 'axios') {
          violations.push(diagnostic(sourceFile, node, 'axios is not allowed in human-authored frontend code'))
        }
        for (const element of node.importClause?.namedBindings && ts.isNamedImports(node.importClause.namedBindings)
          ? node.importClause.namedBindings.elements
          : []) {
          const importedName = element.propertyName?.text ?? element.name.text
          if (BANNED_QUERY_HOOKS.has(importedName)) {
            violations.push(diagnostic(sourceFile, element, `${importedName} is endpoint-level React Query usage; use the generated API operation instead`))
          }
        }
      }

      if (!runtime && ts.isPropertyAccessExpression(node) && BANNED_QUERY_HOOKS.has(node.name.text)) {
        violations.push(diagnostic(sourceFile, node, `${node.name.text} is endpoint-level React Query usage; use the generated API operation instead`))
      }

      if (!runtime && ts.isCallExpression(node) && ts.isIdentifier(node.expression) && BANNED_QUERY_HOOKS.has(node.expression.text)) {
        violations.push(diagnostic(sourceFile, node, `${node.expression.text} is endpoint-level React Query usage; use the generated API operation instead`))
      }

      if (relativePath.split(path.sep).join('/').startsWith('api/') && !isGeneratedPath(relativePath)) {
        if (ts.isInterfaceDeclaration(node)) {
          violations.push(diagnostic(sourceFile, node, 'handwritten object-shaped API interfaces are not allowed; infer from generated Zod schemas'))
        }
        if (ts.isTypeAliasDeclaration(node) && typeContainsObjectShape(node.type) && !isZodInferenceAlias(node.type)) {
          violations.push(diagnostic(sourceFile, node, 'handwritten object-shaped API types are not allowed; infer from generated Zod schemas'))
        }
      }

      ts.forEachChild(node, visit)
    }

    visit(sourceFile)
  }

  return violations
}

export function formatViolations(violations) {
  return violations
    .map((violation) => `${violation.file}:${violation.line}:${violation.column} ${violation.message}`)
    .join('\n')
}

function isMainModule() {
  if (!process.argv[1]) return false
  return pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url
}

if (isMainModule()) {
  const rootArgumentIndex = process.argv.indexOf('--root')
  const sourceRoot = rootArgumentIndex === -1 ? path.resolve('src') : process.argv[rootArgumentIndex + 1]
  if (!sourceRoot) {
    console.error('Missing value for --root')
    process.exit(2)
  }

  const violations = findApiBoundaryViolations(sourceRoot)
  if (violations.length > 0) {
    console.error(formatViolations(violations))
    process.exit(1)
  }
  console.log(`API boundary check passed: ${path.resolve(sourceRoot)}`)
}
