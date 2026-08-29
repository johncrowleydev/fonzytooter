import { defineConfig } from 'orval'

export default defineConfig({
  helixAcademy: {
    input: {
      target: '../openapi/openapi.json',
      filters: {
        includeUnreferencedSchemas: true,
        mode: 'exclude',
        tags: ['tutor'],
      },
    },
    output: {
      clean: true,
      client: 'react-query',
      formatter: 'prettier',
      httpClient: 'fetch',
      target: './src/api/generated/endpoints.ts',
      schemas: {
        path: './src/api/generated/schemas',
        type: 'zod',
      },
      override: {
        fetch: {
          forceSuccessResponse: true,
          includeHttpResponseReturnType: true,
          runtimeValidation: true,
          serializeResponseHeaders: true,
        },
        query: {
          runtimeValidation: true,
        },
        zod: {
          generateReusableSchemas: true,
          strict: {
            body: true,
            response: true,
          },
          variant: 'classic',
          version: 4,
        },
      },
    },
    // Orval 8.24.0 checks Content-Type before its Zod parse even when the
    // OpenAPI success representation is JSON. The api:generate script runs a
    // deterministic normalizer after Orval has finished writing every file.
  },
})
