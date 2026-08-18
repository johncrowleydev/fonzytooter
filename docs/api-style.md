# API style

Fonzytooter's HTTP API should be boring, predictable, and resource-oriented.

The API is not an RPC façade over Go methods. URLs identify resources and collections; HTTP methods express what happens to those resources. Similar operations should look and behave similarly across the entire API.

This document defines the style contract. `docs/api-contract.md` defines how that contract is generated into OpenAPI and the frontend client.

## Core principles

1. **Model resources, not actions.** Endpoint paths name resources and collections. Avoid RPC/action URLs such as `/completeLesson`, `/reviews/{id}/submit`, or `/projects/{id}/archive` when the same state transition can be represented as creation, replacement, partial update, or deletion of a resource.
2. **Use HTTP methods according to their semantics.** `GET`, `POST`, `PUT`, `PATCH`, and `DELETE` are not interchangeable routing verbs.
3. **Keep requests stateless.** Every request contains the information necessary to interpret it. Do not make API semantics depend on hidden conversational/session state from earlier requests.
4. **Preserve idempotency where HTTP defines or naturally supports it.** Retries of `GET`, `PUT`, and `DELETE` must not create additional effects. Design `PATCH` operations to be idempotent whenever practical.
5. **Keep resource bodies resource-shaped.** Do not wrap ordinary resources in generic envelopes merely to carry pagination, counts, links, status strings, or other transport metadata.
6. **Use headers for response metadata.** Pagination/navigation metadata and similar HTTP-level metadata should live in headers so collection bodies can remain arrays of resources.
7. **Consistency is part of the API contract.** Analogous endpoints use the same naming patterns, status codes, error semantics, pagination conventions, and representation conventions.
8. **Prefer extending an existing resource operation over inventing another endpoint.** A small filter, projection, relationship, or optional field is usually preferable to a new one-off endpoint when the existing resource remains the thing being represented.
9. **Keep the endpoint surface small.** New endpoints require a resource-level reason to exist, not merely a convenient backend method boundary.

## Resource paths

Use plural resource nouns for collections and stable IDs for individual resources.

Prefer:

```text
GET    /api/modules
GET    /api/modules/{moduleId}
GET    /api/objectives/{objectiveId}
GET    /api/projects
GET    /api/projects/{projectId}
```

Nested paths are appropriate when they express a real containment or subordinate-resource relationship:

```text
GET  /api/modules/{moduleId}/lessons
GET  /api/objectives/{objectiveId}/reviews
POST /api/exercises/{exerciseId}/attempts
```

Do not nest merely because two Go types refer to each other. Keep nesting shallow enough that the primary resource remains obvious.

### Naming

Path segments use lowercase plural nouns. Use kebab-case only when a resource name genuinely contains multiple words.

Prefer:

```text
/api/review-items
/api/exercise-attempts
```

not:

```text
/api/reviewItems
/api/ReviewItems
/api/get-review-items
```

Path parameter names should be explicit and consistent:

```text
{moduleId}
{objectiveId}
{exerciseId}
```

Use `id` as the resource identifier field in JSON representations unless the representation contains multiple IDs and a more specific name is needed for clarity.

JSON property names use `camelCase` consistently.

## HTTP methods

### GET

`GET` retrieves a resource or collection.

It must be safe and must not intentionally change persistent application state.

Do not use `GET` to trigger commands, evaluations, submissions, or other mutations.

Examples:

```text
GET /api/projects
GET /api/projects/{projectId}
```

### POST

`POST` creates a new subordinate resource in a collection when the server is responsible for assigning its identity.

Example:

```text
POST /api/exercises/{exerciseId}/attempts
```

The request body represents the resource being created or the client-controlled fields of that resource. Do not wrap it in an action object such as `{ "action": "submit", ... }`.

A successful creation returns `201 Created` consistently, includes a `Location` header for the new resource, and returns the created resource representation unless there is a specific documented reason not to.

Do not use `POST` as a generic "do something" verb when `PUT`, `PATCH`, `DELETE`, or creation of a meaningful subordinate resource describes the operation.

### PUT

`PUT` replaces the representation of a resource at a known URI and is idempotent.

Use it when full replacement semantics are useful and the client knows the target resource URI.

A successful `PUT` of an existing resource returns `200 OK` with the resulting representation.

If the API later intentionally supports client-addressed creation through `PUT`, that case must be documented consistently and return `201 Created`; do not allow individual endpoints to invent different conventions ad hoc.

### PATCH

`PATCH` partially updates an existing resource.

Use it for small changes to resource state when sending a complete replacement would be unnatural.

Patch bodies describe resource fields, not commands.

Prefer:

```http
PATCH /api/projects/{projectId}

{
  "status": "completed"
}
```

over:

```text
POST /api/projects/{projectId}/complete
POST /api/completeProject
```

Design patch operations to be idempotent when practical: setting `status` to `completed` twice should leave the same state rather than produce a second side effect.

A successful `PATCH` returns `200 OK` with the resulting resource representation.

Do not alternate between `200` and `204` for analogous update endpoints based on handler preference.

### DELETE

