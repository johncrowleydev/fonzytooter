# Curriculum content authoring

## Source of truth

Curriculum content is stored in Git under `curriculum/`.

A course owns its modules. For example:

```text
curriculum/courses/ai-ml/modules/02-linear-algebra/
├── module.yaml
├── 01-vectors.mdx
├── 02-matrices.mdx
├── exercises.yaml
└── projects.yaml
```

The exact schema can evolve as real modules are authored. Do not build a generic CMS first.

## Stable IDs

All authored IDs use the same deterministic format: `^[a-z0-9]+(?:[.-][a-z0-9]+)*$`.
IDs are application identity; directory and filename choices are storage organization.

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

Each module should have a curated YouTube playlist/resource sequence. A video may support one or more objectives.

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

`curriculum/sources.yaml` is the authoritative source registry.

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

Build-time validation should eventually catch at least:

- unknown source IDs;
- malformed source records;
- lessons with no sources where sources are required;
- broken lesson/objective references.

AI-generated content follows the same rule: draft, verify sources, review, then commit.
