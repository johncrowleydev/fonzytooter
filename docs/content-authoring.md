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

## Worksheets

A module may contain an optional direct-child `worksheets/` directory. Worksheet discovery is file-based; do not add a worksheet list to `module.yaml`.

```text
<module>/
├── module.yaml
├── lesson.mdx
└── worksheets/
    └── function-practice.yaml
```

Each `.yaml` file in that directory has this authored shape:

```yaml
id: function-practice
title: Function Practice
lessonId: functions
order: 0
objectiveIds:
  - mathematics.functions
instructions: |
  Show your reasoning for each problem.
problems:
  - id: evaluate-function
    prompt: |
      Evaluate $f(3)$ when $f(x)=2x+1$.
    objectiveIds:
      - mathematics.functions
    expectedAnswer: |
      $f(3)=2(3)+1=7$.
    requiresWork: true
    responseLines: 4
    rubric:
      - Substitutes the input into the function correctly.
      - Computes the final value as `7`.
```

`lessonId` must name a lesson in the owning module. Worksheet `order` is scoped to that lesson; order values must be unique within a lesson. Module-wide worksheet presentation follows the module's declared lesson order, then worksheet order within each lesson.

`instructions`, `prompt`, and `expectedAnswer` are Markdown and may contain LaTeX math using dollar delimiters. The loader preserves this prose exactly as authored. `requiresWork` and `responseLines` must be explicit for every problem, and `responseLines` must be greater than zero.

The student worksheet JSON API intentionally omits `expectedAnswer` and `rubric`; those fields remain internal curriculum data for solution rendering and future assessment behavior.

Individual worksheet PDFs and module workbooks are generated on demand from this same validated structure using Pandoc and Tectonic. Workbooks are not separately authored and PDFs are not stored. A workbook compiles the module's ordered worksheets as one document rather than concatenating previously rendered PDFs. Student output excludes expected answers and rubrics; solutions output includes expected answers but still excludes rubric criteria.

Install `pandoc` and `tectonic` on `PATH`, then exercise every checked-in worksheet and non-empty module in both student and solutions form with:

```bash
cd server
go run ./cmd/worksheet-render-check ../curriculum
```

The command uses temporary rendering workspaces and leaves no PDFs in the repository. Run it after authoring worksheet Markdown or LaTeX so invalid typesetting fails before review.

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

Worksheet validation is part of the same deterministic command. It checks lesson/objective references, lesson-scoped ordering, required authored fields, stable and unique IDs, response layout hints, and non-empty rubrics.

The frontend MDX validation pass recursively compiles authored lesson MDX under all courses:

```bash
cd web
npm run curriculum:mdx
```

Add new deterministic authoring checks when real content failures justify them; do not replace structural validation with agent instructions or tutor judgment.
