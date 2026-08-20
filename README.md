# Fonzytooter

Fonzytooter is a deliberately small, single-user learning system for self-paced technical courses. **AI & Machine Learning** (`ai-ml`) is the initial/default course and the current content focus, while the implemented curriculum model supports additional authored courses without turning the application into a general-purpose LMS.

It is not intended to become a general-purpose LMS. The system exists to support one learning loop:

> learn → practice → recall → reflect → adapt

## Goals

- Self-paced curriculum with no calendar deadlines or artificial pacing.
- Multiple explicitly modeled courses while keeping the product single-user and small.
- Objective-centered progress rather than vague course-completion percentages.
- Spaced repetition for retention and review scheduling.
- MDX lessons with interactive React components and visualizations.
- A curated YouTube playlist for each module.
- Reputable sources cited throughout lesson content.
- Small in-browser Python exercises using CodeMirror + Pyodide.
- Jupyter notebook labs for exploratory scientific/ML work and intuition-building.
- Larger labs and projects move into normal Git repositories and a real IDE.
- A context-aware AI tutor available from every screen.
- Tutor inference through OpenRouter and Codex-compatible provider adapters with streamed output.
- One Go service, one React app, one SQLite database, one user.

## Architecture

```text
                         Git repository
                  curriculum MDX / YAML / sources
                               │
                               ▼
┌──────────────────────────────────────────────────────────┐
│ React + TypeScript + Vite + Tailwind PWA                │
│                                                          │
│ lessons · reviews · exercises · progress · tutor overlay │
│                                 │                        │
│                         CodeMirror + Pyodide              │
└───────────────────────────┬──────────────────────────────┘
                            │ /api
                            ▼
┌──────────────────────────────────────────────────────────┐
│ Go monolith                                              │
│                                                          │
│ curriculum metadata · progress · FSRS · tutor service    │
│                                     │                    │
│                           provider adapters               │
│                         OpenRouter · Codex                │
└───────────────────────────┬──────────────────────────────┘
                            │
                            ▼
                          SQLite
```

Curriculum is versioned content in Git. SQLite is reserved for learner state.

See [`docs/courses/`](docs/courses/) for course-specific curriculum plans, [`docs/multi-course.md`](docs/multi-course.md) for platform-wide course ownership and routing, [`docs/architecture.md`](docs/architecture.md) for the architectural boundaries, and [`AGENTS.md`](AGENTS.md) before making structural changes.

## Core concepts

### Courses and learning objectives

A course is the top-level authored learning path; modules belong to courses. AI/ML is currently the only authored course, but course identity is explicit in authored ownership, catalog lookup, HTTP resources, frontend routes, and curriculum-related tutor context rather than inferred from that fact.

Objectives are the connective tissue of the system. Lessons teach objectives; videos support them; reviews reinforce them; exercises assess them; projects integrate them.

### No calendar curriculum

Time matters to spaced-repetition scheduling, but the curriculum itself has no week numbers, due dates, or "behind schedule" state. Progress is based on what has been introduced, retained, applied, and demonstrated.

### One in-app Python runtime

Pyodide is the only Python interpreter built into Fonzytooter. Small learning exercises run in-browser.

If an exercise becomes large enough to require a server-side Python runtime, GPU, large dataset, complex environment, or multi-file project, it is no longer an in-app exercise. It becomes a Git-based lab/project completed in a normal editor or IDE.

Jupyter notebooks sit between those two modes: they are used outside Fonzytooter for exploratory experiments, plots, data inspection, and intuition-building. Fonzytooter does not embed or host Jupyter.

### Tutor everywhere

The tutor is mounted at the application-shell level and can be opened from any screen. Each turn receives structured context about the current screen and relevant recent activity. Curriculum page context includes explicit course identity. The tutor does not need screenshots to know what the learner is doing.

## Repository layout

```text
.
├── AGENTS.md
├── curriculum/
│   ├── courses/
│   │   └── ai-ml/
│   │       ├── course.yaml
│   │       └── modules/
│   └── sources.yaml
├── docs/
│   └── courses/
├── openapi/
├── server/
│   ├── cmd/fonzytooter/
│   └── internal/
└── web/
    └── src/
```

