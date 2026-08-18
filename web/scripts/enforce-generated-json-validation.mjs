import fs from 'node:fs'
import path from 'node:path'

const generatedClientPath = path.resolve('src/api/generated/endpoints.ts')
const source = fs.readFileSync(generatedClientPath, 'utf8')

const responseValidationPattern =
  /  const parsedBody = body \? \(contentType\.includes\('json'\) \? JSON\.parse\(body\) : body\) : \{\}\r?\n  const data = contentType\.includes\('json'\) \? ([A-Za-z_$][\w$]*)\.parse\(parsedBody\) : parsedBody/g

const generated = source.replace(
  responseValidationPattern,
  (_match, schemaName) => `  const data = ${schemaName}.parse(body ? JSON.parse(body) : {})`,
)

const contentTypeDeclaration =
  /  const contentType = \(res\.headers\.get\('content-type'\) \?\? ''\)\.toLowerCase\(\)\r*\n/g
const withoutUnusedContentType = generated.replace(contentTypeDeclaration, '')
if (withoutUnusedContentType === source) {
  throw new Error('Expected an Orval JSON response-validation block in endpoints.ts')
}
fs.writeFileSync(generatedClientPath, withoutUnusedContentType)
