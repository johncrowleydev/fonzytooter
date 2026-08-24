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

The review scheduler is FSRS-6 through the pinned `go-fsrs/v4` library with its
default parameters. The server owns all scheduling calculations: it supplies
the real Again, Hard, Good, and Easy previews and applies the selected rating at
the same injected time. The browser only formats those results for display.

Authored items without learner state are treated as virtual New cards. A safe
queue read does not populate SQLite; the first submitted rating creates the
card row, an immutable before/after review log, and a `review_completed`
activity in one transaction. This recall evidence does not become a mastery
percentage or an AI-authored judgment.

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

Core review content is implemented as optional Git-authored YAML under each module's `reviews/` directory. Each item identifies its source lesson and one or more curriculum objectives, and keeps prompt, answer, and an optional hint separate from learner state. The curriculum API exposes this authored content for the later review workflow; it does not attach FSRS scheduling fields or review history to the authored resource.

## Video learning

Curated YouTube videos are another way for a learner to encounter or revisit an objective. They are first-class curriculum material, but watching them is not itself evidence of recall, application, or transfer.

A completed video may establish that the learner encountered an additional explanation and may contribute useful recency context. That state can support progress-aware recommendations, such as suggesting an unwatched visual explanation for the current objective or revisiting a video after repeated difficulty.

Do not convert video completion into a mastery score or use playback telemetry as a proxy for comprehension. A learner can watch every second of a video without understanding it; conversely, a learner may understand the objective without watching the curated video at all.

Video state belongs to authenticated learner history. Public users may view embedded videos and module playlists without creating learner-state rows. See `docs/youtube-learning.md` for the full curation, embedding, playlist, recommendation, and progress model.

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

Embedded exercise checks now persist a code snapshot, server-derived aggregate
counts, normalized authored test results, and one `exercise_checked` activity.
Exploratory runs remain intentionally ephemeral. This is application evidence;
it does not manufacture a mastery percentage or silently change objective
status.
