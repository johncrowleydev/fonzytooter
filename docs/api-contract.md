# API contract and full-stack type safety

Fonzytooter treats the Go API contract as generated infrastructure, not as a set of shapes that the frontend manually re-describes.

The intended flow is:

```text
Go operations + Go request/response structs
        │
        ▼
Huma v2
        │
        ▼
OpenAPI 3.1
openapi/openapi.json
        │
        ▼
Orval v8
        │
        ├── TanStack Query client/hooks
        └── Zod 4 schemas
                 │
                 ▼
           z.infer / z.input / z.output
```

The important property is **one contract flowing in one direction**. Do not hand-maintain parallel Go DTOs, OpenAPI YAML, TypeScript DTOs, React Query hooks, and Zod schemas that can drift independently.

## Source of truth

For HTTP APIs, the backend Go code is authoritative.

Use **Huma v2** to register operations and derive OpenAPI 3.1 plus JSON Schema from the Go request/response types. Huma can wrap the standard-library `http.ServeMux`, so adopting it does not require replacing the current `net/http` foundation with a larger router/framework.

Each operation should have:

- a stable, intentional `operationId`;
- explicit request and response structs;
- explicit path/query/header/body fields;
- documented success responses;
- documented error responses when they are part of the API contract;
- tags/descriptions/examples when they materially improve the generated contract.

Do not add a route by registering an ad-hoc `http.HandlerFunc` and then separately documenting it by hand. An HTTP endpoint that is part of the application API should be registered through the OpenAPI-producing API layer.

## OpenAPI artifact

The repository should contain a generated canonical spec at:

```text
openapi/openapi.json
```

OpenAPI 3.1 is the canonical format.

Generation must be possible **without starting a listening server**. The planned shape is a small Go command such as:

```text
server/cmd/openapi
```

which constructs the same API registration used by the real server and writes the generated spec to stdout. Conceptually:

```bash
cd server
go run ./cmd/openapi > ../openapi/openapi.json
```

The exact command may change during implementation, but these requirements do not:

1. server startup and spec generation must share the same operation registration;
2. the generated spec must be deterministic;
3. CI must be able to regenerate it from source;
4. the committed spec must not be manually edited.

## Frontend generation

Use **Orval v8** to consume `openapi/openapi.json` and generate the frontend API surface.

The target configuration is:

- `client: 'react-query'`;
- `httpClient: 'fetch'`;
- TanStack Query generated hooks/functions for ordinary JSON endpoints;
- Zod-backed generated schemas via `schemas: { type: 'zod' }`;
- Zod generation pinned to major version 4 rather than auto-detected;
- generated reusable schemas for OpenAPI component schemas;
- generated output isolated under `web/src/api/generated/`;
- runtime validation enabled for generated JSON responses through the generated fetch transport.

A representative configuration shape is:

```ts
import { defineConfig } from 'orval'

export default defineConfig({
  fonzytooter: {
    input: {
      target: '../openapi/openapi.json',
    },
    output: {
      client: 'react-query',
      httpClient: 'fetch',
      target: './src/api/generated/endpoints.ts',
      schemas: {
        path: './src/api/generated/schemas',
        type: 'zod',
      },
      override: {
        fetch: {
          runtimeValidation: true,
        },
        query: {
          // React Query's fetch adapter also needs this to retain the runtime
          // schema import when the generated fetch function calls Schema.parse.
          runtimeValidation: true,
        },
        zod: {
          version: 4,
          variant: 'classic',
          generateReusableSchemas: true,
        },
      },
    },
  },
})
```

Treat this as the architectural target, not a promise that every option will survive implementation unchanged. The first implementation PR must inspect the generated output and prove that response validation is actually occurring. If the generator requires a small configuration adjustment, preserve the guarantees in this document rather than preserving example syntax blindly.

## Zod is the frontend runtime contract

Generated Zod schemas are authoritative for API-shaped values in the frontend.

Use them for both:

1. **runtime validation** at trust boundaries;
2. **static TypeScript types** through schema inference.

Prefer:

```ts
import { z } from 'zod'
import { UserSchema } from '@/api/generated/schemas'

export type User = z.infer<typeof UserSchema>
```

or, when input/output differ:

```ts
export type CreateUserInput = z.input<typeof CreateUserSchema>
export type CreateUserOutput = z.output<typeof CreateUserSchema>
```

Do **not** write a second interface/type that merely copies a generated schema:

```ts
// Wrong: parallel type definition that can drift.
interface User {
  id: string
  name: string
}
```

Frontend-only view models and domain types are still allowed when they represent a genuinely different concept or transformed shape. The forbidden case is a handwritten TypeScript declaration whose purpose is simply to mirror an API request/response schema that already exists in generated Zod.

## Request validation

Static typing alone does not validate runtime input.

Values originating from untrusted or weakly typed boundaries should be parsed with the generated request schema before being passed into a generated mutation/client call. Examples include:

- form payloads assembled from uncontrolled input;
- `localStorage`;
- URL/query-string values;
- imported JSON;
- data returned by browser/platform APIs typed as `unknown`;
- any manually decoded external payload.

Use `.parse()` when invalid input is exceptional and should fail immediately, or `.safeParse()` when the UI should surface validation feedback.

Do not manually duplicate validation rules that already exist in generated Zod.

## Response validation

Successful JSON responses from ordinary API endpoints must be runtime-validated against generated Zod schemas before application feature code consumes them.

The preferred implementation is Orval's generated fetch transport with Zod runtime validation enabled, so TanStack Query receives already-validated values.

A successful HTTP status with a payload that violates the OpenAPI/Zod contract is an application error, not a value that should quietly flow through because TypeScript asserted a type at compile time.

