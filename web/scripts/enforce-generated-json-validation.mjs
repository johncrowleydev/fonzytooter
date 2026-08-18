import fs from 'node:fs'
import path from 'node:path'

const endpointArgument = process.argv[3]
const generatedClientPath = path.resolve(endpointArgument ?? 'src/api/generated/endpoints.ts')
const openapiPath = path.resolve(
  endpointArgument === undefined
    ? '../openapi/openapi.json'
    : path.join(path.dirname(generatedClientPath), '../../../../openapi/openapi.json'),
)
const source = fs.readFileSync(generatedClientPath, 'utf8')
const openapi = JSON.parse(fs.readFileSync(openapiPath, 'utf8'))

const responseValidationPattern =
  /  const parsedBody = body \? \(contentType\.includes\('json'\) \? JSON\.parse\(body\) : body\) : \{\}\r?\n  const data = contentType\.includes\('json'\) \? ([A-Za-z_$][\w$]*)\.parse\(parsedBody\) : parsedBody/g

const generated = source.replace(
  responseValidationPattern,
  (_match, schemaName) => `  const data = ${schemaName}.parse(body ? JSON.parse(body) : {})`,
)

const contentTypeDeclaration =
  /^  const contentType = \(res\.headers\.get\('content-type'\) \?\? ''\)\.toLowerCase\(\)\r?\n/gm
const normalized = generated.replace(contentTypeDeclaration, '')

if (normalized.includes('contentType')) {
  throw new Error('Generated JSON transport still contains a Content-Type-gated branch')
}

const escapeRegExp = (value) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
const jsonOperations = Object.values(openapi.paths ?? {}).flatMap((pathItem) =>
  Object.values(pathItem ?? {}).filter(
    (operation) =>
      operation &&
      typeof operation.operationId === 'string' &&
      Object.keys(operation.responses ?? {}).some(
        (status) =>
          /^2\d\d$/.test(status) && operation.responses[status]?.content?.['application/json'],
      ),
  ),
)

for (const operation of jsonOperations) {
  const operationPattern = new RegExp(
    `export const ${escapeRegExp(operation.operationId)} = async[\\s\\S]*?(?=\\nexport const |\\nexport type |$)`,
  )
  const functionBody = normalized.match(operationPattern)?.[0]
  if (!functionBody) {
    throw new Error(`Could not find generated function for ${operation.operationId}`)
  }
  if (!/\.parse\(body \? JSON\.parse\(body\) : \{\}\)/.test(functionBody)) {
    throw new Error(
      `Generated JSON function ${operation.operationId} does not validate its success body`,
    )
  }
}

if (normalized !== source) {
  fs.writeFileSync(generatedClientPath, normalized)
}