`DELETE` removes a resource or relationship and is idempotent in effect.

A successful deletion returns `204 No Content` consistently.

Retrying the same deletion must not create additional side effects. Whether deleting an already-absent resource returns `204` or `404` should be chosen once for a resource class and used consistently; default to `404` when the addressed resource does not exist unless idempotent external retry semantics provide a concrete reason to choose otherwise.

## Avoid action and RPC endpoints

Do not expose backend methods as routes.

Avoid patterns such as:

```text
POST /api/startReview
POST /api/reviews/{id}/answer
POST /api/projects/{id}/markComplete
POST /api/exercises/{id}/runCheck
GET  /api/getCurrentProgress
```

First ask what resource exists or changes.

Examples of more resource-oriented modeling:

```text
POST  /api/review-sessions
POST  /api/review-sessions/{sessionId}/responses
PATCH /api/projects/{projectId}
POST  /api/exercises/{exerciseId}/attempts
GET   /api/progress
```

A process can still be modeled RESTfully when it has a meaningful resource representation. A review session, exercise attempt, tutor turn, note, bookmark, or project status is preferable to an arbitrary command endpoint when that concept actually exists in the domain.

Do not invent a resource noun solely to disguise RPC. If a proposed endpoint feels like an action with `-request`, `-command`, or another artificial noun appended, reconsider the model.

## Before adding a new endpoint

A new route should survive this decision test:

1. What resource or collection does this URI identify?
2. Can the operation be represented by `GET`, `POST` creation, `PUT`, `PATCH`, or `DELETE` on an existing resource?
3. Could a small optional query parameter, filter, field, relationship, or representation extension solve the requirement without another route?
4. Is the proposed endpoint merely exposing a convenient Go service method?
5. Would another analogous feature reasonably use the same endpoint shape and status semantics?
6. Does adding this route make the API easier to understand as a resource model, or merely easier to implement in one handler?

Prefer extending the existing resource operation when the distinction is small and remains conceptually the same resource.

Do not overload an endpoint until it has ambiguous semantics merely to minimize route count. The goal is a small coherent resource surface, not the fewest possible strings in the router.

## Request and response representations

### Resource bodies stay pure

A single-resource response body is the resource representation itself:

```json
{
  "id": "nn.backpropagation",
  "title": "Implement gradient descent and backpropagation",
  "introduced": true
}
```

A collection response body is an array of resource representations:

```json
[
  {
    "id": "nn.neuron",
    "title": "Explain a neuron as a parameterized function"
  },
  {
    "id": "nn.backpropagation",
    "title": "Implement gradient descent and backpropagation"
  }
]
```

Do not introduce generic envelopes like:

```json
{
  "data": [...],
  "meta": {
    "total": 42,
    "page": 2
  }
}
```

unless a future protocol requirement makes an envelope materially necessary.

### Request bodies are resource-shaped

Create/update bodies contain the resource fields that the client is allowed to supply.

Avoid generic wrapper fields such as:

```json
{
  "payload": {...},
  "request": {...},
  "resource": {...}
}
```

unless the wrapper itself is a meaningful domain structure.

Transport concerns such as trace IDs, pagination metadata, and cache validators do not belong in the resource body.

## Collection filtering, sorting, and pagination

Collection selection belongs in query parameters.

Examples:

```text
GET /api/objectives?moduleId=neural-networks
GET /api/projects?status=in-progress
GET /api/activity?limit=50
```

Use the same parameter names and semantics everywhere they apply. Do not use `pageSize` in one collection, `take` in another, and `limit` in a third.

### Pagination inputs

Default to `limit` plus a cursor when stable cursor pagination is useful:

```text
?limit=50&cursor=...
```

Use `offset` only where offset pagination is materially simpler and its tradeoffs are acceptable. Do not mix pagination models arbitrarily across analogous collections.

### Pagination response metadata

Keep the body as the resource collection. Put pagination metadata in response headers.

Prefer standard HTTP mechanisms where they fit:

- `Link` for navigational links such as `next` and `prev`;
- a consistently named total-count header when the total is meaningful and inexpensive to compute.

If a custom total-count header is needed, use:

```text
Pagination-Total-Count
```

Do not invent a family of endpoint-specific `X-*` headers.

Example:

```http
HTTP/1.1 200 OK
Content-Type: application/json
Pagination-Total-Count: 143
Link: </api/activity?limit=50&cursor=abc>; rel="next"

[
  ... resources only ...
]
```

The OpenAPI operation must describe relevant response headers so the generated contract includes them.

## Status codes

Status-code consistency matters more than handler-local preference.

Use these defaults unless a genuinely different HTTP semantic applies:

| Situation | Status |
| --- | --- |
| Successful resource/collection `GET` | `200 OK` |
| Successful `POST` creation | `201 Created` |
| Successful existing-resource `PUT` | `200 OK` |
| Successful `PATCH` | `200 OK` |
| Successful `DELETE` | `204 No Content` |
| Syntactically invalid request | `400 Bad Request` |
| Missing/invalid authentication | `401 Unauthorized` |
| Authenticated but forbidden | `403 Forbidden` |
| Resource does not exist | `404 Not Found` |
| Resource/state conflict | `409 Conflict` |
| Semantically invalid well-formed input | `422 Unprocessable Content` |
| Unexpected server failure | `500 Internal Server Error` |
| Upstream dependency failure when the distinction matters | `502 Bad Gateway` |
| Service temporarily unavailable | `503 Service Unavailable` |

