# Curriculum content authoring

## Source of truth

Curriculum content is stored in Git under `curriculum/`.

A course owns its modules. The current AI/ML course uses:

```text
curriculum/courses/ai-ml/
├── course.yaml
└── modules/
    ├── 00-orientation/
    └── 01-scientific-python/
```

A representative future module may look like:

```text
curriculum/courses/ai-ml/modules/02-linear-algebra/
├── module.yaml
├── 01-vectors.mdx
├── 02-matrices.mdx
├── exercises.yaml
└── projects.yaml
```

The exact activity schemas can evolve as real modules are authored. Do not build a generic CMS first.

Long-range subject planning is separate from runtime-authored content. Course plans live under [`courses/`](courses/); for example, [`courses/ai-ml.md`](courses/ai-ml.md) describes the broader AI/ML curriculum direction while `curriculum/courses/ai-ml/` contains the concrete content the application loads.

## Course metadata

Each course directory contains a small `course.yaml`.

Current shape:

```yaml
id: ai-ml
title: AI & Machine Learning
description: Build practical foundations in Python, mathematics, and machine learning.
order: 0
```

Course metadata is intentionally limited to authored identity, presentation, and ordering. Do not add enrollment, instructor, tenancy, permissions, semester, or other LMS administration fields without a real product requirement.

## Stable IDs

All authored IDs use the same deterministic format: `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`.
IDs are application identity; directory and filename choices are storage organization.

Course identity is explicit when modules and lessons are resolved. Do not write new code that assumes a module or lesson is globally rooted merely because the current authored catalog contains one course.

## Module metadata

Example shape:

```yaml
id: linear-algebra
title: Linear Algebra
order: 1
objectives:
  - id: linear-algebra.vectors
    title: Work with vectors
    description: Represent and manipulate vectors.
    prerequisites: []
videos:
  - id: linear-algebra-introduction
    title: Example title
    url: https://example.com/video
    objectiveIds:
      - linear-algebra.vectors
lessons:
  - 01-vectors
```

Module order is scoped to its owning course. Its ordered `lessons` list is the canonical lesson sequence.

Each module should have a curated YouTube playlist/resource sequence where useful. A video may support one or more objectives.

## Lesson MDX

MDX is used because a lesson needs normal prose plus React-powered interactive material.

Conceptual example:

```mdx
---
id: 01-vectors
title: Vectors
objectiveIds:
  - linear-algebra.vectors
sourceIds:
  - some-stable-source-id
---

# Vectors

Vectors ...

<Cite source="some-stable-source-id" section="..." />

<GradientDescentExplorer />

<KnowledgeCheck objective="optimization.gradient-descent" />
```

Interactive components should be added when they genuinely improve understanding, not because MDX makes them possible.

## Sources

`curriculum/sources.yaml` is the shared authoritative source registry across courses.

Source IDs should be stable. Record enough bibliographic metadata to verify what was used.

Acceptable source types include:

- original research papers;
- official documentation/specifications;
- reputable textbooks;
- reputable university course material;
- other high-quality primary or educational sources when appropriate.

Do not fabricate citations or bibliographic fields.

## Citation rule

Every substantial explanatory section should have an inspectable source basis.

This does not mean mechanically adding a footnote to every sentence. It means a reader should be able to tell which reputable sources support the technical material being taught.

AI-generated content follows the same rule: draft, verify sources, review, then commit.

## Validation

The Go curriculum loader validates authored structure before the application starts. The standalone check uses the same loader:

```bash
cd server
go run ./cmd/curriculum-check ../curriculum
```

Current validation includes the curriculum root/course/module structure, strict YAML fields, stable IDs, course/module ordering constraints, lesson declarations/frontmatter, source/objective references, prerequisite references/cycles, duplicate authored identities, and other catalog invariants.

The frontend MDX validation pass recursively compiles authored lesson MDX under all courses:

```bash
cd web
npm run curriculum:mdx
```

Add new deterministic authoring checks when real content failures justify them; do not replace structural validation with agent instructions or tutor judgment.
