# AGENTS.md

Instructions for coding agents working in this repository.

## What this project is

Fonzytooter is a personal, single-user AI/ML learning system. It is intentionally small. Favor straightforward code and explicit data flow over generalized infrastructure.

Read the relevant files in `docs/` before making architectural changes. Frontend work must also follow `docs/frontend.md`. API/backend/frontend-contract work must follow `docs/api-contract.md`.

## Non-negotiable architectural constraints

1. **Keep it a monolith.** One Go backend, one React frontend, one SQLite database.
2. **Do not add server-side Python execution.** Pyodide is the only Python interpreter inside the learning app.
3. **Outgrown Pyodide means a real project.** GPU work, large datasets, substantial training jobs, or multi-file ML work belong in Git repositories and a normal IDE.
4. **Curriculum content lives in Git.** Lessons/modules/sources are MDX/YAML and version controlled. Learner state belongs in SQLite.
5. **Learning objectives are the primary domain concept.** Avoid inventing parallel progress models that cannot map back to objectives.
6. **No calendar pacing.** Do not add week numbers, deadlines, streak pressure, or "behind schedule" concepts unless explicitly requested.
7. **The tutor is global.** It should be available from any screen and receive structured semantic page context plus relevant recent activity.
8. **Do not use screenshots as the normal tutor-context mechanism.** The app already knows what is on screen.
9. **Tutor providers are adapters.** OpenRouter and Codex may have different request/runtime models. Do not force them into identical upstream APIs; normalize streamed events coming out.
10. **Prefer HTTP streaming/SSE semantics over WebSockets** unless a concrete requirement makes bidirectional persistent transport necessary.
11. **AI does not award mastery by vibe.** The tutor may recommend an assessment, but progress/mastery should come from review history, exercises, tests, and explicit learner actions.
12. **Sources are required for substantive lesson content.** Never invent citations. Reputable primary or high-quality educational sources should be recorded in `curriculum/sources.yaml` and referenced by stable IDs.
13. **Avoid speculative infrastructure.** No microservices, Redis, queues, Kubernetes, GraphQL, vector database, event sourcing, generic CMS, or enterprise RBAC without a demonstrated need.
14. **The API contract is generated end-to-end.** Go operations/types generate OpenAPI; OpenAPI generates the TanStack Query client and Zod schemas. Frontend feature code must not hand-write API DTOs or bypass the generated API boundary with raw `fetch`/Axios. See `docs/api-contract.md`.

## Engineering preferences

### Go

- Prefer the standard library where it is sufficient.
- Keep HTTP handlers thin and domain logic in focused packages.
- Prefer explicit structs and interfaces to framework magic.
- Keep dependencies few and intentional.
- Run `gofmt` on changed Go files.
- Add tests around real behavior rather than scaffolding abstractions for hypothetical future use.
- Application API operations should participate in the Huma/OpenAPI contract rather than being separately documented ad-hoc handlers.

### TypeScript / React

- TypeScript should be strict.
- Prefer explicit types at domain boundaries.
- Keep components small and feature-oriented.
- Extract components around meaningful UI/behavior boundaries, not arbitrary line counts.
- Minimize `useEffect`; use it for synchronization with external systems, not as a general control-flow tool.
- Avoid global state libraries until React state/context becomes materially painful.
- **Use React Router for application routing. Do not hand-roll routing with the History API.**
- **Tailwind is the styling system. Do not add custom CSS unless Tailwind cannot reasonably express the requirement; document the justification when custom CSS is necessary.**
- **All TypeScript/TSX must be automatically formatted for human readability. Prettier is the standard formatter. Do not compress JSX onto long single lines.**
- **Ordinary server-state/API access must use the generated TanStack Query client. Do not add feature-level raw `fetch`, Axios, or handwritten API DTOs.**
- **Generated Zod schemas are the frontend runtime API contract and the source for inferred API types via `z.infer`, `z.input`, or `z.output`.**
- MDX is the lesson-content format.

See `docs/frontend.md` for the complete frontend conventions and `docs/api-contract.md` for API generation/validation rules.

### Python exercises

- CodeMirror is the planned editor.
- Pyodide runs in a Web Worker so user code cannot freeze the UI thread.
- `Run` means exploratory execution; `Check`/`Submit` means assessment against tests.
- Small deterministic tests and mathematically meaningful tolerances/properties are preferred over exact-output-only grading when appropriate.
- Do not create a backend code runner.

## Tutor contract

The tutor should eventually support modes such as explain, Socratic, exercise help, quiz, and explore.

A tutor turn may include:

- conversation history;
- current page/lesson/exercise/objective IDs;
- selected text or current exercise code when explicitly relevant;
- recent attempts and test failures;
- relevant objective state;
- relevant learner notes;
- curriculum source excerpts/metadata.

Do not dump the learner's entire history into every request. Build context intentionally.

Tutor writes that affect persistent learner state should be explicit actions. Reading progress can be automatic; changing progress, creating cards, or recording notes should be traceable.

## Content rules

- Do not write technical lesson content from model memory alone.
- Verify claims against reputable sources before committing curriculum content.
- Prefer primary sources for papers, official documentation for tools/frameworks, and reputable textbooks/course material for explanations.
- Every substantial lesson section should make its source basis inspectable.
- YouTube videos are curated curriculum resources, not authoritative citations by default.

## Before adding a dependency or subsystem

Ask:

1. What current user-facing problem does this solve?
2. Can the existing Go/React/SQLite stack solve it simply?
3. Does it violate an explicit boundary above?
4. Will this make maintaining the learning system compete with actually using it to learn?

When in doubt, choose the smaller implementation.
