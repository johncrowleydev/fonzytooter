# Fonzytooter

Fonzytooter is a deliberately small, single-user learning system for self-paced technical courses. The AI/ML curriculum is the initial/default course and the current content focus, but the architecture is intended to support additional authored courses without becoming a general-purpose LMS.

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

The curriculum is versioned content in Git. SQLite stores learner state.

See [`docs/curriculum.md`](docs/curriculum.md) for the high-level AI/ML course plan, [`docs/multi-course.md`](docs/multi-course.md) for the multi-course ownership and migration model, [`docs/architecture.md`](docs/architecture.md) for the architectural boundaries, and [`AGENTS.md`](AGENTS.md) before making structural changes.

## Core concepts

### Courses and learning objectives

A course is the top-level authored learning path; modules belong to courses. The AI/ML course is currently the only authored course, but course identity should be explicit in routing, state, and curriculum ownership rather than inferred from that fact.

Objectives are the connective tissue of the system. Lessons teach objectives; videos support them; reviews reinforce them; exercises assess them; projects integrate them.

### No calendar curriculum

Time matters to spaced-repetition scheduling, but the curriculum itself has no week numbers, due dates, or "behind schedule" state. Progress is based on what has been introduced, retained, applied, and demonstrated.

### One in-app Python runtime

Pyodide is the only Python interpreter built into Fonzytooter. Small learning exercises run in-browser.

If an exercise becomes large enough to require a server-side Python runtime, GPU, large dataset, complex environment, or multi-file project, it is no longer an in-app exercise. It becomes a Git-based lab/project completed in a normal editor or IDE.

Jupyter notebooks sit between those two modes: they are used outside Fonzytooter for exploratory experiments, plots, data inspection, and intuition-building. Fonzytooter does not embed or host Jupyter.

### Tutor everywhere

The tutor is mounted at the application-shell level and can be opened from any screen. Each turn receives structured context about the current screen and relevant recent activity. The tutor does not need screenshots to know what the learner is doing.

## Repository layout

The current repository still uses the initial single-course curriculum layout; the planned course-aware migration is documented in [`docs/multi-course.md`](docs/multi-course.md).

```text
.
├── AGENTS.md
├── curriculum/
│   ├── modules/
│   └── sources.yaml
├── docs/
├── server/
│   ├── cmd/fonzytooter/
│   └── internal/
└── web/
    └── src/
```

## Development

### Prerequisites

- Go 1.26+
- Node.js compatible with Vite 8
- npm

### Start the API

```bash
cd server
go run ./cmd/fonzytooter
```

The API listens on `:8080` by default. Override it with `FONZYTOOTER_ADDR`.
The development default loads Git-authored curriculum from `../curriculum`;
override it with `FONZYTOOTER_CURRICULUM_PATH`. Invalid curriculum prevents the
server from starting. To validate it without starting the API:

```bash
cd server
go run ./cmd/curriculum-check ../curriculum
```

### Start the web app

```bash
cd web
npm install
npm run dev
```

Vite proxies `/api` to `http://localhost:8080` during local development.

### Current scaffold status

The initial scaffold intentionally implements very little business logic:

- the Go API has a health endpoint and the streaming tutor endpoint shape;
- the tutor UI is globally available and already consumes streamed tutor events;
- the default tutor provider is an explicit "not configured" provider;
- curriculum folders and content conventions are established;
- the Pyodide exercise boundary is documented and typed, but the editor/runtime is not wired yet;
- persistence, FSRS, provider integrations, and real curriculum content come next.

That is intentional. Add capabilities when the learning workflow needs them rather than pre-building an LMS platform.
