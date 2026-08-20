# Course plans

This directory contains **course-specific curriculum planning documents**.

Platform-wide curriculum architecture belongs in the top-level `docs/` files, especially [`../multi-course.md`](../multi-course.md), [`../learning-model.md`](../learning-model.md), [`../content-authoring.md`](../content-authoring.md), and [`../worksheets.md`](../worksheets.md). A course plan describes what one authored course intends to teach and how that subject is organized pedagogically.

The distinction is intentional:

```text
docs/
├── multi-course.md          platform ownership/routing model
├── learning-model.md        platform learning/evidence model
├── content-authoring.md     authored-content conventions
└── courses/
    ├── README.md            this boundary
    ├── ai-ml.md             AI/ML high-level curriculum plan
    └── ai-ml-syllabus.md    AI/ML canonical core syllabus skeleton

curriculum/
├── courses/
│   └── ai-ml/               runtime-authored course content
└── sources.yaml             shared source registry
```

The planning documents and the authored curriculum have different jobs. `docs/courses/` may describe long-range phases, pedagogical priorities, the canonical syllabus sequence, representative projects, and future modules. `curriculum/courses/<course-id>/` contains the concrete course metadata, modules, lessons, objective definitions, and other authored artifacts the application can load now.

## Current courses

- [`ai-ml.md`](ai-ml.md) — high-level philosophy and long-range plan for the **AI & Machine Learning** course (`ai-ml`), currently the default and only authored course.
- [`ai-ml-syllabus.md`](ai-ml-syllabus.md) — canonical mandatory core sequence for the AI/ML course, including module scope, dependency shape, math/history spines, mastery checkpoints, and optional post-core specializations.

## Adding another course

When a second course is actually ready to be designed:

1. create a course-specific planning document here;
2. keep platform-wide decisions out of that course plan;
3. add runtime-authored content under `curriculum/courses/<course-id>/` when implementation begins;
4. use a stable authored course ID and the existing course-aware catalog/API/routing model;
5. add visible course-selection UX only when multiple real courses make it useful.

Do not create generic LMS, enrollment, marketplace, or course-discovery abstractions merely because more than one authored course exists.
