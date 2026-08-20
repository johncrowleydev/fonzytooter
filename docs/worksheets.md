# Worksheets and printable practice

## Purpose

Fonzytooter may attach one or more optional worksheets to a lesson when repeated pencil-and-paper practice is useful.

Worksheets are especially important for mathematics, where understanding an explanation is not the same as being able to perform the operations reliably. They may also be useful for other concepts that benefit from tracing, calculation, diagramming, or written reasoning.

A worksheet is a **concept-focused practice artifact**, not merely a printable copy of a lesson and not a generic end-of-lesson quiz.

Examples include:

- evaluating functions;
- domain and codomain;
- function composition;
- algebraic manipulation;
- matrix multiplication;
- probability calculations;
- derivative practice;
- tracing a Python function by hand;
- tracing an algorithm;
- calculating a small neural-network forward pass;
- working through a small backpropagation example.

A lesson may have zero worksheets, one worksheet, or several worksheets covering distinct skills.

## Relationship to the curriculum

Worksheets should map their problems to the learning objectives and concepts they practice. The curriculum should be able to distinguish, for example, strong function-evaluation performance from weak function-composition performance rather than treating an entire lesson as one indivisible score.

Worksheets complement the other activity types rather than replacing them:

```text
lesson / explanation
    -> understand the idea

interactive visualization
    -> build intuition

worksheet
    -> practice calculations and written reasoning

embedded coding exercise
    -> apply the idea in code

notebook / project
    -> explore or transfer the idea
```

Printability is an explicit product goal. Some learning work is better done away from the keyboard with room to calculate, draw, annotate, and show intermediate steps.

## Worksheet and workbook output

Each worksheet is available on demand as a printable student PDF and a solutions PDF. Fonzytooter converts the validated structured worksheet to Markdown with authored LaTeX preserved, uses Pandoc to produce LaTeX, and uses Tectonic to typeset PDF bytes. Generated PDFs are temporary responses; they are not stored in Git, SQLite, or a PDF cache.

A module with worksheets also produces a **workbook** and solutions workbook. The workbook is an aggregation of the same worksheet content, not a separately authored curriculum artifact. It is compiled as one structured document with a cover, table of contents, lesson context, page numbering, and page breaks between worksheets. Fonzytooter does not render separate PDFs and concatenate them.

Student documents include response space based on each problem's authored `responseLines`. Solutions documents include authored expected answers. Neither solutions output exposes internal rubric criteria, which remain available for future assessment behavior.

Worksheet and workbook ordering follows module lesson order, worksheet order within each lesson, and stable worksheet ID as the final deterministic tie-breaker.

## Problem design

Problems should identify the concepts/objectives they exercise and, where possible, have an independently specified expected result or rubric.

For deterministic mathematics problems, the AI tutor should not be the sole source of truth for the expected answer. Curriculum data should provide the known answer, expected properties, or grading rubric when practical.

A conceptual problem model might eventually represent information such as:

```text
problem
├── prompt
├── objective/concept IDs
├── expected answer or properties
└── expected reasoning/steps where useful
```

This is a pedagogical requirement, not a commitment to a particular storage schema.

## Showing work

For appropriate problems, assessment should consider both the final answer and the reasoning visible in the submitted work.

A correct final answer with no intermediate reasoning may demonstrate less than a correct answer with a clear derivation. Conversely, an incorrect final answer may still show that most of the concept is understood and reveal one specific mistake.

For example, if a learner writes:

```text
2(x + 3) = 10
2x + 3 = 10
2x = 7
x = 3.5
```

the useful feedback is not merely "incorrect." The earliest error is the distribution step: `2(x + 3)` should become `2x + 6`. The later algebra is internally consistent with that mistaken step.

Worksheet feedback should therefore prefer identifying the earliest meaningful conceptual or computational error when the work provides enough evidence to do so.

## Tutor submission and review

A completed worksheet may be uploaded to the tutor as images or a document for multimodal review.

The tutor should use the authored problem answers/rubrics together with the learner's submitted work to evaluate:

- whether each final answer is correct;
- whether sufficient work is shown where the problem calls for reasoning;
- which intermediate steps are correct;
- the earliest identifiable error in an incorrect solution;
- whether an error appears conceptual, procedural, or arithmetic;
- recurring misconceptions across multiple problems;
- concepts that appear solid;
- concepts that need more practice or prerequisite review.

The tutor should give specific feedback rather than reducing the worksheet to only a numeric grade.

Vision-model interpretation is inherently fallible. When handwriting, problem association, or an intermediate step is ambiguous, the tutor should express that uncertainty rather than fabricating a confident reading.

## Learning evidence

Worksheet review contributes evidence about learning objectives, especially **application** and **conceptual understanding**.

It should not directly award mastery simply because an AI model says the work looks good. The system should preserve the existing distinction between evidence and mastery decisions.

Useful learner-state conclusions may look like:

```text
Function notation              Strong
Evaluating functions           Strong
Domain/codomain distinction    Developing
Function composition           Needs practice
```

These judgments should accumulate from multiple relevant activities over time rather than being treated as permanent labels from one worksheet.

The tutor may recommend actions based on the evidence, such as:

- continue to the next concept;
- attempt another worksheet;
- revisit a prerequisite;
- review a specific lesson section;
- complete a mastery check.

## Scope boundary

This document establishes the curriculum and tutoring role of worksheets. The implemented Git-authoring schema and render-validation workflow are documented in [`content-authoring.md`](content-authoring.md). This document does **not** yet decide:

- how uploads are stored;
- which vision model/provider performs handwriting interpretation;
- the final grading data model;

Those choices should be made when implementation requirements are concrete enough to justify them.
