# Multi-course architecture

## Decision

Fonzytooter is a single-user learning system with an explicitly multi-course curriculum model.

The **AI & Machine Learning** course (`ai-ml`) is currently the only authored course, the default course, and the product's immediate content focus. It is not a platform-wide singleton: course identity is explicit in authored content ownership, catalog lookup, HTTP resources, frontend routes, and curriculum-related tutor context.

This does **not** turn Fonzytooter into a generic LMS. The existing simplicity constraints still apply: one user, one Go service, one React app, one SQLite database, Git-authored curriculum, and no speculative enrollment, tenancy, marketplace, classroom, CMS, or enterprise infrastructure.

A likely future course is computer science for experienced working programmers who did not study computer science formally. That is useful design pressure, but no computer-science course content is currently authored.

Course-specific curriculum plans live under [`courses/`](courses/). Platform-wide ownership and routing rules belong here.

## Domain hierarchy

`Course` is a first-class curriculum concept above `Module`.

```text
Catalog
└── Course
    ├── metadata
    └── Module
        ├── objectives
        ├── lessons
        │   └── worksheet(s)
        ├── videos
        ├── exercises
        ├── notebook labs
        └── projects / synthesis work
```

The exact set of activity types will continue to evolve, but the ownership rule is stable:

> A course owns modules. A module belongs to exactly one course. Course-bound learning artifacts must be traceable back to that course.

The in-memory curriculum catalog represents the full authored catalog. Course-aware lookups verify the complete ownership hierarchy instead of relying on globally rooted module or lesson resources.

## Course identity

Every course has a stable authored ID, for example:

```text
ai-ml
computer-science
```

At routing, API, persistence, tutor-context, and other durable boundaries, course membership must be explicit whenever a resource is course-bound. Code must not infer the course from the fact that only one currently exists.

The current catalog keeps objective IDs globally unique, preserving the existing prerequisite-reference model without inventing cross-course reference syntax prematurely. Module IDs are unique within their owning course and are always resolved together with course identity. Two courses may therefore use the same module ID without weakening ownership isolation.

A future explicit cross-course reference can be designed when one is actually authored. A generic cross-course prerequisite graph is not required.

The source registry remains shared across courses so the same authoritative source may support material in more than one course.

## Git-authored content layout

Authored curriculum uses this course-aware structure:

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

A future second course follows the same shape:

```text
curriculum/courses/computer-science/
├── course.yaml
└── modules/
```

The current `course.yaml` schema is intentionally small:

```yaml
id: ai-ml
title: AI & Machine Learning
description: Build practical foundations in Python, mathematics, and machine learning.
order: 0
```

Course metadata exists to identify, order, and present authored courses. It is not an LMS enrollment or administration schema.

The loader validates the curriculum root, course metadata, module/lesson/objective/source references, ordering constraints, prerequisite relationships, and other authored invariants before constructing the immutable in-memory catalog. `FONZYTOOTER_CURRICULUM_PATH` points to the curriculum root, not to one course.

See [`../curriculum/README.md`](../curriculum/README.md) and [`content-authoring.md`](content-authoring.md) for authoring conventions.

## Application routing

Curriculum-bound routes carry explicit course identity:

```text
/courses/:courseId
/courses/:courseId/modules/:moduleId
/courses/:courseId/modules/:moduleId/lessons/:lessonId
```

The convenience route:

```text
/curriculum
```

redirects to the current default course:

```text
/courses/ai-ml
```

That default is an application-navigation preference kept in one obvious frontend location. It is not used by backend catalog lookups or as hidden durable identity.

There is intentionally no visible course picker or generic `/courses` catalog screen while only one real course exists. When a second authored course makes course selection useful, the navigation can expose it then.

## HTTP API

The curriculum read API follows the same ownership hierarchy:

```text
GET /api/courses
GET /api/courses/{courseId}
GET /api/courses/{courseId}/modules/{moduleId}
GET /api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}
```

`GET /api/courses/{courseId}` includes the ordered module summaries needed to render the course page, so a separate module-collection endpoint is not currently necessary.

Module and lesson resources include course identity. Handlers use course-aware catalog lookups, so a valid module or lesson requested beneath the wrong course returns the standard not-found response rather than being resolved globally.

The former globally rooted `/api/modules/...` operations have been removed.

As with the rest of the application API, Go operations/types are authoritative and flow through Huma-generated OpenAPI to Orval-generated TanStack Query clients and Zod schemas. See [`api-contract.md`](api-contract.md) and [`api-style.md`](api-style.md).

## Tutor context

The tutor remains global to the application, but curriculum context is course-qualified.

Curriculum, module, and lesson page context includes explicit `courseId` and `courseTitle` along with the relevant module/lesson/objective identity. Selected-text lesson tutoring preserves that course identity as well.

Route identity is established even while data is loading or when a resource is invalid, so the tutor does not silently retain semantic context from the previously viewed course or lesson.

The tutor must not infer course ownership from module IDs or from the default course.

Cross-course tutoring is still allowed naturally: the tutor may eventually recommend useful material from another course. That does not require dumping every course or all learner state into every turn.

## Learner state

Course identity must be part of persisted state for course-bound learning activity as persistence features are implemented.

This includes:

- objective progress and evidence;
- lesson activity;
- spaced-repetition items tied to curriculum objectives;
- worksheet attempts and tutor-reviewed worksheet evidence;
- exercise workspaces and attempts;
- notebook/lab/project status;
- course-specific continue-learning state;
- tutor activity referring to curriculum resources.

The default course is a navigation preference, not a substitute for durable course identity.

The multi-course foundation was deliberately established before substantial learner-state persistence so new state schemas can use the correct ownership model from the beginning.

## Worksheets and workbooks

Worksheets remain optional practice artifacts attached to lessons as defined in [`worksheets.md`](worksheets.md).

Their ownership chain is:

```text
course -> module -> lesson -> worksheet
```

A module workbook aggregates worksheets from that module within its course. Worksheet implementation should build directly on the course-aware catalog, routes, and future persistence keys rather than introducing a single-course compatibility layer.

## Course-specific planning documents

Long-range curriculum planning is separate from the runtime-authored catalog.

```text
docs/courses/<course-id>.md
    -> subject-specific curriculum direction and future syllabus planning

curriculum/courses/<course-id>/
    -> concrete authored metadata, modules, lessons, and objectives loaded by the app
```

The current AI/ML curriculum plan is [`courses/ai-ml.md`](courses/ai-ml.md).

This keeps phrases such as "the AI/ML curriculum" from being confused with platform-wide curriculum architecture as additional courses are introduced.

## Current user experience

With only AI/ML authored, the multi-course architecture is intentionally unobtrusive:

- AI/ML remains the default course;
- `/curriculum` still provides the normal global navigation entry point;
- existing modules and lessons preserve their authored ordering and content;
- course/module/lesson URLs now carry explicit course identity;
- no empty marketplace or course-selection interface is shown.

The value is architectural: new platform features can now be built without embedding an assumption that AI/ML is the only possible course.

## Non-goals

Do not use multi-course support as justification to build:

- multi-user accounts or organizations;
- enrollment workflows;
- classroom/teacher features;
- permissions or RBAC around courses;
- a course marketplace;
- a generic course CMS or visual authoring system;
- course search/discovery infrastructure for a tiny personal catalog;
- calendar semesters, due dates, or cohort pacing;
- a generic cross-course dependency/knowledge graph;
- separate services or databases per course.

Fonzytooter remains a small personal learning application. The architecture is **multi-course**, not **general-purpose LMS**.
