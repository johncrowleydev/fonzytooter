import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import * as ts from 'typescript'

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
const TANSTACK_QUERY_MODULE = '@tanstack/react-query'

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

function isRawFetchExpression(expression) {
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
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node))
    return stringStartsWithApi(node.text)
  return ts.isTemplateExpression(node) && stringStartsWithApi(node.head.text)
}

function isRequestUrlArgument(node) {
  const parent = node.parent
  if (ts.isCallExpression(parent) && parent.arguments.some((argument) => argument === node)) {
    return isRawFetchExpression(parent.expression)
  }
  if (ts.isNewExpression(parent) && parent.arguments?.some((argument) => argument === node)) {
    return isRequestConstructor(parent.expression)
  }
  if (ts.isVariableDeclaration(parent) && parent.initializer === node) {
    const variableName = parent.name.getText().toLowerCase()
    return (
      variableName.includes('url') ||
      variableName.includes('path') ||
      variableName.includes('endpoint')
    )
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
  return (
    node.typeName.left.getText() === 'z' &&
    ['infer', 'input', 'output'].includes(node.typeName.right.text)
  )
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

function importDeclarationFor(node) {
  if (ts.isImportSpecifier(node) || ts.isNamespaceImport(node) || ts.isImportClause(node)) {
    const importClause = ts.isImportClause(node) ? node : node.parent.parent
    return ts.isImportDeclaration(importClause.parent) ? importClause.parent : undefined
  }
  if (ts.isExportSpecifier(node) && ts.isExportDeclaration(node.parent.parent)) {
    return node.parent.parent
  }
  return undefined
}

function isTanStackImportDeclaration(node) {
  return node?.moduleSpecifier && ts.isStringLiteral(node.moduleSpecifier)
    ? node.moduleSpecifier.text === TANSTACK_QUERY_MODULE
    : false
}

function symbolDeclarations(symbol) {
  return symbol?.declarations ?? []
}

function isTanStackBinding(symbol, checker, seen = new Set()) {
  if (!symbol || seen.has(symbol)) return false
  seen.add(symbol)

  if (
    symbolDeclarations(symbol).some((declaration) => {
      const importDeclaration = importDeclarationFor(declaration)
      return importDeclaration && isTanStackImportDeclaration(importDeclaration)
    })
  ) {
    return true
  }

  if (symbol.flags & ts.SymbolFlags.Alias) {
    return isTanStackBinding(checker.getAliasedSymbol(symbol), checker, seen)
  }

  return false
}

function isTanStackNamespaceBinding(symbol, checker, seen = new Set()) {
  if (!symbol || seen.has(symbol)) return false
  seen.add(symbol)

  if (
    symbolDeclarations(symbol).some((declaration) => {
      return (
        ts.isNamespaceImport(declaration) && isTanStackImportDeclaration(declaration.parent.parent)
      )
    })
  ) {
    return true
  }

  if (symbol.flags & ts.SymbolFlags.Alias) {
    return isTanStackNamespaceBinding(checker.getAliasedSymbol(symbol), checker, seen)
  }

  return false
}

function isTanStackHookIdentifier(node, checker) {
  return (
    ts.isIdentifier(node) &&
    BANNED_QUERY_HOOKS.has(node.text) &&
    isTanStackBinding(checker.getSymbolAtLocation(node), checker)
  )
}

function isTanStackNamespaceHookAccess(node, checker) {
  return (
    ts.isPropertyAccessExpression(node) &&
    BANNED_QUERY_HOOKS.has(node.name.text) &&
    isTanStackNamespaceBinding(checker.getSymbolAtLocation(node.expression), checker)
  )
}

export function findApiBoundaryViolations(sourceRoot) {
  const root = path.resolve(sourceRoot)
  const violations = []
  const sourceFiles = collectSourceFiles(root)
  const program = ts.createProgram(sourceFiles, {
    allowJs: false,
    isolatedModules: true,
    jsx: ts.JsxEmit.ReactJSX,
    module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.NodeJs,
    noEmit: true,
    skipLibCheck: true,
    target: ts.ScriptTarget.ES2022,
  })
  const checker = program.getTypeChecker()

  for (const fileName of sourceFiles) {
    const relativePath = path.relative(root, fileName)
    if (isGeneratedPath(relativePath)) continue
    const rawFetchAllowed = isRuntimePath(relativePath)
    const sourceFile = program.getSourceFile(fileName)
    if (!sourceFile) continue

    function visit(node) {
      if (!rawFetchAllowed && ts.isCallExpression(node) && isRawFetchExpression(node.expression)) {
        violations.push(
          diagnostic(
            sourceFile,
            node,
            'raw global fetch is not allowed outside generated or runtime API infrastructure',
          ),
        )
      }

      if (ts.isNewExpression(node) && expressionName(node.expression) === 'XMLHttpRequest') {
        violations.push(
          diagnostic(sourceFile, node, 'XMLHttpRequest is not allowed for application API access'),
        )
      }

      if (ts.isPropertyAccessExpression(node) && node.name.text === 'XMLHttpRequest') {
        violations.push(
          diagnostic(sourceFile, node, 'XMLHttpRequest is not allowed for application API access'),
        )
      }

      if (isAxiosRequire(node)) {
        violations.push(
          diagnostic(sourceFile, node, 'axios is not allowed in human-authored frontend code'),
        )
      }

      if (
        (ts.isStringLiteral(node) ||
          ts.isNoSubstitutionTemplateLiteral(node) ||
          ts.isTemplateExpression(node)) &&
        nodeStartsWithApi(node) &&
        isRequestUrlArgument(node)
      ) {
        violations.push(
          diagnostic(sourceFile, node, 'hard-coded /api URL is not allowed in feature code'),
        )
      }

      if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
        if (node.moduleSpecifier.text === 'axios') {
          violations.push(
            diagnostic(sourceFile, node, 'axios is not allowed in human-authored frontend code'),
          )
        }
        for (const element of node.importClause?.namedBindings &&
        ts.isNamedImports(node.importClause.namedBindings)
          ? node.importClause.namedBindings.elements
          : []) {
          const importedName = element.propertyName?.text ?? element.name.text
          if (
            BANNED_QUERY_HOOKS.has(importedName) &&
            isTanStackBinding(checker.getSymbolAtLocation(element.name), checker)
          ) {
            violations.push(
              diagnostic(
                sourceFile,
                element,
                `${importedName} is endpoint-level React Query usage; use the generated API operation instead`,
              ),
            )
          }
        }
      }

      if (isTanStackNamespaceHookAccess(node, checker)) {
        violations.push(
          diagnostic(
            sourceFile,
            node,
            `${node.name.text} is endpoint-level React Query usage; use the generated API operation instead`,
          ),
        )
      }

      if (ts.isCallExpression(node) && isTanStackHookIdentifier(node.expression, checker)) {
        violations.push(
          diagnostic(
            sourceFile,
            node,
            `${node.expression.text} is endpoint-level React Query usage; use the generated API operation instead`,
          ),
        )
      }

      if (
        relativePath.split(path.sep).join('/').startsWith('api/') &&
        !isGeneratedPath(relativePath)
      ) {
        if (ts.isInterfaceDeclaration(node)) {
          violations.push(
            diagnostic(
              sourceFile,
              node,
              'handwritten object-shaped API interfaces are not allowed; infer from generated Zod schemas',
            ),
          )
        }
        if (
          ts.isTypeAliasDeclaration(node) &&
          typeContainsObjectShape(node.type) &&
          !isZodInferenceAlias(node.type)
        ) {
          violations.push(
            diagnostic(
              sourceFile,
              node,
              'handwritten object-shaped API types are not allowed; infer from generated Zod schemas',
            ),
          )
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
    .map(
      (violation) => `${violation.file}:${violation.line}:${violation.column} ${violation.message}`,
    )
    .join('\n')
}

function isMainModule() {
  if (!process.argv[1]) return false
  return pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url
}

if (isMainModule()) {
  const rootArgumentIndex = process.argv.indexOf('--root')
  const sourceRoot =
    rootArgumentIndex === -1 ? path.resolve('src') : process.argv[rootArgumentIndex + 1]
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
