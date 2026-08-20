# Curriculum

This directory contains authored curriculum content. It is version controlled and separate from learner state.

Fonzytooter uses an explicit multi-course catalog. **AI & Machine Learning** (`ai-ml`) remains the initial/default course. See [`../docs/multi-course.md`](../docs/multi-course.md) for the platform ownership model and [`../docs/courses/ai-ml.md`](../docs/courses/ai-ml.md) for the high-level AI/ML curriculum plan.

## Layout

```text
curriculum/
├── courses/
│   └── ai-ml/
│       ├── course.yaml
│       └── modules/
│           ├── 00-orientation/
│           └── 01-scientific-python/
└── sources.yaml
```

`FONZYTOOTER_CURRICULUM_PATH` points at this curriculum root. The loader discovers and validates authored courses beneath `courses/`.

## Conventions

- Stable IDs use `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`: lowercase letters and numbers separated by single dots or hyphens.
- Courses are the top-level authored learning paths; modules belong to exactly one course.
- Courses live under `courses/<course-directory>/`, with metadata in `course.yaml` and modules under that course's `modules/` directory.
- `course.yaml` currently uses `id`, `title`, `description`, and explicit non-negative `order`.
- `module.yaml` uses `id`, `title`, and explicit non-negative `order`. Its ordered `lessons` list is the canonical lesson sequence.
- Module ordering is scoped to the owning course.
- Objectives require descriptions and may reference prerequisite objective IDs. Objective IDs remain globally unique in the current catalog model.
- Lessons are MDX.
- Lesson metadata is YAML frontmatter at the beginning of each MDX file. The Go loader preserves the remaining MDX source body and does not execute or compile it.
- Videos are curated resources attached to modules/objectives.
- Optional `worksheets/`, `exercises/`, and `reviews/` directories are closed-world: every direct child must be a regular `.yaml` file.
- Reference arrays do not allow duplicate IDs.
- Technical lesson content must cite reputable sources from the shared `sources.yaml` registry.
- Larger labs/projects reference external Git repositories rather than embedding a second execution environment into Fonzytooter.

The authoring schemas are intentionally small. Validate curriculum changes with:

```bash
cd server
go run ./cmd/curriculum-check ../curriculum
```

Lesson MDX across all authored courses is also checked by the frontend validation workflow:

```bash
cd web
npm run curriculum:mdx
```