Do not return `201` from one create endpoint and `200` from another analogous create endpoint because different handlers happened to be written by different agents.

Do not return `200` with an error object in the body.

### Empty and missing resources

An empty collection is a successful collection result:

```http
200 OK
[]
```

It is not `404`.

A missing individual resource is `404`.

## Errors

Errors should have one consistent resource-independent representation rather than endpoint-specific string/object shapes.

The exact Huma/OpenAPI error schema should be selected during implementation and then used everywhere.

At minimum, clients need stable machine-readable information plus a human-readable message. Validation errors should identify the relevant field/location when possible.

Do not expose Go internals, stack traces, SQL messages, or upstream secrets through error bodies.

Once the error representation is selected, generated Zod schemas and frontend handling must use that same contract rather than redefining errors per feature.

## Idempotency and retries

Respect standard HTTP semantics:

- `GET` is safe and idempotent;
- `PUT` is idempotent;
- `DELETE` is idempotent in effect;
- `PATCH` should be designed to be idempotent whenever practical;
- `POST` creation is not inherently idempotent.

Do not put non-idempotent side effects behind an ostensibly idempotent method.

If a future `POST` operation is important enough that network retries could accidentally create duplicates, add an explicit idempotency mechanism such as an `Idempotency-Key` header rather than pretending repeated `POST`s are inherently safe.

Idempotency keys are not required preemptively for every endpoint.

## Statelessness

The API is stateless at the HTTP interaction level.

Each request must carry the identifiers, authorization context, payload, and conditional metadata needed to process that request.

Server-side persistent learner state is of course allowed; statelessness does not mean the application has no database. It means the interpretation of request N does not depend on an undocumented in-memory conversational state created by request N-1.

The tutor conversation itself may be a resource with persisted history. That is still resource state, not hidden transport session state.

## Relationships

Represent relationships consistently.

If a resource naturally contains identifiers for related resources, use explicit IDs rather than polymorphic `type/id` pairs or opaque blobs.

Use nested collection routes when the relationship itself is the useful resource view:

```text
GET /api/modules/{moduleId}/objectives
```

Do not create both nested and top-level variants unless each serves a clear purpose. Often a top-level filtered collection is sufficient:

```text
GET /api/objectives?moduleId={moduleId}
```

Choose one convention for analogous relationships and reuse it.

## Conditional requests and concurrency

When concurrent editing becomes relevant, prefer HTTP-native conditional request mechanisms rather than inventing action/version endpoints.

Potential tools include:

- `ETag` response headers;
- `If-Match` for optimistic concurrency;
- `If-None-Match` for cache validation.

Do not add these until real concurrent-change behavior requires them, but preserve room for them by keeping transport metadata out of resource bodies.

## OpenAPI requirements

The generated OpenAPI contract must reflect the REST semantics rather than merely documenting whatever handlers happen to exist.

Each operation should define:

- stable `operationId`;
- resource-oriented path;
- correct HTTP method;
- expected success status code;
- request schema where applicable;
- response resource schema;
- documented response headers such as `Location`, `Link`, or pagination count where applicable;
- consistent error schemas;
- query parameters for filtering/sorting/pagination.

Operation IDs are code-generation identifiers, not an excuse to make URLs RPC-like. An operation ID such as `createExerciseAttempt` is fine while the route remains:

```text
POST /api/exercises/{exerciseId}/attempts
```

## Exceptions

Not every HTTP interaction is ordinary CRUD.

The tutor SSE stream is an intentional streaming transport. Its path and request/event shapes should still be resource-oriented and generated through the OpenAPI contract, but incremental event delivery means some ordinary response/body conventions do not apply.

Other exceptions require a concrete protocol reason. Do not treat "this was easier to code as an action endpoint" as an exception.

When an exception is necessary:

1. keep it narrow;
2. document why the normal resource model is insufficient;
3. preserve naming, status, validation, and OpenAPI consistency wherever they still apply;
4. do not generalize one exception into a second API style.

## Review checklist

Before adding or changing an API operation, verify:

- the URI identifies a resource or collection rather than an action;
- collection names are plural and naming conventions match existing routes;
- the chosen HTTP method matches the operation semantics;
- `GET` is safe;
- idempotent methods remain idempotent;
- create/update/delete status codes follow the shared conventions;
- request and response bodies are resource-shaped rather than generic envelopes;
- pagination/filter inputs use established query parameters;
- pagination/navigation metadata is represented in headers rather than body wrappers;
- errors use the common error schema;
- the requirement could not be more cleanly satisfied by a small extension to an existing resource operation;
- the endpoint participates in the generated OpenAPI contract described by `docs/api-contract.md`;
- generated frontend code and Zod validation remain the only application API boundary.
