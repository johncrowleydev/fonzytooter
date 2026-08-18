import { defineConfig } from 'orval'

export default defineConfig({
  fonzytooter: {
    input: {
      target: '../openapi/openapi.json',
    },
    output: {
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