The runtime curriculum layout and ownership model are documented in [`docs/multi-course.md`](docs/multi-course.md). The current AI/ML long-range curriculum plan is [`docs/courses/ai-ml.md`](docs/courses/ai-ml.md).

## Curriculum routes and API

The current authored curriculum is navigated through course-aware routes:

```text
/courses/:courseId
/courses/:courseId/modules/:moduleId
/courses/:courseId/modules/:moduleId/lessons/:lessonId
```

`/curriculum` redirects to the current default course at `/courses/ai-ml`.

The corresponding read API is:

```text
GET /api/courses
GET /api/courses/{courseId}
GET /api/courses/{courseId}/modules/{moduleId}
GET /api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}
GET /api/courses/{courseId}/modules/{moduleId}/exercises/{exerciseId}
```

Go operations/types generate OpenAPI, which generates the frontend TanStack Query client and Zod schemas. Feature code should not create a parallel handwritten API contract.

## Development

### Prerequisites

- Go 1.26+
- Node.js compatible with Vite 8
- npm
- [Pandoc](https://pandoc.org/) for worksheet Markdown-to-LaTeX conversion
- [Tectonic](https://tectonic-typesetting.github.io/) for worksheet PDF typesetting

Both `pandoc` and `tectonic` must be available on `PATH` for worksheet and
solution PDF downloads. The API returns `503 Service Unavailable` for those
document resources when either tool is missing.

### Start the API

```bash
cd server
go run ./cmd/fonzytooter
```

The API listens on `:8080` by default. Override it with `FONZYTOOTER_ADDR`.
The development default loads Git-authored curriculum from `../curriculum`;
override it with `FONZYTOOTER_CURRICULUM_PATH`. The path is the curriculum root, not an individual course. Invalid curriculum prevents the server from starting. To validate it without starting the API:

```bash
cd server
go run ./cmd/curriculum-check ../curriculum
```

Learner state is stored in SQLite at `./data/fonzytooter.db` by default. Override
the location with `FONZYTOOTER_DB_PATH`; production deployments should use an
absolute path on persistent storage. The server creates the parent directory and
runs embedded migrations before it begins serving HTTP. Back up the database
file as part of the deployment backup policy (including a SQLite-safe checkpoint
or backup procedure while WAL mode is active).

With Pandoc and Tectonic installed, render every authored student/solutions worksheet and module workbook without storing output:

```bash
cd server
go run ./cmd/worksheet-render-check ../curriculum
```

### Start the web app

```bash
cd web
npm install
npm run dev
```

Vite proxies `/api` to `http://localhost:8080` during local development.

### Current implementation status

The project now has the core curriculum delivery path rather than only an initial scaffold:

- the Go curriculum loader validates and indexes an explicitly multi-course Git-authored catalog;
- the curriculum HTTP API exposes course-qualified course/module/lesson resources;
- Huma-generated OpenAPI drives the generated Orval/TanStack Query + Zod frontend contract;
- React Router renders real course, module, and lesson pages from that generated API;
- authored lesson MDX renders through the trusted lesson component registry, including the current Scientific Python interactives;
- Git-authored worksheets render in the web UI and as on-demand Pandoc/Tectonic student, solutions, and module workbook PDFs;
- optional Git-authored embedded exercise definitions are strictly validated and exposed through a student-safe course-qualified API that omits hidden test code;
- `/curriculum` remains the simple navigation entry point and redirects to the default AI/ML course;
- the global tutor receives explicit course/module/lesson semantic page context and consumes streamed tutor events;
- the default tutor provider is still an explicit "not configured" provider;
- embedded exercises run in a Pyodide worker with CodeMirror, saved SQLite workspaces, checked attempts, and authored visible/hidden tests;
- lesson completion, objective introduction, and recent learner activity now persist through SQLite-backed generated APIs and drive the lesson, progress, and dashboard UI;
- FSRS reviews and checked exercises now provide factual recall/application evidence on the dashboard and progress views; transfer, real tutor providers, worksheet uploads, and tutor grading remain future implementation work.

Add capabilities when the learning workflow needs them rather than pre-building an LMS platform.
