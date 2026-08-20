# Learning model

## Principles

The curriculum is self-paced. It has sequence and prerequisites, but no calendar schedule.

The system should answer questions like:

- What have I encountered?
- What am I likely to remember?
- What can I apply?
- Where am I repeatedly failing?
- What prerequisites are blocking the next concept?

It should not answer them with fake precision.

## Objectives

An objective is a durable capability with a stable ID.

Example:

```yaml
id: optimization.gradient-descent
name: Explain and implement gradient descent
prerequisites:
  - calculus.derivative.intuition
  - linear-algebra.vectors
```

Objectives form a simple directed prerequisite graph. A graph database is not required.

## Evidence of learning

Different mechanisms establish different kinds of evidence.

### Recall

Spaced repetition answers questions such as: can the learner retrieve or explain the idea after time has passed?

### Application

Exercises and worksheets answer questions such as: can the learner calculate, implement, or use the idea in a constrained problem?

For worksheets, the reasoning shown on the page can provide evidence that a final-answer-only check cannot. A wrong result may still reveal mostly correct reasoning plus one localized mistake; a correct result with no work may provide weaker evidence for a problem intended to assess a procedure or derivation.

Worksheet review should map findings back to the relevant objective/concept IDs and accumulate evidence over time rather than turning one AI judgment into a permanent skill label.

### Transfer

Projects, novel questions, and synthesis tasks answer questions such as: can the learner recognize and apply the idea in a less-scripted context?

Avoid collapsing these into a single magic mastery percentage unless a concrete UX need later justifies it.

An objective might eventually display:

```text
Gradient descent
Introduced    ✓
Recall        Strong
Application   Developing
Transfer      Not assessed
```

AI-interpreted worksheet feedback is evidence, not an independent authority for mastery. The tutor may recommend more practice, prerequisite review, or a mastery check, but it should not directly mark an objective mastered based on a vision-model judgment.

See `docs/worksheets.md` for the printable-practice and worksheet-review model.

## Spaced repetition

The planned scheduler is FSRS rather than a custom algorithm.

Important distinction: FSRS uses elapsed time internally, but that does not make the curriculum calendar-paced.

The user-facing experience should look like:

```text
Reviews ready: 14
Continue learning: Backpropagation
Ready for mastery check: Matrix multiplication
```

not:

```text
Week 8
You are 4 days behind
```

## Flashcards/review items

Core review items may be authored with the curriculum. Additional items may be proposed during study.

AI-generated cards should be candidates requiring approval/editing rather than silently flooding the review queue.

Review items should map to objective IDs so review performance has semantic meaning beyond "card 123 was hard."

## Activities

Keep a small activity history to provide useful recency context:

- lesson opened/completed;
- video completed;
- review attempted;
- exercise run/checked;
- worksheet attempted/reviewed;
- repeated test failure;
- project status change;
- note created.

This is not intended to become an event-sourced architecture. It is a practical learner-history table.

The implemented first learner-state slice records explicit lesson completion
and one `lesson_completed` activity per incomplete-to-complete transition.
Completing a lesson marks its referenced objectives as introduced in the
derived course progress view. Recall, application, and transfer remain
explicitly not assessed until their evidence-producing workflows exist.
