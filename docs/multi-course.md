# Multi-course architecture

## Decision

Fonzytooter is a single-user learning system that may contain multiple authored courses.

The current AI/ML curriculum remains the initial course, the default course, and the product's immediate content focus. Multi-course support exists so that new courses can be added later without extracting single-course assumptions from routing, curriculum loading, learner state, tutor context, worksheets, and other platform features.

This is **not** a decision to turn Fonzytooter into a generic LMS. The existing simplicity constraints still apply: one user, one Go service, one React app, one SQLite database, Git-authored curriculum, and no speculative enrollment, tenancy, marketplace, authoring-platform, or enterprise infrastructure.

A likely second course is computer science for an experienced working programmer who is largely self-taught: data structures and algorithms, computer architecture, compilers and programming languages, operating systems, networking, databases, theory of computation, and computing history. That course is useful as the concrete design pressure for this architecture, but creating it is not part of the current implementation scope.

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

The in-memory curriculum catalog should represent the full authored catalog rather than treating the AI/ML course as synonymous with "the curriculum."

## Course identity

Every course needs a stable authored ID, for example:

```text
ai-ml
computer-science
```

At routing, API, persistence, tutor-context, and other durable boundaries, course membership must be explicit whenever a resource is course-bound. Code must not infer the course from the fact that only one currently exists.

Existing module, lesson, and objective ID conventions do not need to be redesigned merely for theoretical purity. The important invariant is that canonical resource identity includes course context wherever ambiguity could otherwise exist.

A future cross-course reference must therefore be qualified explicitly rather than relying on a globally implicit active curriculum. For example, an AI/ML lesson could eventually link to a computer-science lesson about asymptotic complexity. Cross-course links are allowed; a generic cross-course prerequisite graph is not required now.

The global source registry may remain shared across courses so the same authoritative source can be cited from more than one course.

## Git-authored content layout

The current `curriculum/modules/` layout is a single-course shape. The multi-course implementation should move toward a structure similar to:

```text
curriculum/
├── courses/
│   ├── ai-ml/
│   │   ├── course.yaml
│   │   └── modules/
│   │       └── ...
│   └── computer-science/
│       ├── course.yaml
│       └── modules/
│           └── ...
└── sources.yaml
```

The exact `course.yaml` schema should stay small. It needs enough metadata to identify and present the course, but should not grow into a generic LMS course-definition format.

The current AI/ML content should migrate into the explicit `ai-ml` course without changing its pedagogical meaning.

## Application routing

Curriculum-bound application routes should become course-aware. A clear target shape is:

```text
/courses/:courseId
/courses/:courseId/modules/:moduleId
/courses/:courseId/modules/:moduleId/lessons/:lessonId
```

The exact URL naming may be adjusted during implementation if the real router/API constraints justify it, but the invariant is that module and lesson routes carry course identity.

While AI/ML is the only course, the application does **not** need a visible course picker. Existing entry points such as `/` or `/curriculum` may redirect to the default AI/ML course for convenience. That convenience must not recreate a hidden global singleton inside feature code.

When a second course is actually added, course selection/navigation can become visible then.

## API shape

The HTTP API should follow the existing resource-oriented REST conventions and generated Go -> OpenAPI -> Orval -> TanStack Query/Zod contract.

Course-aware resources should be modeled as resources, not as ad hoc action endpoints. A likely shape is:

```text
GET /api/courses
GET /api/courses/{courseId}
GET /api/courses/{courseId}/modules
GET /api/courses/{courseId}/modules/{moduleId}
GET /api/courses/{courseId}/modules/{moduleId}/lessons/{lessonId}
```

This is implementation guidance rather than a requirement to create every endpoint immediately. The implementation should expose only the operations needed by the application while preserving explicit course ownership.

## Learner state

Course identity must be part of persisted state for course-bound learning activity.

This includes, as those features are implemented:

- objective progress and evidence;
- lesson activity;
- spaced-repetition items tied to curriculum objectives;
- worksheet attempts and tutor-reviewed worksheet evidence;
- exercise workspaces and attempts;
- notebook/lab/project status;
- course-specific "continue learning" state;
- tutor page/activity context that refers to curriculum resources.

The default course is a navigation preference, not a substitute for durable course identity.

This is a major reason to establish the multi-course foundation before substantial learner-state persistence and before worksheet implementation: those systems should be built against the correct ownership hierarchy once rather than retrofitted immediately afterward.

## Tutor context

The tutor remains global to the application, but curriculum context must be course-qualified.

A tutor turn associated with a lesson, module, objective, worksheet, exercise, or project should include enough structured identity to resolve the correct course explicitly. The context builder must not rely on a single globally active curriculum.

Cross-course tutoring is allowed naturally: the tutor may recommend material in another course when useful. That does not require loading every course or every learner-state record into every tutor turn.

## Worksheets and workbooks

Worksheets remain optional practice artifacts attached to lessons as defined in [`worksheets.md`](worksheets.md).

Their ownership chain is:

```text
course -> module -> lesson -> worksheet
```

A module workbook aggregates worksheets from that module within its course. No worksheet or workbook implementation should depend on an implicit single-course catalog.

The multi-course foundation should therefore land before worksheet implementation.

## User experience during the foundation refactor

The first multi-course implementation should be intentionally boring from the user's perspective.

After the refactor:

- AI/ML should still be the only authored course;
- it should remain the default course;
- current lessons and modules should preserve their content and ordering;
- the application should look and behave substantially the same;
- no empty course marketplace/catalog UI is required;
- no computer-science course content is required.

The value of the refactor is architectural: platform features added afterward will be built on explicit course ownership.

## Implementation sequence

A focused implementation should proceed roughly in this order:

1. introduce the `Course` domain model and explicit AI/ML course metadata;
2. migrate Git-authored content into a course-aware directory/catalog structure;
3. make curriculum validation and lookup operate across courses;
4. make curriculum API resources and generated clients course-aware;
5. make frontend curriculum routes/navigation and tutor page context course-aware;
6. ensure new persistence schemas use explicit course identity for course-bound state;
7. implement worksheets against that foundation.

This sequence is guidance, not a mandate for one giant PR. Prefer small coherent PRs if the implementation naturally separates.

## Non-goals

Do not use multi-course support as justification to build:

- multi-user accounts or organizations;
- enrollment workflows;
- classroom/teacher features;
- permissions or RBAC around courses;
- a course marketplace;
- a generic course CMS or visual authoring system;
- course search/discovery infrastructure for two local courses;
- calendar semesters, due dates, or cohort pacing;
- a generic cross-course dependency/knowledge graph;
- separate services or databases per course.

Fonzytooter remains a small personal learning application. The architecture is becoming **multi-course**, not **general-purpose LMS**.
