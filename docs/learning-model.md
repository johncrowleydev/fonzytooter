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

Exercises answer questions such as: can the learner calculate, implement, or use the idea in a constrained problem?

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
- repeated test failure;
- project status change;
- note created.

This is not intended to become an event-sourced architecture. It is a practical learner-history table.
