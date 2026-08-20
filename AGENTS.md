# AGENTS.md

Instructions for coding agents working in this repository.

## What this project is

Fonzytooter is a personal, single-user technical learning system. The AI/ML curriculum is the initial/default course, but the platform is intentionally course-aware so additional authored courses can be added later without retrofitting single-course assumptions. It is still intentionally small. Favor straightforward code and explicit data flow over generalized infrastructure.

Read the relevant files in `docs/` before making architectural changes. Multi-course work must follow `docs/multi-course.md`. Frontend work must also follow `docs/frontend.md`. API/backend/frontend-contract work must follow `docs/api-contract.md`, and HTTP API design must follow `docs/api-style.md`.

## Non-negotiable architectural constraints

1. **Keep it a monolith.** One Go backend, one React frontend, one SQLite database.
2. **Do not add server-side Python execution.** Pyodide is the only Python interpreter inside the learning app.
3. **Outgrown Pyodide means a real project.** GPU work, large datasets, substantial training jobs, or multi-file ML work belong in Git repositories and a normal IDE.
4. **Curriculum content lives in Git.** Courses/modules/lessons/sources are MDX/YAML and version controlled. Learner state belongs in SQLite.
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
15. **The HTTP API is resource-oriented REST.** Paths identify resources and collections; HTTP methods carry CRUD/state-transition semantics; resource bodies stay resource-shaped; transport metadata belongs in headers; analogous operations use consistent naming and status codes; action/RPC endpoints are avoided. See `docs/api-style.md`.
16. **Course is a first-class ownership boundary.** Do not add curriculum routes, APIs, tutor context, persistence keys, worksheets, exercises, projects, or other course-bound state that assumes AI/ML is the only course. AI/ML may be the default for navigation, but it is not the platform-wide identity. See `docs/multi-course.md`.

## Agent workspace safety

- **Never extract archives into tracked repository content.** ZIP files and other temporary source bundles must be extracted either outside the repository or into an explicitly untracked/gitignored working directory. Temporary extracted contents must never enter Git history.
- **Use Git worktrees for agent tasks.** Each agent/task should normally work in its own worktree and branch, especially when work may happen in parallel. Multiple agents must not mutate the same working tree and risk overwriting or interleaving each other's changes.
- **Use purpose-based branch and commit prefixes.** Names should describe the reason for the change, such as `docs/`, `fix/`, or `feat/`, rather than the agent or author identity.
- **Inspect the working tree before committing.** Review `git status` and the relevant diff so temporary files, extracted artifacts, or another agent's work are not accidentally included.

## Engineering preferences

### Go

- Prefer the standard library where it is sufficient.
- Keep HTTP handlers thin and domain logic in focused packages.
- Prefer explicit structs and interfaces to framework magic.
- Keep dependencies few and intentional.
- Run `gofmt` on changed Go files.
- Add tests around real behavior rather than scaffolding abstractions for hypothetical future use.
- Application API operations should participate in the Huma/OpenAPI contract rather than being separately documented ad-hoc handlers.
- Before adding a route, model the resource first and verify the requirement cannot be expressed cleanly by extending an existing resource operation. Follow `docs/api-style.md` for paths, verbs, bodies, headers, idempotency, and status codes.

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

See `docs/frontend.md` for the complete frontend conventions, `docs/api-contract.md` for API generation/validation rules, and `docs/api-style.md` for HTTP resource design conventions.

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
- current course/module/lesson/exercise/objective IDs;
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
