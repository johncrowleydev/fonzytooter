import { defineConfig } from 'orval'

export default defineConfig({
  fonzytooter: {
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
          variant: 'classic',
          version: 4,
        },
      },
    },
  },
})