The first implementation must include at least one test that supplies an invalid mock response and proves the generated/runtime boundary rejects it with a Zod validation error.

## React Query is the normal server-state boundary

Ordinary JSON API calls from application code must use the **generated TanStack Query client/hooks**.

Feature code must not:

- call `fetch()` directly for application API endpoints;
- import or use Axios for application API endpoints;
- construct ad-hoc `/api/...` requests;
- manually recreate query keys that the generated client already owns;
- write a second handwritten hook around an endpoint merely to reproduce generated request behavior.

Feature-level hooks are allowed when they add real application semantics, combine multiple generated operations, select/transform data, or coordinate UI behavior. They should compose the generated API surface rather than replace it.

## Streaming exception: tutor SSE

The tutor stream is intentionally different from ordinary JSON request/response traffic.

A streamed POST response is not naturally modeled as a normal TanStack Query result because the consumer needs incremental events before the response completes. OpenAPI describes the endpoint contract, but it does not make React Query the correct incremental-stream transport.

Therefore the tutor stream may use one narrow infrastructure adapter under:

```text
web/src/api/runtime/
```

Rules for this exception:

- feature/components still must not call raw `fetch`;
- there should be one central streaming transport implementation, not per-feature fetch calls;
- the request body must use the generated request Zod schema/type;
- every decoded SSE event must be validated with generated Zod before entering application state;
- endpoint paths and event shapes must come from the OpenAPI/generated contract rather than duplicated handwritten DTOs;
- the exception must remain specific to transports that genuinely cannot use the generated React Query client.

On the Go side, prefer Huma's typed SSE registration/support so the stream's event contract remains part of the same API/schema system.

Do not generalize this exception into a parallel handwritten API client.

## Generated code policy

Generated artifacts are source-controlled so that contract changes are visible in review and frontend code does not require generation merely to compile after checkout.

Generated files must carry an obvious generated-code header where supported and must never be manually edited.

Expected generated areas:

```text
openapi/openapi.json
web/src/api/generated/
```

Human-authored API infrastructure belongs outside the generated directory, for example:

```text
web/src/api/runtime/
web/orval.config.ts
web/scripts/check-api-boundaries.mjs
```

## Deterministic enforcement

These rules are not an agent honor system. CI must enforce the contract mechanically.

### 1. Generated-artifact drift check

CI should regenerate both layers:

```text
Go API registration
    -> openapi/openapi.json
    -> Orval output
```

and then fail if Git contains a diff in generated artifacts.

Conceptually:

```bash
git diff --exit-code -- openapi/openapi.json web/src/api/generated
```

This catches:

- a Go endpoint changed without regenerating OpenAPI;
- OpenAPI changed without regenerating the frontend client;
- generated files edited by hand;
- stale request/response/Zod types.

### 2. Frontend API-boundary checker

Add a deterministic static checker, preferably using the TypeScript compiler API already present in the frontend toolchain rather than regex-only source inspection.

Outside approved generated/runtime infrastructure it should fail on at least:

- calls to global `fetch`;
- imports from `axios`;
- `XMLHttpRequest` usage for application APIs;
- hard-coded same-origin `/api/...` request URLs;
- direct `useQuery` / `useMutation` / `useInfiniteQuery` usage for server endpoint access when the generated client already supplies the operation hook;
- handwritten object-shaped API types inside human-authored `src/api/` modules.

`useQueryClient` and other cache-management APIs may still be used when application behavior requires them.

The checker should explicitly exclude generated code and allow only narrowly documented runtime transport exceptions such as the tutor SSE adapter.

### 3. No duplicate API DTO declarations

There is no useful semantic reason for a human-authored TypeScript DTO to mirror a generated Zod schema.

Enforce the boundary structurally:

- API wire contracts live in generated schemas;
- human-authored `src/api/` code may re-export generated types or define aliases using `z.infer`, `z.input`, or `z.output`;
- human-authored `src/api/` code must not declare object-shaped API interfaces/types;
- feature/domain types may exist only when they represent a different application concept, not a copied wire DTO.

Because TypeScript is structurally typed, a static checker cannot prove that no developer anywhere in the repository has written an accidentally identical shape. The deterministic guarantee comes from controlling the **network boundary**: only generated/validated API contracts are allowed to cross it.

### 4. CI integration

The existing Frontend and Server checks should eventually add contract validation rather than create a parallel optional workflow.

A reasonable final validation sequence is:

```text
Server
- gofmt check
- go vet ./...
- go test ./...
- go build ./...
- generate OpenAPI

Frontend
- npm ci
- API boundary check
- generate Orval client/Zod
- generated-artifact drift check
- prettier check
- TypeScript check
- build
```

Exact job placement can change if one small dedicated `Contracts` job proves clearer, but API drift checks must be required CI, not an optional developer command.

## Change workflow

When adding or changing an ordinary JSON endpoint:

1. change the Go operation/request/response types;
2. regenerate `openapi/openapi.json`;
3. regenerate the Orval TanStack Query client and Zod schemas;
4. use the generated client/schema from frontend code;
5. run the API-boundary checker;
6. run normal frontend/server CI.

Do not start by writing a TypeScript DTO or frontend request and then work backward to make the Go API match it.

## Why this direction

This gives Fonzytooter four complementary protections:

- Go compiler checks the server models and handlers;
- OpenAPI makes the cross-language contract explicit and reviewable;
- Zod validates actual runtime values rather than trusting TypeScript assertions;
- generated TanStack Query hooks make the normal frontend network path hard to misuse.

The result should be boring: change a Go contract, regenerate, and let deterministic tooling tell us everywhere the change matters.
