# Architecture

## Design target

Fonzytooter is a single-user personal learning application that may host multiple authored courses. It is not a multi-tenant LMS product. Simplicity is a feature.

The AI/ML curriculum is the initial/default course, not a permanent singleton baked into the platform. See [`multi-course.md`](multi-course.md) for the course ownership, routing, state, and implementation boundaries.

The system should remain understandable enough that one developer can hold most of it in their head.

## Runtime components

```text
Browser / installed PWA
┌─────────────────────────────────────────────┐
│ React + TypeScript                         │
│                                             │
│ App shell                                   │
│ ├─ curriculum                              │
│ ├─ reviews                                 │
│ ├─ exercises                               │
│ │   └─ CodeMirror + Pyodide Web Worker     │
│ ├─ progress                                │
│ └─ global tutor overlay                    │
└───────────────────┬─────────────────────────┘
                    │ HTTP / streamed HTTP
                    ▼
EC2
┌─────────────────────────────────────────────┐
│ Go process                                  │
│                                             │
│ HTTP API                                    │
│ ├─ curriculum metadata                     │
│ ├─ progress / activity                     │
│ ├─ reviews / FSRS                          │
│ └─ tutor orchestration                     │
│     ├─ context builder                     │
│     └─ providers                           │
│         ├─ OpenRouter                      │
│         └─ Codex                           │
└───────────────────┬─────────────────────────┘
                    │
                    ▼
                  SQLite
```

## Full-stack API contract

The Go API contract flows one way into the frontend:

```text
Go operations/types
    -> Huma-generated OpenAPI 3.1
    -> Orval
    -> generated TanStack Query client + generated Zod schemas
```

Do not maintain duplicate handwritten TypeScript DTOs or hand-written frontend request code alongside the Go API. Generated Zod schemas are the frontend runtime contract and the source of inferred API types.

CI will eventually regenerate the OpenAPI and frontend artifacts and fail on drift, and a deterministic frontend boundary check will reject raw application `fetch`/Axios calls and handwritten API-contract types outside explicitly approved infrastructure.

See [`api-contract.md`](api-contract.md) for the complete generation, validation, streaming, and enforcement rules.

## Content versus state

### Git is authoritative for curriculum content

Version-controlled content includes:

- courses and course metadata;
- modules;
- lessons;
- learning objectives and prerequisite IDs;
- curated videos;
- exercises and their definitions;
- project/lab descriptions;
- source metadata and citation IDs.

This content should be reviewable like source code.

### SQLite is authoritative for learner state

Persistent learner state will eventually include:

- objective progress;
- spaced-repetition scheduling and review history;
- exercise workspaces and attempts;
- activity history;
- notes/bookmarks;
- tutor conversations;
- project status;
- user preferences.

Course-bound learner state must carry explicit course identity. The fact that AI/ML is currently the only course must not be used as a persistence key or hidden global assumption.

Do not put authored curriculum prose into SQLite merely because it is convenient to query.

## Course and objective-centered model

A course is the top-level authored learning path. A module is an organizational container inside one course. A lesson is a teaching artifact. An objective represents a capability the learner is expected to develop.

```text
course
  └─ module
      ├─ lesson ───────────────┐
      ├─ video ────────────────┤
      ├─ review item ──────────┼──► objective ◄── prerequisite objective
      ├─ exercise ─────────────┤
      └─ project/lab ──────────┘
```

Progress should be explainable in terms of objectives rather than arbitrary page-completion percentages, while objective/activity identity remains qualified by its course where needed.

## Python execution boundary

There is one in-app Python runtime: Pyodide in the browser.

### Appropriate for Pyodide

- syntax/fundamentals exercises;
- numerical experiments;
- small NumPy tasks;
- calculus/linear-algebra exercises;
- small classical-ML exercises where supported packages fit comfortably;
- tiny algorithm implementations and tests.

### Not appropriate for the app runtime

- GPU training;
- substantial PyTorch work;
- large datasets;
- complex native dependencies;
- multi-file applications;
- long-running training or profiling work.

Those become repository-based labs/projects and are completed in a normal development environment.

**Do not add a Go-to-Python execution service when Pyodide is outgrown.** Crossing that boundary is intentional.

## Tutor architecture

The embedded tutor is a first-class application feature rather than a detached chatbot.

```text
current screen ───────┐
recent activity ──────┤
objective state ──────┼─► Tutor context builder
relevant notes ───────┤            │
curriculum sources ───┘            ▼
                              Tutor service
                                   │
                       ┌───────────┴───────────┐
                       ▼                       ▼
                 OpenRouter provider       Codex provider
                       │                       │
                       └───────────┬───────────┘
                                   ▼
                         normalized TutorEvent
                                   │
                                   ▼
                             streamed to UI
```

Provider adapters are allowed to differ internally. Normalize the event stream, not the provider request protocol.

Curriculum-related tutor context must include explicit course identity rather than assuming one global curriculum.

## Transport

Use ordinary HTTP for requests and streamed HTTP/SSE-style events for tutor output. The browser sends one tutor turn and the server streams events back.

Do not add WebSockets simply because chat is interactive. Reconsider only when there is a concrete bidirectional persistent-connection requirement.

The tutor stream is the narrow transport exception to the normal generated TanStack Query request path. It must still use the generated OpenAPI/Zod request and event contracts, with raw streaming transport isolated in one API-runtime adapter rather than feature code.

## Deployment shape

Initial deployment target:

```text
one EC2 instance
├─ Go process
├─ built React assets / reverse proxy arrangement
└─ SQLite database file
```

Back up the database file. Do not introduce distributed infrastructure unless the deployment model actually changes.

## Deferred decisions

The following should remain open until implementation pressure clarifies them:

- exact SQLite driver and migration package;
- exact FSRS library/schema mapping;
- authentication/reverse-proxy strategy for public internet exposure;
- how Codex subscription authentication is hosted on EC2;
- whether GitHub integration is useful for project/lab tracking;
- offline PWA behavior beyond cached app assets;
- exact citation rendering UX;
- the visible course-selection UX once a second course actually exists.
