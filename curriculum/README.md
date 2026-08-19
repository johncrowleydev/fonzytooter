# Curriculum

This directory contains authored curriculum content. It is version controlled and separate from learner state.

Fonzytooter uses an explicit multi-course catalog. The AI/ML curriculum remains the initial/default course. See [`../docs/multi-course.md`](../docs/multi-course.md) for the ownership and directory model.

## Conventions

- Stable IDs use `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`: lowercase letters and numbers separated by single dots or hyphens.
- Courses are the top-level authored learning paths; modules belong to exactly one course.
- Courses live under `courses/<course-directory>/`, with metadata in `course.yaml` and modules under that course's `modules/` directory.
- `module.yaml` uses `id`, `title`, and explicit `order`. Its ordered `lessons` list is the canonical lesson sequence.
- Objectives describe capabilities and may reference prerequisite objective IDs, including objectives in other modules. Course identity must be explicit where required by the multi-course model.
- Lessons are MDX.
- Lesson metadata is YAML frontmatter at the beginning of each MDX file. The Go loader preserves the remaining MDX source body and does not execute or compile it.
- Videos are curated resources attached to modules/objectives.
- Technical lesson content must cite reputable sources from `sources.yaml`.
- Larger labs/projects reference external Git repositories rather than embedding a second execution environment into Fonzytooter.

The authoring schemas are intentionally small. Run `cd server && go run ./cmd/curriculum-check ../curriculum` before committing curriculum changes.
