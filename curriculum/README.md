# Curriculum

This directory contains authored curriculum content. It is version controlled and separate from learner state.

## Conventions

- Stable IDs use `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`: lowercase letters and numbers separated by single dots or hyphens.
- Modules organize learning material.
- `module.yaml` uses `id`, `title`, and explicit `order`. Its ordered `lessons` list is the canonical lesson sequence.
- Objectives describe capabilities and may reference globally unique prerequisite objective IDs, including objectives in other modules.
- Lessons are MDX.
- Lesson metadata is YAML frontmatter at the beginning of each MDX file. The Go loader preserves the remaining MDX source body and does not execute or compile it.
- Videos are curated resources attached to modules/objectives.
- Technical lesson content must cite reputable sources from `sources.yaml`.
- Larger labs/projects reference external Git repositories rather than embedding a second execution environment into Fonzytooter.

The authoring schemas are intentionally small. Run `cd server && go run ./cmd/curriculum-check ../curriculum` before committing curriculum changes.
